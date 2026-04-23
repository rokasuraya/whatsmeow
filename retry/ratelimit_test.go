package retry

import (
	"context"
	"testing"
	"time"
)

func TestRateLimiter_AllowConsumesToken(t *testing.T) {
	rl := NewRateLimiter(RateLimiterConfig{Rate: 1, Burst: 2})
	if !rl.Allow() {
		t.Fatal("expected first Allow() to succeed")
	}
	if !rl.Allow() {
		t.Fatal("expected second Allow() to succeed (burst=2)")
	}
	if rl.Allow() {
		t.Fatal("expected third Allow() to fail (tokens exhausted)")
	}
}

func TestRateLimiter_RefillsOverTime(t *testing.T) {
	rl := NewRateLimiter(RateLimiterConfig{Rate: 100, Burst: 1})
	if !rl.Allow() {
		t.Fatal("expected first Allow() to succeed")
	}
	if rl.Allow() {
		t.Fatal("expected Allow() to fail immediately after exhaustion")
	}
	time.Sleep(20 * time.Millisecond)
	if !rl.Allow() {
		t.Fatal("expected Allow() to succeed after refill period")
	}
}

func TestRateLimiter_WaitSucceeds(t *testing.T) {
	rl := NewRateLimiter(RateLimiterConfig{Rate: 100, Burst: 1})
	// drain the token
	rl.Allow()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := rl.Wait(ctx); err != nil {
		t.Fatalf("expected Wait to succeed, got: %v", err)
	}
}

func TestRateLimiter_WaitCancelledContext(t *testing.T) {
	rl := NewRateLimiter(RateLimiterConfig{Rate: 0.01, Burst: 1})
	// drain the token so Wait must block
	rl.Allow()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := rl.Wait(ctx); err == nil {
		t.Fatal("expected Wait to return error for cancelled context")
	}
}

func TestRateLimiter_DefaultConfig(t *testing.T) {
	rl := NewRateLimiter(DefaultRateLimiterConfig)
	count := 0
	for rl.Allow() {
		count++
	}
	if count != int(DefaultRateLimiterConfig.Burst) {
		t.Fatalf("expected %d tokens, got %d", int(DefaultRateLimiterConfig.Burst), count)
	}
}
