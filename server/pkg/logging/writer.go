package logging

import (
	"io"

	"github.com/mcpany/core/server/pkg/util"
)

// RedactingWriter is an io.Writer that redacts sensitive information from JSON logs.
type RedactingWriter struct {
	w io.Writer
}

// Write implements io.Writer.
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
