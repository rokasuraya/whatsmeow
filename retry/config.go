package retry

import "time"

// Config holds the configuration for retry behavior.
type Config struct {
	// MaxAttempts is the maximum number of attempts (including the first call).
	// A value of 0 means unlimited retries.
	MaxAttempts int
	// Interval is the base wait duration between retries.
	Interval time.Duration
	// Backoff determines how intervals grow between retries.
	// Defaults to ConstantBackoff if nil.
	Backoff BackoffStrategy
}

// WithMaxAttempts returns a copy of the Config with MaxAttempts set.
func (c Config) WithMaxAttempts(n int) Config {
	c.MaxAttempts = n
	return c
}

// WithInterval returns a copy of the Config with Interval set.
func (c Config) WithInterval(d time.Duration) Config {
	c.Interval = d
	return c
}

// WithBackoff returns a copy of the Config with the given BackoffStrategy.
func (c Config) WithBackoff(b BackoffStrategy) Config {
	c.Backoff = b
	return c
}

// WithExponentialBackoff is a convenience method to set exponential backoff
// using the default ExponentialBackoff configuration.
func (c Config) WithExponentialBackoff() Config {
	c.Backoff = DefaultExponentialBackoff()
	return c
}

// effectiveBackoff returns the configured backoff or a ConstantBackoff default.
func (c Config) effectiveBackoff() BackoffStrategy {
	if c.Backoff != nil {
		return c.Backoff
	}
	return &ConstantBackoff{}
}

// nextInterval computes the wait duration for the given attempt index.
func (c Config) nextInterval(attempt int) time.Duration {
	return c.effectiveBackoff().NextInterval(attempt, c.Interval)
}
