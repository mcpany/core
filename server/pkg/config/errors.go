// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"errors"
	"fmt"
)

// ActionableError is an error that includes a suggestion for fixing the issue.
//
// Summary: Represents an error that provides a user-friendly suggestion for resolution.
type ActionableError struct {
	Err        error
	Suggestion string
}

// Error implements the error interface.
//
// Summary: Returns the string representation of the error.
//
// Returns:
//   - string: The error message including the suggestion.
func (e *ActionableError) Error() string {
	return fmt.Sprintf("%v\n\t-> Fix: %s", e.Err, e.Suggestion)
}

// Unwrap returns the underlying error.
//
// Summary: Retrieves the original error wrapped by ActionableError.
//
// Returns:
//   - error: The original error wrapped by ActionableError.
func (e *ActionableError) Unwrap() error {
	return e.Err
}

// WrapActionableError wraps an error with context, preserving ActionableError semantics if present.
//
// Summary: Wraps an error with additional context while maintaining the ActionableError type if the underlying error is actionable.
//
// Parameters:
//   - context: string. Descriptive context to add to the error message.
//   - err: error. The original error to wrap.
//
// Returns:
//   - error: A new ActionableError if the cause was actionable, otherwise a standard wrapped error.
//
// Errors/Throws:
//   - Returns nil if the input error is nil.
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
