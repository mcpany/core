// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

// Package cli provides a JSON executor for CLI commands.
package cli

import (
	"encoding/json"
	"fmt"
	"io"
)

// JSONExecutor represents the public JSONExecutor entity.
//
// Summary: Defines the structured data model representing a executor.
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
type JSONExecutor struct {
	// in is the writer where JSON commands are written to (e.g. stdin of a process).
	in io.Writer
	// out is the reader where JSON responses are read from (e.g. stdout of a process).
	out io.Reader
}

// NewJSONExecutor serves as a public interface for interacting with NewJSONExecutor.
//
// Summary: Constructs and returns an initialized json executor ready for consumption.
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
func NewJSONExecutor(in io.Writer, out io.Reader) *JSONExecutor {
	return &JSONExecutor{
		in:  in,
		out: out,
	}
}

// Execute serves as a public interface for interacting with Execute.
//
// Summary: Execute the  appropriately based on current system conditions.
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
func (e *JSONExecutor) Execute(data, result any) error {
	if err := json.NewEncoder(e.in).Encode(data); err != nil {
		return fmt.Errorf("failed to encode data: %w", err)
	}

	if err := json.NewDecoder(e.out).Decode(result); err != nil {
		return fmt.Errorf("failed to decode data: %w", err)
	}

	return nil
}
