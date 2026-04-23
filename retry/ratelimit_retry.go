package retry

import (
	"context"
	"fmt"
)

// RateLimitedRetrier combines a RateLimiter with retry logic so that
// operations are both rate-limited and retried on transient failures.
type RateLimitedRetrier struct {
	limiter *RateLimiter
	config  Config
}

// RateLimitedRetrierConfig holds configuration for a RateLimitedRetrier.
type RateLimitedRetrierConfig struct {
	Retry       Config
	RateLimiter RateLimiterConfig
}

// DefaultRateLimitedRetrierConfig is the default configuration.
var DefaultRateLimitedRetrierConfig = RateLimitedRetrierConfig{
	Retry:       DefaultConfig,
	RateLimiter: DefaultRateLimiterConfig,
}

// NewRateLimitedRetrier creates a new RateLimitedRetrier.
func NewRateLimitedRetrier(cfg RateLimitedRetrierConfig) *RateLimitedRetrier {
	return &RateLimitedRetrier{
		limiter: NewRateLimiter(cfg.RateLimiter),
		config:  cfg.Retry,
	}
}

// Do executes fn with rate limiting and retry logic.
// It waits for a rate-limit token before each attempt.
func (r *RateLimitedRetrier) Do(ctx context.Context, fn func() error) error {
	var attempt int
	for {
		if err := r.limiter.Wait(ctx); err != nil {
			return fmt.Errorf("rate limiter wait: %w", err)
		}

		err := fn()
		if err == nil {
			return nil
		}

		if IsPermanent(err) {
			return err
		}

		attempt++
		if attempt >= r.config.MaxAttempts {
			return fmt.Errorf("max attempts reached: %w", err)
		}

		interval := nextInterval(r.config, attempt)
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-sleepTimer(interval):
		}
	}
}
