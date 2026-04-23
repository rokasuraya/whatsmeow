// Package retry provides retry utilities with optional circuit breaker support.
//
// # Circuit Breaker
//
// The CircuitBreaker type implements the circuit breaker pattern to prevent
// cascading failures when a downstream service is unavailable.
//
// States:
//
//   - CircuitClosed: normal operation, requests are allowed through.
//   - CircuitOpen: the circuit has tripped; requests are rejected immediately
//     with ErrCircuitOpen to give the downstream service time to recover.
//   - CircuitHalfOpen: after the configured Timeout has elapsed, one probe
//     request is allowed through. A success closes the circuit; a failure
//     re-opens it.
//
// # Combined Usage
//
// CircuitRetrier combines retry logic with a circuit breaker:
//
//	cr := retry.NewCircuitRetrier(retry.DefaultCircuitRetryConfig())
//	err := cr.Do(ctx, func() error {
//		return callExternalService()
//	})
//
// The circuit breaker state is shared across all calls made through the same
// CircuitRetrier instance, making it suitable for wrapping a single downstream
// dependency.
package retry
