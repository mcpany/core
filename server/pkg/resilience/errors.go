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
// Parameters:
//   - e: input parameter for Error.
//
// Errors:
//   - No specific errors documented.
//
// Side Effects:
//   - None known.
//
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
// Parameters:
//   - e: input parameter for Unwrap.
//
// Errors:
//   - Returns an error if the underlying operation fails or inputs are invalid.
//
// Side Effects:
//   - None known.
//
func (e *PermanentError) Unwrap() error {
	return e.Err
}
