package retry

import (
	"context"
	"sync"
	"time"
)

// HedgedRetryConfig configures the hedged retry strategy.
// A hedged request sends duplicate requests after a delay if the original
// has not yet completed, returning the first successful response.
type HedgedRetryConfig struct {
	// MaxHedges is the maximum number of additional (hedged) requests to send.
	MaxHedges int
	// HedgeDelay is the duration to wait before sending each hedged request.
	HedgeDelay time.Duration
	// IsRetryable determines whether an error should trigger a hedge.
	// If nil, all non-nil errors are considered retryable.
	IsRetryable func(err error) bool
}

// DefaultHedgedRetryConfig returns a HedgedRetryConfig with sensible defaults.
// Note: increased HedgeDelay to 200ms to reduce unnecessary duplicate requests
// on connections that are just slightly slow rather than actually failing.
func DefaultHedgedRetryConfig() HedgedRetryConfig {
	return HedgedRetryConfig{
		MaxHedges:  2,
		HedgeDelay: 200 * time.Millisecond,
		IsRetryable: func(err error) bool {
			return err != nil && !IsPermanent(err)
		},
	}
}

type hedgeResult struct {
	value interface{}
	err   error
}

// DoHedged executes fn using a hedged retry strategy. It sends up to
// cfg.MaxHedges additional requests after cfg.HedgeDelay each, and returns
// the first successful result. If all attempts fail, the last error is returned.
//
// Example:
//
//	result, err := retry.DoHedged(ctx, retry.DefaultHedgedRetryConfig(), func(ctx context.Context) (interface{}, error) {
//		return fetchData(ctx)
//	})
func DoHedged(ctx context.Context, cfg HedgedRetryConfig, fn func(ctx context.Context) (interface{}, error)) (interface{}, error) {
	if cfg.MaxHedges < 0 {
		cfg.MaxHedges = 0
	}
	if cfg.IsRetryable == nil {
		cfg.IsRetryable = func(err error) bool { return err != nil }
	}

	total := cfg.MaxHedges + 1
	resultCh := make(chan hedgeResult, total)

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	var wg sync.WaitGroup
	var lastErr error
	var mu sync.Mutex

	launch := func() {
		wg.Add(1)
		go func() {
			defer wg.Done()
			v, err := fn(ctx)
			resultCh <- hedgeResult{value: v, err: err}
		}()
	}

	// Launch the first request immediately.
	launch()

	successCh := make(chan hedgeResult, 1)

	go func() {
		sent := 1
		ticker := time.NewTicker(cfg.HedgeDelay)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case res := <-resultCh:
				if res.err == nil {
					successCh <- res
					return
				}
				mu.Lock()
				lastErr = res.err
				mu.Unlock()
				if !cfg.IsRetryable(res.err) || sent >= total {
					// All hedges exhausted or permanent error; wait for remaining.
					for i := sent; i < total; i++ {
						r := <-resultCh
						if r.err == nil {
							successCh <- r
							return
						}
						mu.Lock()
						lastErr = r.err
						mu.Unlock()
					}
					successCh <- hedgeResult{err: lastErr}
					return
				}
			case <-ticker.C:
				if sent < total {
					launch()
					sent++
				}
			}
		}
	}()

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case res := <-successCh:
		cancel()
		wg.Wait()
		return res.value, res.err
	}
}
