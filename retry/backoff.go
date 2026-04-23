package retry

import (
	"math"
	"math/rand"
	"time"
)

// BackoffStrategy defines how wait intervals are calculated between retries.
type BackoffStrategy interface {
	NextInterval(attempt int, baseInterval time.Duration) time.Duration
}

// ExponentialBackoff increases the wait interval exponentially with each attempt.
type ExponentialBackoff struct {
	// Multiplier is the base of the exponential growth (default: 2.0).
	Multiplier float64
	// MaxInterval caps the maximum wait duration.
	MaxInterval time.Duration
	// Jitter adds randomness to prevent thundering herd (0.0 to 1.0).
	Jitter float64
}

// DefaultExponentialBackoff returns an ExponentialBackoff with sensible defaults.
func DefaultExponentialBackoff() *ExponentialBackoff {
	return &ExponentialBackoff{
		Multiplier:  2.0,
		MaxInterval: 30 * time.Second,
		Jitter:      0.1,
	}
}

// NextInterval calculates the next wait interval for the given attempt number.
func (e *ExponentialBackoff) NextInterval(attempt int, baseInterval time.Duration) time.Duration {
	multiplier := e.Multiplier
	if multiplier <= 0 {
		multiplier = 2.0
	}

	interval := float64(baseInterval) * math.Pow(multiplier, float64(attempt))

	if e.Jitter > 0 {
		jitter := e.Jitter * interval * (rand.Float64()*2 - 1) //nolint:gosec
		interval += jitter
	}

	result := time.Duration(interval)
	if e.MaxInterval > 0 && result > e.MaxInterval {
		return e.MaxInterval
	}
	return result
}

// ConstantBackoff returns the same interval for every attempt.
type ConstantBackoff struct{}

// NextInterval returns the base interval unchanged.
func (c *ConstantBackoff) NextInterval(_ int, baseInterval time.Duration) time.Duration {
	return baseInterval
}

// LinearBackoff increases the wait interval linearly with each attempt.
type LinearBackoff struct {
	// MaxInterval caps the maximum wait duration.
	MaxInterval time.Duration
}

// NextInterval calculates the next wait interval linearly.
func (l *LinearBackoff) NextInterval(attempt int, baseInterval time.Duration) time.Duration {
	result := baseInterval * time.Duration(attempt+1)
	if l.MaxInterval > 0 && result > l.MaxInterval {
		return l.MaxInterval
	}
	return result
}
