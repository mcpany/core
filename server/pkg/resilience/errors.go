// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package resilience

// PermanentError is an error that should not be retried.
//
// Summary: Wrapper error indicating that an operation failed permanently and should not be retried.
type PermanentError struct {
	Err error
}

// Error provides error functionality.
//
// Summary: Error.
//
// Parameters.
//   - None.
//
// Returns.
//   - result: The result.
//
// Parameters:
//   - None.
//
// Returns:
//   - string.
func (e *PermanentError) Error() string {
	if e.Err == nil {
		return "permanent error"
	}
	return e.Err.Error()
}

// Unwrap provides unwrap functionality.
//
// Summary: Unwrap.
//
// Parameters.
//   - None.
//
// Returns.
//   - result: The result.
//
// Parameters:
//   - None.
//
// Returns:
//   - error.
func (e *PermanentError) Unwrap() error {
	return e.Err
}
