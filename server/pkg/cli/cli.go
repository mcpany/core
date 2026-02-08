// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

// Package cli provides a JSON executor for CLI commands.
package cli

import (
	"encoding/json"
	"fmt"
	"io"
)

// JSONExecutor is a struct that sends JSON-encoded data to a writer and decodes.
//
// Summary: is a struct that sends JSON-encoded data to a writer and decodes.
type JSONExecutor struct {
	// in is the writer where JSON commands are written to (e.g. stdin of a process).
	in io.Writer
	// out is the reader where JSON responses are read from (e.g. stdout of a process).
	out io.Reader
}

// NewJSONExecutor creates a new JSONExecutor with the given writer and reader.
//
// Summary: creates a new JSONExecutor with the given writer and reader.
//
// Parameters:
//   - in: io.Writer. The in.
//   - out: io.Reader. The out.
//
// Returns:
//   - *JSONExecutor: The *JSONExecutor.
func NewJSONExecutor(in io.Writer, out io.Reader) *JSONExecutor {
	return &JSONExecutor{
		in:  in,
		out: out,
	}
}

// Execute sends the given data as a JSON-encoded message to the writer and.
//
// Summary: sends the given data as a JSON-encoded message to the writer and.
//
// Parameters:
//   - data: any. The data.
//   - result: any. The result.
//
// Returns:
//   - error: An error if the operation fails.
//
// Throws/Errors:
//   Returns an error if the operation fails.
func (e *JSONExecutor) Execute(data, result any) error {
	if err := json.NewEncoder(e.in).Encode(data); err != nil {
		return fmt.Errorf("failed to encode data: %w", err)
	}

	if err := json.NewDecoder(e.out).Decode(result); err != nil {
		return fmt.Errorf("failed to decode data: %w", err)
	}

	return nil
}
