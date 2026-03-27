// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

// Package cli provides a JSON executor for CLI commands.
package cli

import (
	"encoding/json"
	"fmt"
	"io"
)

// JSONExecutor is a struct that sends JSON-encoded data to a writer and decodes JSON-encoded data from a reader.
//
// Summary: JSONExecutor is a struct that sends JSON-encoded data to a writer and decodes JSON-encoded data from a reader.
type JSONExecutor struct {
	// in is the writer where JSON commands are written to (e.g. stdin of a process).
	in io.Writer
	// out is the reader where JSON responses are read from (e.g. stdout of a process).
	out io.Reader
}

// NewJSONExecutor provides newjsonexecutor functionality.
//
// Summary: NewJSONExecutor.
//
// Parameters.
//   - in: The parameter.
//   - out: The parameter.
//
// Returns.
//   - result: The result.
func NewJSONExecutor(in io.Writer, out io.Reader) *JSONExecutor {
	return &JSONExecutor{
		in:  in,
		out: out,
	}
}

// Execute provides execute functionality.
//
// Summary: Execute.
//
// Parameters.
//   - data: The parameter.
//   - result: The parameter.
//
// Returns.
//   - result: The result.
func (e *JSONExecutor) Execute(data, result any) error {
	if err := json.NewEncoder(e.in).Encode(data); err != nil {
		return fmt.Errorf("failed to encode data: %w", err)
	}

	if err := json.NewDecoder(e.out).Decode(result); err != nil {
		return fmt.Errorf("failed to decode data: %w", err)
	}

	return nil
}
