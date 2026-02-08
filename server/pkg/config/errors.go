// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"errors"
	"fmt"
)

// ActionableError is an error that includes a suggestion for fixing the issue.
//
// Summary:
//   An error type that attaches a fix suggestion to the underlying error.
//
// Fields:
//   - Err: error. The original error.
//   - Suggestion: string. A user-friendly suggestion to resolve the error.
type ActionableError struct {
	Err        error
	Suggestion string
}

// Error implements the error interface.
//
// Summary:
//   Returns the string representation of the error, including the suggestion.
//
// Returns:
//   - string: The error message with the fix suggestion appended.
func (e *ActionableError) Error() string {
	return fmt.Sprintf("%v\n\t-> Fix: %s", e.Err, e.Suggestion)
}

// Unwrap returns the underlying error.
//
// Summary:
//   Retrieves the original error wrapped by this ActionableError.
//
// Returns:
//   - error: The underlying error.
func (e *ActionableError) Unwrap() error {
	return e.Err
}

// WrapActionableError wraps an error with context, preserving ActionableError semantics if present.
//
// Summary:
//   Wraps an error while preserving any existing fix suggestions.
//   If the cause is an ActionableError, it returns a new ActionableError with the context added.
//   Otherwise, it returns a standard wrapped error.
//
// Parameters:
//   - context: string. The context message to prepend.
//   - err: error. The error to wrap.
//
// Returns:
//   - error: The wrapped error.
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
