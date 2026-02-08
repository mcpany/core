package config

import (
	"errors"
	"fmt"
)

// ActionableError is an error that includes a suggestion for fixing the issue.
type ActionableError struct {
	Err        error
	Suggestion string
}

// Error implements the error interface.
//
// Returns the error message including the suggestion.
func (e *ActionableError) Error() string {
	return fmt.Sprintf("%v\n\t-> Fix: %s", e.Err, e.Suggestion)
}

// Unwrap returns the underlying error.
//
// Returns the original error wrapped by ActionableError.
func (e *ActionableError) Unwrap() error {
	return e.Err
}

// WrapActionableError wraps an error with context, preserving ActionableError semantics if present.
// If the cause is an ActionableError, it returns a new ActionableError with the context added to the error message.
// Otherwise, it returns a standard wrapped error.
func WrapActionableError(context string, err error) error {
	if err == nil {
		return nil
	}
	var ae *ActionableError
	if errors.As(err, &ae) {
		// Preserve the ActionableError type so it can be detected upstream (e.g. by LoadServices).
		// We wrap the entire original error chain to preserve all context.
		// Note: This may result in the suggestion being printed twice if the inner ActionableError's .Error()
		// is also displayed, but it ensures no error context is lost.
		return &ActionableError{
			Err:        fmt.Errorf("%s: %w", context, err),
			Suggestion: ae.Suggestion,
		}
	}
	return fmt.Errorf("%s: %w", context, err)
}
