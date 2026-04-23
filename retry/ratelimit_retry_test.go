package retry

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestRateLimitedRetrier_SuccessOnFirstAttempt(t *testing.T) {
	r := NewRateLimitedRetrier(DefaultRateLimitedRetrierConfig)
	calls := 0
	err := r.Do(context.Background(), func() error {
		calls++
		return nil
	})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if calls != 1 {
		t.Fatalf("expected 1 call, got %d", calls)
	}
}

func TestRateLimitedRetrier_RetriesTransientErrors(t *testing.T) {
	cfg := RateLimitedRetrierConfig{
		Retry: Config{
			MaxAttempts: 3,
			Backoff:     &ConstantBackoff{Interval: time.Millisecond},
		},
		RateLimiter: RateLimiterConfig{Rate: 1000, Burst: 100},
	}
	r := NewRateLimitedRetrier(cfg)
	calls := 0
	transient := errors.New("transient error")
	err := r.Do(context.Background(), func() error {
		calls++
		if calls < 3 {
			return transient
		}
		return nil
	})
	if err != nil {
		t.Fatalf("expected success after retries, got: %v", err)
	}
	if calls != 3 {
		t.Fatalf("expected 3 calls, got %d", calls)
	}
}

func TestRateLimitedRetrier_PermanentErrorStopsRetry(t *testing.T) {
	cfg := RateLimitedRetrierConfig{
		Retry: Config{
			MaxAttempts: 5,
			Backoff:     &ConstantBackoff{Interval: time.Millisecond},
		},
		RateLimiter: RateLimiterConfig{Rate: 1000, Burst: 100},
	}
	r := NewRateLimitedRetrier(cfg)
	calls := 0
	err := r.Do(context.Background(), func() error {
		calls++
		return Permanent(errors.New("fatal"))
	})
	if err == nil {
		t.Fatal("expected error for permanent failure")
	}
	if calls != 1 {
		t.Fatalf("expected 1 call for permanent error, got %d", calls)
	}
}

func TestRateLimitedRetrier_ContextCancellation(t *testing.T) {
	cfg := RateLimitedRetrierConfig{
		Retry:       DefaultConfig,
		RateLimiter: RateLimiterConfig{Rate: 0.001, Burst: 1},
	}
	r := NewRateLimitedRetrier(cfg)
	// drain the token so Wait must block
	r.limiter.Allow()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err := r.Do(ctx, func() error { return nil })
	if err == nil {
		t.Fatal("expected error due to cancelled context")
	}
}
