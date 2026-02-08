// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package logging

import (
	"io"

	"github.com/mcpany/core/server/pkg/util"
)

// RedactingWriter is an io.Writer that redacts sensitive information from JSON logs.
//
// Summary: is an io.Writer that redacts sensitive information from JSON logs.
type RedactingWriter struct {
	w io.Writer
}

// Write implements io.Writer.
//
// Summary: implements io.Writer.
//
// Parameters:
//   - p: []byte. The p.
//
// Returns:
//   - n: int. The int.
//   - err: error. An error if the operation fails.
//
// Throws/Errors:
//   Returns an error if the operation fails.
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
