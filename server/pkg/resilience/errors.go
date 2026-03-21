// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package resilience

// Summary: PermanentError is an error that should not be retried. Wrapper error indicating that an operation failed permanently and should not be retried.
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
type PermanentError struct {
	Err error
}

// Summary: Error returns the error message. Returns the string representation of the error.
//
// Parameters:
//   - None.
//
// Returns:
//   - string: The resulting string.
//
// Errors:
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

// Summary: Unwrap returns the wrapped error. Unwraps the underlying error.
//
// Parameters:
//   - None.
//
// Returns:
//   - error: An error if the operation fails.
//
// Errors:
//   - Returns an error if the operation fails or is invalid.
//
// Side Effects:
//   - None.
func (e *PermanentError) Unwrap() error {
	return e.Err
}
