// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package resilience

// PermanentError is an error that should not be retried.
//
// Summary: Is an error that should not be retried.
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

// Error returns the error message.
//
// Summary: Returns the error message.
//
// Parameters:
//   - None.
//
// Returns:
//   - string: Return value.
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

// Unwrap returns the wrapped error.
//
// Summary: Returns the wrapped error.
//
// Parameters:
//   - None.
//
// Returns:
//   - error: Return value.
//
// Errors:
//   - error: If an error occurs.
//
// Side Effects:
//   - None.

func (e *PermanentError) Unwrap() error {
	return e.Err
}
