package retry_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"go.mau.fi/whatsmeow/retry"
)

var errFake = errors.New("fake error")

func TestDo_SuccessOnFirstAttempt(t *testing.T) {
	cfg := retry.DefaultConfig()
	cfg.InitialInterval = time.Millisecond
	calls := 0
	err := retry.Do(context.Background(), cfg, func(attempt int) error {
		calls++
		return nil
	})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if calls != 1 {
		t.Fatalf("expected 1 call, got %d", calls)
	}
}

func TestDo_RetriesUntilMaxAttempts(t *testing.T) {
	cfg := retry.Config{
		MaxAttempts:     3,
		InitialInterval: time.Millisecond,
		MaxInterval:     5 * time.Millisecond,
		Multiplier:      1.5,
	}
	calls := 0
	err := retry.Do(context.Background(), cfg, func(attempt int) error {
		calls++
		return errFake
	})
	if !errors.Is(err, errFake) {
		t.Fatalf("expected errFake, got %v", err)
	}
	if calls != 3 {
		t.Fatalf("expected 3 calls, got %d", calls)
	}
}

func TestDo_SucceedsAfterRetries(t *testing.T) {
	cfg := retry.Config{
		MaxAttempts:     5,
		InitialInterval: time.Millisecond,
		MaxInterval:     10 * time.Millisecond,
		Multiplier:      2.0,
	}
	calls := 0
	err := retry.Do(context.Background(), cfg, func(attempt int) error {
		calls++
		if attempt < 3 {
			return errFake
		}
		return nil
	})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if calls != 3 {
		t.Fatalf("expected 3 calls, got %d", calls)
	}
}

func TestDo_ContextCancellation(t *testing.T) {
	cfg := retry.Config{
		MaxAttempts:     10,
		InitialInterval: 50 * time.Millisecond,
		MaxInterval:     500 * time.Millisecond,
		Multiplier:      2.0,
	}
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Millisecond)
	defer cancel()
	err := retry.Do(ctx, cfg, func(attempt int) error {
		return errFake
	})
	if err == nil {
		t.Fatal("expected context error, got nil")
	}
}
