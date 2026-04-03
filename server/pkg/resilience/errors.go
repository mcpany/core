// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package resilience

// PermanentError represents the public PermanentError entity.
//
// Summary: Defines the structured data model representing a error.
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

// Error serves as a public interface for interacting with Error.
//
// Summary: Error the  appropriately based on current system conditions.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the successfully computed domain model or execution state.
//
// Errors:
//   - No explicit errors are thrown by this operation.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
func (e *PermanentError) Error() string {
	if e.Err == nil {
		return "permanent error"
	}
	return e.Err.Error()
}

// Unwrap serves as a public interface for interacting with Unwrap.
//
// Summary: Unwrap the  appropriately based on current system conditions.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the expected domain model and an error upon failure.
//
// Errors:
//   - Propagates exceptions from underlying I/O or validation layers.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
func (e *PermanentError) Unwrap() error {
	return e.Err
}
