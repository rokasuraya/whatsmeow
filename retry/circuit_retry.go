package retry

import (
	"context"
	"fmt"
)

// CircuitRetryConfig combines retry and circuit breaker configuration.
type CircuitRetryConfig struct {
	Retry          Config
	CircuitBreaker CircuitBreakerConfig
}

// DefaultCircuitRetryConfig returns a default combined config.
func DefaultCircuitRetryConfig() CircuitRetryConfig {
	return CircuitRetryConfig{
		Retry:          DefaultConfig,
		CircuitBreaker: DefaultCircuitBreakerConfig(),
	}
}

// CircuitRetrier wraps a CircuitBreaker with retry logic.
type CircuitRetrier struct {
	cb  *CircuitBreaker
	cfg CircuitRetryConfig
}

// NewCircuitRetrier creates a CircuitRetrier with the given config.
func NewCircuitRetrier(cfg CircuitRetryConfig) *CircuitRetrier {
	return &CircuitRetrier{
		cb:  NewCircuitBreaker(cfg.CircuitBreaker),
		cfg: cfg,
	}
}

// Do executes fn with retry logic guarded by the circuit breaker.
// If the circuit is open, it returns ErrCircuitOpen immediately.
func (cr *CircuitRetrier) Do(ctx context.Context, fn func() error) error {
	if err := cr.cb.Allow(); err != nil {
		return fmt.Errorf("circuit breaker prevented execution: %w", err)
	}

	err := Do(ctx, cr.cfg.Retry, func() error {
		if allowErr := cr.cb.Allow(); allowErr != nil {
			return Permanent(allowErr)
		}
		callErr := fn()
		if callErr != nil {
			cr.cb.RecordFailure()
			return callErr
		}
		cr.cb.RecordSuccess()
		return nil
	})
	return err
}

// State returns the current circuit breaker state.
func (cr *CircuitRetrier) State() CircuitState {
	return cr.cb.State()
}

// Reset resets the circuit breaker to closed state.
func (cr *CircuitRetrier) Reset() {
	cr.cb.RecordSuccess()
}
