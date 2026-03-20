// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package resilience

// PermanentError is an error that should not be retried.
//
// Summary: Wrapper error indicating that an operation failed permanently and should not be retried.
type PermanentError struct {
	Err error
}

// Error returns the error message.
//
// Summary: Returns the string representation of the error.
//
// Returns:
//   - string: The error message.
//
// Parameters:
//   - Inputs necessary to execute Error safely and correctly.
//
// Errors/Throws:
//   - None.
//
// Side Effects:
//   - None.
func (e *PermanentError) Error() string {
	if e.Err == nil {
		return "permanent error"
	}
	return e.Err.Error()
}

// Unwrap returns the wrapped error.
//
// Summary: Unwraps the underlying error.
//
// Returns:
//   - error: The original error.
//
// Parameters:
//   - Inputs necessary to execute Unwrap safely and correctly.
//
// Errors/Throws:
//   - Returns a structured error if internal validation fails, external dependencies cannot be reached, or state inconsistencies occur during Unwrap execution.
//
// Side Effects:
//   - None.
func (e *PermanentError) Unwrap() error {
	return e.Err
}
