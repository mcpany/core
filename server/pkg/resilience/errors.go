// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package resilience

// PermanentError - Auto-generated documentation.
//
// Summary: PermanentError is an error that should not be retried.
//
// Fields:
//   - Various fields for PermanentError.
type PermanentError struct {
	Err error
}

// Error - Auto-generated documentation.
//
// Summary: Error returns the error message.
//
// Parameters:
//   - args: Variable arguments.
//
// Returns:
//   - result: The result of the operation.
//
// Errors:
//   - Returns an error if the operation fails.
//
// Side Effects:
//   - May modify internal state or perform external calls.
func (e *PermanentError) Error() string {
	if e.Err == nil {
		return "permanent error"
	}
	return e.Err.Error()
}

// Unwrap - Auto-generated documentation.
//
// Summary: Unwrap returns the wrapped error.
//
// Parameters:
//   - args: Variable arguments.
//
// Returns:
//   - result: The result of the operation.
//
// Errors:
//   - Returns an error if the operation fails.
//
// Side Effects:
//   - May modify internal state or perform external calls.
func (e *PermanentError) Unwrap() error {
	return e.Err
}
