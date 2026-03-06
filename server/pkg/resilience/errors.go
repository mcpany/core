// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package resilience

// PermanentError is an error that should not be retried. Summary: Wrapper error indicating that an operation failed permanently and should not be retried.
//
// Summary: PermanentError is an error that should not be retried. Summary: Wrapper error indicating that an operation failed permanently and should not be retried.
//
// Fields:
//   - Contains the configuration and state properties required for PermanentError functionality.
type PermanentError struct {
	Err error
}

// Error returns the error message. Summary: Returns the string representation of the error. Returns: - string: The error message.
//
// Summary: Error returns the error message. Summary: Returns the string representation of the error. Returns: - string: The error message.
//
// Parameters:
//   - None.
//
// Returns:
//   - (string): A string value representing the operation's result.
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

// Unwrap returns the wrapped error. Summary: Unwraps the underlying error. Returns: - error: The original error.
//
// Summary: Unwrap returns the wrapped error. Summary: Unwraps the underlying error. Returns: - error: The original error.
//
// Parameters:
//   - None.
//
// Returns:
//   - (error): An error object if the operation fails, otherwise nil.
//
// Errors:
//   - Returns an error if the underlying operation fails or encounters invalid input.
//
// Side Effects:
//   - None.
func (e *PermanentError) Unwrap() error {
	return e.Err
}
