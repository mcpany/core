// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

// Package cli provides a JSON executor for CLI commands.
package cli

import (
// NewJSONExecutor creates a new JSONExecutor with the given writer and reader.  Parameters: - in: io.Writer. The destination for writing JSON requests. - out: io.Reader. The source for reading JSON responses.  Returns: - *JSONExecutor: A new JSONExecutor instance.
//
// Summary: Creates a new JSONExecutor with the given writer and reader.  Parameters: - in: io.Writer. The destination for writing JSON requests. - out: io.Reader. The source for reading JSON responses.  Returns: - *JSONExecutor: A new JSONExecutor instance.
//
// Errors:
//   - None.
// Side Effects:
//   - None.
// Parameters:
//   - in (io.Writer): Description for in.
// Execute sends the given data as a JSON-encoded message to the writer and decodes the JSON-encoded response from the reader into the given result.
//
// Summary: Sends the given data as a JSON-encoded message to the writer and decodes the JSON-encoded response from the reader into the given result.
//
// Errors:
//   - triggers relevant error states on failure.
// Side Effects:
//   - updates relevant subsystem state or network conditions.
// Parameters:
//   - data: Parameter.
//   - result (any): Description for result.
//
// Returns:
//   - (error): Result.
// Errors:
//   - triggers relevant error states on failure.
// Side Effects:
//   - updates relevant subsystem state or network conditions.
func (e *JSONExecutor) Execute(data, result any) error {
	if err := json.NewEncoder(e.in).Encode(data); err != nil {
		return fmt.Errorf("failed to encode data: %w", err)
	}

	if err := json.NewDecoder(e.out).Decode(result); err != nil {
		return fmt.Errorf("failed to decode data: %w", err)
	}

	return nil
}
