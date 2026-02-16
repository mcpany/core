// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGzipCompressionMiddleware_Flush(t *testing.T) {
	// 1. Setup a handler that attempts to flush
	flushed := false
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		_, _ = w.Write([]byte("data: event1\n\n"))

		if f, ok := w.(http.Flusher); ok {
			f.Flush()
			flushed = true
		} else {
			// This is failure for now
			t.Log("ResponseWriter does not implement http.Flusher")
		}
	})

	// 2. Wrap with middleware
	gzipHandler := GzipCompressionMiddleware(handler)

	// 3. Execute request
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Accept-Encoding", "gzip")
	rec := httptest.NewRecorder()

	gzipHandler.ServeHTTP(rec, req)

	// 4. Verify Flush support
	// Current behavior: flushed should be false because middleware hides Flush()
	// Desired behavior: flushed should be true
	if !flushed {
		t.Error("Test confirms: Middleware does NOT support Flush() - w.(http.Flusher) failed")
	}

	// We also want to verify that the underlying ResponseWriter.Flush() was called.
	// httptest.ResponseRecorder tracks this via Flushed boolean.
	if !rec.Flushed {
		t.Error("Underlying ResponseWriter.Flush() was NOT called")
	}
}
