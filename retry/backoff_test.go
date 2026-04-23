package retry

import (
	"testing"
	"time"
)

func TestExponentialBackoff_NextInterval(t *testing.T) {
	b := &ExponentialBackoff{
		Multiplier:  2.0,
		MaxInterval: 60 * time.Second,
		Jitter:      0,
	}
	base := time.Second

	tests := []struct {
		attempt  int
		expected time.Duration
	}{
		{0, 1 * time.Second},
		{1, 2 * time.Second},
		{2, 4 * time.Second},
		{3, 8 * time.Second},
	}

	for _, tt := range tests {
		got := b.NextInterval(tt.attempt, base)
		if got != tt.expected {
			t.Errorf("attempt %d: expected %v, got %v", tt.attempt, tt.expected, got)
		}
	}
}

func TestExponentialBackoff_MaxInterval(t *testing.T) {
	b := &ExponentialBackoff{
		Multiplier:  2.0,
		MaxInterval: 5 * time.Second,
		Jitter:      0,
	}
	result := b.NextInterval(10, time.Second)
	if result != 5*time.Second {
		t.Errorf("expected max interval 5s, got %v", result)
	}
}

func TestExponentialBackoff_Jitter(t *testing.T) {
	b := &ExponentialBackoff{
		Multiplier:  2.0,
		MaxInterval: 60 * time.Second,
		Jitter:      0.5,
	}
	base := time.Second
	result := b.NextInterval(1, base)
	// With jitter 0.5 on base*2=2s, result should be between 1s and 3s
	if result < time.Second || result > 3*time.Second {
		t.Errorf("jitter result %v out of expected range [1s, 3s]", result)
	}
}

func TestConstantBackoff_NextInterval(t *testing.T) {
	c := &ConstantBackoff{}
	base := 500 * time.Millisecond
	for i := 0; i < 5; i++ {
		if got := c.NextInterval(i, base); got != base {
			t.Errorf("attempt %d: expected %v, got %v", i, base, got)
		}
	}
}

func TestLinearBackoff_NextInterval(t *testing.T) {
	l := &LinearBackoff{MaxInterval: 10 * time.Second}
	base := time.Second

	if got := l.NextInterval(0, base); got != time.Second {
		t.Errorf("attempt 0: expected 1s, got %v", got)
	}
	if got := l.NextInterval(1, base); got != 2*time.Second {
		t.Errorf("attempt 1: expected 2s, got %v", got)
	}
	if got := l.NextInterval(20, base); got != 10*time.Second {
		t.Errorf("attempt 20: expected max 10s, got %v", got)
	}
}

func TestDefaultExponentialBackoff(t *testing.T) {
	b := DefaultExponentialBackoff()
	if b.Multiplier != 2.0 {
		t.Errorf("expected multiplier 2.0, got %v", b.Multiplier)
	}
	if b.MaxInterval != 30*time.Second {
		t.Errorf("expected max interval 30s, got %v", b.MaxInterval)
	}
}
