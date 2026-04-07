// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package resilience

// Summary: PermanentError represents a data structure.
//
// Parameters:
//   - None
//
// Returns:
//   - None
//
// Errors:
//   - None
//
// Side Effects:
//   - None
type PermanentError struct {
	Err error
}

// Error returns the error message.
//
// Summary: Error executes the operation.
//
// Parameters:
//   - None
//
// Returns:
//   - string {
: Result of the operation.
//
// Errors:
//   - None
//
// Side Effects:
//   - None
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
// Summary: Unwrap executes the operation.
//
// Parameters:
//   - None
//
// Returns:
//   - error {
: Result of the operation.
//
// Errors:
//   - Returns an error if the operation fails.
//
// Side Effects:
//   - None
func (e *PermanentError) Unwrap() error {
	return e.Err
}
