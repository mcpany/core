// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestJSONRPCComplianceMiddleware_LargeResponse ...
// Summary: TestJSONRPCComplianceMiddleware_LargeResponse
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	// Create a large response > 32KB
	largeData := strings.Repeat("a", maxErrorBufferSize+100)

	handler := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(largeData))
	}

	req := httptest.NewRequest(http.MethodPost, "/", nil)
	rec := httptest.NewRecorder()

	middleware := JSONRPCComplianceMiddleware(http.HandlerFunc(handler))
	middleware.ServeHTTP(rec, req)

	res := rec.Result()
	defer res.Body.Close()

	// Should NOT be rewritten because it is too large
	assert.Equal(t, http.StatusBadRequest, res.StatusCode)

	// Check Content-Type is not application/json unless set by handler (it wasn't)
	assert.NotContains(t, res.Header.Get("Content-Type"), "application/json")
}

// TestJSONRPCComplianceMiddleware_Flush ...
// Summary: TestJSONRPCComplianceMiddleware_Flush
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("part1"))
		if f, ok := w.(http.Flusher); ok {
			f.Flush()
		}
		w.Write([]byte("part2"))
	}

	req := httptest.NewRequest(http.MethodPost, "/", nil)
	rec := httptest.NewRecorder()

	middleware := JSONRPCComplianceMiddleware(http.HandlerFunc(handler))
	middleware.ServeHTTP(rec, req)

	res := rec.Result()
	defer res.Body.Close()

	assert.Equal(t, http.StatusOK, res.StatusCode)
	assert.Equal(t, "part1part2", rec.Body.String())
}

type mockFlusherResponseWriter struct {
	http.ResponseWriter
	flushed bool
}

// Flush ...
// Summary: Flush
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	m.flushed = true
}

// TestJSONRPCComplianceMiddleware_Flush_PassThrough ...
// Summary: TestJSONRPCComplianceMiddleware_Flush_PassThrough
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	rec := httptest.NewRecorder()
	mockWriter := &mockFlusherResponseWriter{ResponseWriter: rec}

	handler := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK) // Pass through because code < 400
		if f, ok := w.(http.Flusher); ok {
			f.Flush()
		}
	}

	// We need to bypass httptest.NewRecorder because it implements Flusher,
	// but we want to inspect if Flush was called on our mock.
	// However, the middleware wraps the writer.

	middleware := JSONRPCComplianceMiddleware(http.HandlerFunc(handler))

	req := httptest.NewRequest(http.MethodPost, "/", nil)
	// We need to pass mockWriter to the middleware
	// But middleware returns a Handler.

	middleware.ServeHTTP(mockWriter, req)

	assert.True(t, mockWriter.flushed)
}
