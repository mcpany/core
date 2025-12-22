package resilience

// PermanentError is an error that should not be retried.
type PermanentError struct {
	Err error
}

// Error returns the error message.
func (e *PermanentError) Error() string {
	if e.Err == nil {
		return "permanent error"
	}
	return e.Err.Error()
}

// Unwrap returns the wrapped error.
func (e *PermanentError) Unwrap() error {
	return e.Err
}
