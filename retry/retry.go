// Package retry provides utilities for retrying failed operations with
// configurable backoff strategies.
package retry

import (
	"context"
	"math"
	"time"
)

// Config holds the configuration for retry behavior.
type Config struct {
	// MaxAttempts is the maximum number of attempts (0 means unlimited).
	MaxAttempts int
	// InitialInterval is the initial wait interval between retries.
	InitialInterval time.Duration
	// MaxInterval is the maximum wait interval between retries.
	MaxInterval time.Duration
	// Multiplier is the factor by which the interval increases each retry.
	Multiplier float64
	// Jitter adds randomness to the interval to avoid thundering herd.
	Jitter bool
}

// DefaultConfig returns a Config with sensible defaults.
func DefaultConfig() Config {
	return Config{
		MaxAttempts:     5,
		InitialInterval: 1 * time.Second,
		MaxInterval:     30 * time.Second,
		Multiplier:      2.0,
		Jitter:          true,
	}
}

// Do executes fn with retry logic according to cfg.
// It stops retrying if ctx is cancelled or MaxAttempts is reached.
func Do(ctx context.Context, cfg Config, fn func(attempt int) error) error {
	var lastErr error
	for attempt := 1; ; attempt++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		lastErr = fn(attempt)
		if lastErr == nil {
			return nil
		}
		if cfg.MaxAttempts > 0 && attempt >= cfg.MaxAttempts {
			return lastErr
		}
		wait := nextInterval(cfg, attempt)
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(wait):
		}
	}
}

// nextInterval calculates the wait duration for a given attempt.
func nextInterval(cfg Config, attempt int) time.Duration {
	interval := float64(cfg.InitialInterval) * math.Pow(cfg.Multiplier, float64(attempt-1))
	max := float64(cfg.MaxInterval)
	if interval > max {
		interval = max
	}
	return time.Duration(interval)
}
