package retry

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestCircuitRetrier_SuccessfulCall(t *testing.T) {
	cfg := DefaultCircuitRetryConfig()
	cr := NewCircuitRetrier(cfg)

	err := cr.Do(context.Background(), func() error {
		return nil
	})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if cr.State() != CircuitClosed {
		t.Errorf("expected circuit closed after success")
	}
}

func TestCircuitRetrier_OpensCircuitAfterFailures(t *testing.T) {
	cfg := DefaultCircuitRetryConfig()
	cfg.CircuitBreaker.MaxFailures = 2
	cfg.Retry.MaxAttempts = 1
	cr := NewCircuitRetrier(cfg)

	sentinel := errors.New("service error")
	for i := 0; i < 3; i++ {
		_ = cr.Do(context.Background(), func() error {
			return sentinel
		})
	}

	if cr.State() != CircuitOpen {
		t.Errorf("expected circuit open after repeated failures, got %v", cr.State())
	}
}

func TestCircuitRetrier_BlocksWhenOpen(t *testing.T) {
	cfg := DefaultCircuitRetryConfig()
	cfg.CircuitBreaker.MaxFailures = 1
	cfg.Retry.MaxAttempts = 1
	cr := NewCircuitRetrier(cfg)

	_ = cr.Do(context.Background(), func() error {
		return errors.New("fail")
	})

	err := cr.Do(context.Background(), func() error {
		return nil
	})
	if !errors.Is(err, ErrCircuitOpen) {
		t.Errorf("expected ErrCircuitOpen, got %v", err)
	}
}

func TestCircuitRetrier_ResetAllowsRequests(t *testing.T) {
	cfg := DefaultCircuitRetryConfig()
	cfg.CircuitBreaker.MaxFailures = 1
	cfg.Retry.MaxAttempts = 1
	cr := NewCircuitRetrier(cfg)

	_ = cr.Do(context.Background(), func() error {
		return errors.New("fail")
	})

	cr.Reset()

	err := cr.Do(context.Background(), func() error {
		return nil
	})
	if err != nil {
		t.Errorf("expected nil after reset, got %v", err)
	}
}

func TestCircuitRetrier_ContextCancellation(t *testing.T) {
	cfg := DefaultCircuitRetryConfig()
	cfg.Retry.MaxAttempts = 10
	cfg.Retry.Backoff = ConstantBackoff{Interval: 50 * time.Millisecond}
	cr := NewCircuitRetrier(cfg)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Millisecond)
	defer cancel()

	err := cr.Do(ctx, func() error {
		return errors.New("transient")
	})
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("expected DeadlineExceeded, got %v", err)
	}
}
