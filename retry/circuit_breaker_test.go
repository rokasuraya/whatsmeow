package retry

import (
	"testing"
	"time"
)

func TestCircuitBreaker_InitiallyClosed(t *testing.T) {
	cb := NewCircuitBreaker(DefaultCircuitBreakerConfig())
	if cb.State() != CircuitClosed {
		t.Fatalf("expected CircuitClosed, got %v", cb.State())
	}
	if err := cb.Allow(); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
}

func TestCircuitBreaker_OpensAfterMaxFailures(t *testing.T) {
	cfg := DefaultCircuitBreakerConfig()
	cfg.MaxFailures = 3
	cb := NewCircuitBreaker(cfg)

	for i := 0; i < 3; i++ {
		cb.RecordFailure()
	}

	if cb.State() != CircuitOpen {
		t.Fatalf("expected CircuitOpen after max failures, got %v", cb.State())
	}
	if err := cb.Allow(); err != ErrCircuitOpen {
		t.Fatalf("expected ErrCircuitOpen, got %v", err)
	}
}

func TestCircuitBreaker_ClosesOnSuccess(t *testing.T) {
	cfg := DefaultCircuitBreakerConfig()
	cfg.MaxFailures = 2
	cb := NewCircuitBreaker(cfg)

	cb.RecordFailure()
	cb.RecordFailure()
	if cb.State() != CircuitOpen {
		t.Fatal("expected open circuit")
	}

	cb.RecordSuccess()
	if cb.State() != CircuitClosed {
		t.Fatalf("expected CircuitClosed after success, got %v", cb.State())
	}
}

func TestCircuitBreaker_HalfOpenAfterTimeout(t *testing.T) {
	cfg := DefaultCircuitBreakerConfig()
	cfg.MaxFailures = 1
	cfg.Timeout = 10 * time.Millisecond
	cb := NewCircuitBreaker(cfg)

	cb.RecordFailure()
	if cb.State() != CircuitOpen {
		t.Fatal("expected open circuit")
	}

	time.Sleep(20 * time.Millisecond)
	if cb.State() != CircuitHalfOpen {
		t.Fatalf("expected CircuitHalfOpen after timeout, got %v", cb.State())
	}
	if err := cb.Allow(); err != nil {
		t.Fatalf("expected nil error in half-open, got %v", err)
	}
}

func TestCircuitBreaker_OnStateChangeCallback(t *testing.T) {
	var transitions []CircuitState
	cfg := DefaultCircuitBreakerConfig()
	cfg.MaxFailures = 1
	cfg.OnStateChange = func(_, to CircuitState) {
		transitions = append(transitions, to)
	}
	cb := NewCircuitBreaker(cfg)

	cb.RecordFailure()
	cb.RecordSuccess()

	if len(transitions) != 2 {
		t.Fatalf("expected 2 transitions, got %d", len(transitions))
	}
	if transitions[0] != CircuitOpen {
		t.Errorf("expected first transition to CircuitOpen")
	}
	if transitions[1] != CircuitClosed {
		t.Errorf("expected second transition to CircuitClosed")
	}
}
