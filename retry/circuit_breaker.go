package retry

import (
	"errors"
	"sync"
	"time"
)

// CircuitState represents the state of a circuit breaker.
type CircuitState int

const (
	// CircuitClosed allows requests through.
	CircuitClosed CircuitState = iota
	// CircuitOpen blocks requests.
	CircuitOpen
	// CircuitHalfOpen allows a single probe request.
	CircuitHalfOpen
)

// ErrCircuitOpen is returned when the circuit breaker is open.
var ErrCircuitOpen = errors.New("circuit breaker is open")

// CircuitBreakerConfig holds configuration for a circuit breaker.
type CircuitBreakerConfig struct {
	// MaxFailures is the number of consecutive failures before opening.
	MaxFailures int
	// Timeout is how long to wait before transitioning to half-open.
	Timeout time.Duration
	// OnStateChange is called when the circuit state changes.
	OnStateChange func(from, to CircuitState)
}

// DefaultCircuitBreakerConfig returns a sensible default config.
func DefaultCircuitBreakerConfig() CircuitBreakerConfig {
	return CircuitBreakerConfig{
		MaxFailures: 5,
		Timeout:     30 * time.Second,
	}
}

// CircuitBreaker implements the circuit breaker pattern.
type CircuitBreaker struct {
	mu          sync.Mutex
	cfg         CircuitBreakerConfig
	state       CircuitState
	failures    int
	lastFailure time.Time
}

// NewCircuitBreaker creates a new CircuitBreaker with the given config.
func NewCircuitBreaker(cfg CircuitBreakerConfig) *CircuitBreaker {
	return &CircuitBreaker{cfg: cfg, state: CircuitClosed}
}

// State returns the current circuit state.
func (cb *CircuitBreaker) State() CircuitState {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.evalState()
	return cb.state
}

// Allow returns nil if a request is permitted, or ErrCircuitOpen.
func (cb *CircuitBreaker) Allow() error {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.evalState()
	if cb.state == CircuitOpen {
		return ErrCircuitOpen
	}
	return nil
}

// RecordSuccess records a successful call, potentially closing the circuit.
func (cb *CircuitBreaker) RecordSuccess() {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	prev := cb.state
	cb.failures = 0
	cb.state = CircuitClosed
	if prev != CircuitClosed && cb.cfg.OnStateChange != nil {
		cb.cfg.OnStateChange(prev, CircuitClosed)
	}
}

// RecordFailure records a failed call, potentially opening the circuit.
func (cb *CircuitBreaker) RecordFailure() {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.failures++
	cb.lastFailure = time.Now()
	if cb.failures >= cb.cfg.MaxFailures && cb.state != CircuitOpen {
		prev := cb.state
		cb.state = CircuitOpen
		if cb.cfg.OnStateChange != nil {
			cb.cfg.OnStateChange(prev, CircuitOpen)
		}
	}
}

// evalState checks if the circuit should transition from open to half-open.
func (cb *CircuitBreaker) evalState() {
	if cb.state == CircuitOpen && time.Since(cb.lastFailure) >= cb.cfg.Timeout {
		prev := cb.state
		cb.state = CircuitHalfOpen
		if cb.cfg.OnStateChange != nil {
			cb.cfg.OnStateChange(prev, CircuitHalfOpen)
		}
	}
}
