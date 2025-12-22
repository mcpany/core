// Package cli provides a JSON executor for CLI commands.

package cli

import (
	"encoding/json"
	"fmt"
	"io"
)

// JSONExecutor is a struct that sends JSON-encoded data to a writer and decodes
// JSON-encoded data from a reader.
type JSONExecutor struct {
	// in is the writer to which the JSON-encoded data is sent.
	in io.Writer
	// out is the reader from which the JSON-encoded data is read.
	out io.Reader
}

// NewJSONExecutor creates a new JSONExecutor with the given writer and reader.
func NewJSONExecutor(in io.Writer, out io.Reader) *JSONExecutor {
	return &JSONExecutor{
		in:  in,
		out: out,
	}
}

// Execute sends the given data as a JSON-encoded message to the writer and
// decodes the JSON-encoded response from the reader into the given result.
func (e *JSONExecutor) Execute(data, result any) error {
	if err := json.NewEncoder(e.in).Encode(data); err != nil {
		return fmt.Errorf("failed to encode data: %w", err)
	}

	if err := json.NewDecoder(e.out).Decode(result); err != nil {
		return fmt.Errorf("failed to decode data: %w", err)
	}

	return nil
}
