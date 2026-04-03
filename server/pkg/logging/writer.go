// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package logging

import (
	"io"

	"github.com/mcpany/core/server/pkg/util"
)

// RedactingWriter represents the public RedactingWriter entity.
//
// Summary: Defines the structured data model representing a writer.
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
type RedactingWriter struct {
	w io.Writer
}

// Write serves as a public interface for interacting with Write.
//
// Summary: Write the  appropriately based on current system conditions.
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
