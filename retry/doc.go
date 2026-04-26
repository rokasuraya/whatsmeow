// Package retry provides a simple, context-aware retry mechanism with
// exponential backoff for use throughout the whatsmeow library.
//
// Basic usage:
//
//	cfg := retry.DefaultConfig()
//	err := retry.Do(ctx, cfg, func(attempt int) error {
//		return doSomething()
//	})
//
// To prevent retrying on unrecoverable errors, wrap them with retry.Permanent:
//
//	err := retry.Do(ctx, cfg, func(attempt int) error {
//		result, err := doSomething()
//		if errors.Is(err, ErrUnrecoverable) {
//			return retry.Permanent(err)
//		}
//		return err
//	})
//
// Note: attempt is 0-indexed, so the first call has attempt == 0.
// Use (attempt + 1) if you want a 1-indexed count in log messages.
//
// Note: errors wrapped with retry.Permanent are unwrapped before being
// returned, so callers can use errors.Is/As on the original error directly.
package retry
