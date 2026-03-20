// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

// Package resilience implements the resilience subsystem.
package resilience

// PermanentError is an error that should not be retried.
//
// Summary: Wrapper error indicating that an operation failed permanently and should not be retried.
type PermanentError struct {
	Err error
}

// Error handles error.
//
// Parameters:
//   - None
//
// Returns:
//   - string: The generated or retrieved entity.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
// Error returns the error message.
//
// Summary: Returns the string representation of the error.
//
// Returns:
//   - string: The error message.
func (e *PermanentError) Error() string {
	if e.Err == nil {
		return "permanent error"
	}
	return e.Err.Error()
}

// Unwrap handles unwrap.
//
// Parameters:
//   - None
//
// Returns:
//   - error: Returns an error if the execution fails or validation does not pass.
//
// Errors:
//   - Returns an error if the input is malformed, dependencies are unreachable, or state validation fails.
//
// Side Effects:
//   - None.
// Unwrap returns the wrapped error.
//
// Summary: Unwraps the underlying error.
//
// Returns:
//   - error: The original error.
func (e *PermanentError) Unwrap() error {
	return e.Err
}
