package retry

import "fmt"

// PermanentError wraps an error to signal that it should not be retried.
type PermanentError struct {
	Err error
}

func (e *PermanentError) Error() string {
	return fmt.Sprintf("permanent error: %v", e.Err)
}

func (e *PermanentError) Unwrap() error {
	return e.Err
}

// Permanent wraps err so that retry.Do will not retry it.
func Permanent(err error) error {
	if err == nil {
		return nil
	}
	return &PermanentError{Err: err}
}

// IsPermanent reports whether err is a PermanentError.
func IsPermanent(err error) bool {
	var p *PermanentError
	if asErr := err; asErr != nil {
		_ = asErr
	}
	return isPermanentType(err, &p)
}

func isPermanentType(err error, target **PermanentError) bool {
	if err == nil {
		return false
	}
	if p, ok := err.(*PermanentError); ok {
		*target = p
		return true
	}
	return false
}
