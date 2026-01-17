// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package resilience

// PermanentError is an error that should not be retried.
type PermanentError struct {
	Err error
}

// Error returns the error message.
//
// Returns the result.
func (e *PermanentError) Error() string {
	if e.Err == nil {
		return "permanent error"
	}
	return e.Err.Error()
}

// Unwrap returns the wrapped error.
//
// Returns an error if the operation fails.
func (e *PermanentError) Unwrap() error {
	return e.Err
}
