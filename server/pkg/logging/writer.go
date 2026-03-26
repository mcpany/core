// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package logging

import (
	"io"

	"github.com/mcpany/core/server/pkg/util"
)

// RedactingWriter redactingWriter represents a redacting writer.
//
// Summary: RedactingWriter represents a redacting writer.
type RedactingWriter struct {
	w io.Writer
}

// Write implements io.Writer.
//
// Parameters: - None.
//   - p ([]byte): The p parameter.
//
// Returns: - None.
//   - int: The resulting int.
//   - error: An error if the operation fails.
//
// Errors: - None.
//   - Returns an error if the operation fails or is invalid.
//
// Side Effects: - None.
//   - None.
//
// Summary: Updates Write operation.
//
// Parameters: - None.
//
// Returns: - None.
//
// Errors: - None.
//
// Side Effects: - None.
//   - None.
func (w *RedactingWriter) Write(p []byte) (n int, err error) {
	// Attempt to redact JSON. RedactJSON handles validation internally.
	// If it's not valid JSON (e.g. partial write), it returns original input.
	redacted := util.RedactJSON(p)

	_, err = w.w.Write(redacted)
	if err != nil {
		// We can't easily map the written bytes of 'redacted' back to 'p'.
		// So we return 0 and the error.
		return 0, err
	}

	// If successful, we claim to have written all of p.
	return len(p), nil
}
