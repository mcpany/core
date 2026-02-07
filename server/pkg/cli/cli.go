// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

// Package cli provides a JSON executor for CLI commands.
package cli

import (
	"encoding/json"
	"fmt"
	"io"
)

// JSONExecutor is a struct that sends JSON-encoded data to a writer and decodes
// JSON-encoded data from a reader.
//
// Summary: Handles JSON communication over IO streams.
type JSONExecutor struct {
	// in is the writer where JSON commands are written to (e.g. stdin of a process).
	in io.Writer
	// out is the reader where JSON responses are read from (e.g. stdout of a process).
	out io.Reader
}

// NewJSONExecutor creates a new JSONExecutor with the given writer and reader.
//
// Summary: Initializes a new JSONExecutor.
//
// Parameters:
//   - in: io.Writer. The destination for writing JSON requests.
//   - out: io.Reader. The source for reading JSON responses.
//
// Returns:
//   - *JSONExecutor: A new JSONExecutor instance.
func NewJSONExecutor(in io.Writer, out io.Reader) *JSONExecutor {
	return &JSONExecutor{
		in:  in,
		out: out,
	}
}

// Execute sends the given data as a JSON-encoded message to the writer and
// decodes the JSON-encoded response from the reader into the given result.
//
// Summary: Encodes data to JSON, writes it, and decodes the response.
//
// Parameters:
//   - data: any. The data to encode and send.
//   - result: any. The pointer to decode the response into.
//
// Returns:
//   - error: An error if encoding or decoding fails.
//
// Side Effects:
//   - Writes to the 'in' writer.
//   - Reads from the 'out' reader.
func (e *JSONExecutor) Execute(data, result any) error {
	if err := json.NewEncoder(e.in).Encode(data); err != nil {
		return fmt.Errorf("failed to encode data: %w", err)
	}

	if err := json.NewDecoder(e.out).Decode(result); err != nil {
		return fmt.Errorf("failed to decode data: %w", err)
	}

	return nil
}
