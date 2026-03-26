// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package resilience

// PermanentError permanentError represents a permanent error.
//
// Summary: PermanentError represents a permanent error.
type PermanentError struct {
	Err error
}

// Error error error.
//
// Summary: Error error.
//
// Parameters:
//   - None.
//
// Returns:
//   - string: The result.
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

// Unwrap unwrap unwrap.
//
// Summary: Unwrap unwrap.
//
// Parameters:
//   - None.
//
// Returns:
//   - error: An error if the operation fails.
//
// Errors:
//   - Returns an error if the operation fails.
//
// Side Effects:
//   - None.
func (e *PermanentError) Unwrap() error {
	return e.Err
}
