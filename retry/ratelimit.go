package retry

import (
	"context"
	"sync"
	"time"
)

// RateLimiter limits the rate of operations using a token bucket algorithm.
type RateLimiter struct {
	mu       sync.Mutex
	tokens   float64
	max      float64
	rate     float64 // tokens per second
	lastTime time.Time
}

// RateLimiterConfig holds configuration for a RateLimiter.
type RateLimiterConfig struct {
	// Rate is the number of tokens added per second.
	Rate float64
	// Burst is the maximum number of tokens that can accumulate.
	Burst float64
}

// DefaultRateLimiterConfig is the default configuration for a RateLimiter.
var DefaultRateLimiterConfig = RateLimiterConfig{
	Rate:  10,
	Burst: 20,
}

// NewRateLimiter creates a new RateLimiter with the given configuration.
func NewRateLimiter(cfg RateLimiterConfig) *RateLimiter {
	return &RateLimiter{
		tokens:   cfg.Burst,
		max:      cfg.Burst,
		rate:     cfg.Rate,
		lastTime: time.Now(),
	}
}

// refill adds tokens based on elapsed time since the last refill.
func (r *RateLimiter) refill() {
	now := time.Now()
	elapsed := now.Sub(r.lastTime).Seconds()
	r.tokens = min(r.max, r.tokens+elapsed*r.rate)
	r.lastTime = now
}

// Allow reports whether one token is available and consumes it.
func (r *RateLimiter) Allow() bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.refill()
	if r.tokens >= 1 {
		r.tokens--
		return true
	}
	return false
}

// Wait blocks until a token is available or the context is cancelled.
func (r *RateLimiter) Wait(ctx context.Context) error {
	for {
		if r.Allow() {
			return nil
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(time.Duration(float64(time.Second) / r.rate)):
		}
	}
}

func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}
