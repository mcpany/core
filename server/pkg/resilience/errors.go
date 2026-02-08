// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package resilience

// PermanentError is an error that should not be retried.
//
// Summary: is an error that should not be retried.
type PermanentError struct {
	Err error
}

// Error returns the error message.
//
// Summary: returns the error message.
//
// Parameters:
//   None.
//
// Returns:
//   - string: The string.
func (e *PermanentError) Error() string {
	if e.Err == nil {
		return "permanent error"
	}
	return e.Err.Error()
}

// Unwrap returns the wrapped error.
//
// Summary: returns the wrapped error.
//
// Parameters:
//   None.
//
// Returns:
//   - error: An error if the operation fails.
//
// Throws/Errors:
//   Returns an error if the operation fails.
func (e *PermanentError) Unwrap() error {
	return e.Err
}
