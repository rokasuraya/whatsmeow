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
package retry
