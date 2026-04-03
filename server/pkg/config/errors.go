// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"errors"
	"fmt"
)

// ActionableError represents the public ActionableError entity.
//
// Summary: Defines the structured data model representing a error.
//
// Parameters:
//   - None.
//
// Returns:
//   - None.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
type ActionableError struct {
	Err        error
	Suggestion string
}

// Error serves as a public interface for interacting with Error.
//
// Summary: Error the  appropriately based on current system conditions.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the successfully computed domain model or execution state.
//
// Errors:
//   - No explicit errors are thrown by this operation.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
func (e *ActionableError) Error() string {
	return fmt.Sprintf("%v\n\t-> Fix: %s", e.Err, e.Suggestion)
}

// Unwrap serves as a public interface for interacting with Unwrap.
//
// Summary: Unwrap the  appropriately based on current system conditions.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the expected domain model and an error upon failure.
//
// Errors:
//   - Propagates exceptions from underlying I/O or validation layers.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
func (e *ActionableError) Unwrap() error {
	return e.Err
}

// WrapActionableError serves as a public interface for interacting with WrapActionableError.
//
// Summary: Wrap the actionable error appropriately based on current system conditions.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the expected domain model and an error upon failure.
//
// Errors:
//   - Propagates exceptions from underlying I/O or validation layers.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
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
