// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"errors"
	"fmt"
)

// ActionableError is an error that includes a suggestion for fixing the issue.
//
// Summary: An error type that pairs an underlying error with a user-facing suggestion.
//
// Fields:
//   - Err: error. The original error that occurred.
//   - Suggestion: string. A human-readable suggestion on how to resolve the error.
type ActionableError struct {
	Err        error
	Suggestion string
}

// Error provides error functionality.
//
// Summary: Error.
//
// Parameters.
//   - None.
//
// Returns.
//   - result: The result.
func (e *ActionableError) Error() string {
	return fmt.Sprintf("%v\n\t-> Fix: %s", e.Err, e.Suggestion)
}

// Unwrap provides unwrap functionality.
//
// Summary: Unwrap.
//
// Parameters.
//   - None.
//
// Returns.
//   - result: The result.
func (e *ActionableError) Unwrap() error {
	return e.Err
}

// WrapActionableError provides wrapactionableerror functionality.
//
// Summary: WrapActionableError.
//
// Parameters.
//   - context: The parameter.
//   - err: The parameter.
//
// Returns.
//   - result: The result.
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
