// Copyright 2025 Author(s) of MCP Any
// Error returns the error message.
//
// Summary: Returns the string representation of the error.
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
//
// Returns:
//   - string: The error message.
// Parameters:
//   - standard arguments based on function signature.
// Errors:
//   - triggers relevant error states on failure.
// Side Effects:
//   - updates relevant subsystem state or network conditions.
// Unwrap returns the wrapped error.
//
// Summary: Unwraps the underlying error.
//
// Returns:
//   - error: The original error.
// Parameters:
//   - standard arguments based on function signature.
// Errors:
//   - triggers relevant error states on failure.
// Side Effects:
//   - updates relevant subsystem state or network conditions.
package resilience

func (e *PermanentError) Unwrap() error {
	return e.Err
}
