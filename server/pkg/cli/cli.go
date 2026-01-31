// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

// Package cli provides a JSON executor for CLI commands.
package cli

import (
	"encoding/json"
	"fmt"
	"io"
)

// JSONExecutor handles JSON-based communication between an input reader and an output writer.
//
// It is designed to facilitate command-line interactions or inter-process communication
// where messages are exchanged as JSON objects.
type JSONExecutor struct {
	// in is the writer to which the JSON-encoded data is sent.
	in io.Writer
	// out is the reader from which the JSON-encoded data is read.
	out io.Reader
}

// NewJSONExecutor creates a new JSONExecutor with the given writer and reader.
//
// Parameters:
//   - in: io.Writer. The destination for JSON-encoded output.
//   - out: io.Reader. The source for JSON-encoded input.
//
// Returns:
//   - *JSONExecutor: A new instance of JSONExecutor.
func NewJSONExecutor(in io.Writer, out io.Reader) *JSONExecutor {
	return &JSONExecutor{
		in:  in,
		out: out,
	}
}

// Execute sends the given data as a JSON-encoded message to the writer and
// decodes the JSON-encoded response from the reader into the given result.
//
// It first encodes the 'data' parameter to the executor's input writer, then
// waits to decode the response from the executor's output reader into 'result'.
//
// Parameters:
//   - data: any. The payload to encode and send.
//   - result: any. A pointer to the target struct/map where the response will be decoded.
//
// Returns:
//   - error: An error if encoding or decoding fails.
func (e *JSONExecutor) Execute(data, result any) error {
	if err := json.NewEncoder(e.in).Encode(data); err != nil {
		return fmt.Errorf("failed to encode data: %w", err)
	}

	if err := json.NewDecoder(e.out).Decode(result); err != nil {
		return fmt.Errorf("failed to decode data: %w", err)
	}

	return nil
}
