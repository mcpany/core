// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"errors"
	"fmt"
)

// ActionableError is an error that includes a suggestion for fixing the issue.
//
// Summary: is an error that includes a suggestion for fixing the issue.
type ActionableError struct {
	Err        error
	Suggestion string
}

// Error implements the error interface.
//
// Summary: implements the error interface.
//
// Parameters:
//   None.
//
// Returns:
//   - string: The string.
func (e *ActionableError) Error() string {
	return fmt.Sprintf("%v\n\t-> Fix: %s", e.Err, e.Suggestion)
}

// Unwrap returns the underlying error.
//
// Summary: returns the underlying error.
//
// Parameters:
//   None.
//
// Returns:
//   - error: An error if the operation fails.
//
// Throws/Errors:
//   Returns an error if the operation fails.
func (e *ActionableError) Unwrap() error {
	return e.Err
}

// WrapActionableError wraps an error with context, preserving ActionableError semantics if present.
//
// Summary: wraps an error with context, preserving ActionableError semantics if present.
//
// Parameters:
//   - context: string. The context.
//   - err: error. The err.
//
// Returns:
//   - error: An error if the operation fails.
//
// Throws/Errors:
//   Returns an error if the operation fails.
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
