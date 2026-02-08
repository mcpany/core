// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/mcpany/core/server/pkg/logging"
)

// JSONRPCError represents a JSON-RPC 2.0 error object.
//
// Summary: represents a JSON-RPC 2.0 error object.
type JSONRPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

// JSONRPCResponse represents a JSON-RPC 2.0 response object.
//
// Summary: represents a JSON-RPC 2.0 response object.
type JSONRPCResponse struct {
	JSONRPC string        `json:"jsonrpc"`
	ID      any           `json:"id"`
	Error   *JSONRPCError `json:"error,omitempty"`
}

// JSONRPCComplianceMiddleware ensures that errors are returned as valid JSON-RPC responses.
//
// Summary: ensures that errors are returned as valid JSON-RPC responses.
//
// Parameters:
//   - next: http.Handler. The next.
//
// Returns:
//   - http.Handler: The http.Handler.
func JSONRPCComplianceMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Only intercept POST requests (likely JSON-RPC)
		if r.Method != http.MethodPost {
			next.ServeHTTP(w, r)
			return
		}

		// Don't intercept if it's explicitly SSE
		if strings.Contains(r.Header.Get("Accept"), "text/event-stream") && !strings.Contains(r.Header.Get("Accept"), "application/json") {
			next.ServeHTTP(w, r)
			return
		}

		// Don't intercept gRPC requests
		if strings.HasPrefix(r.Header.Get("Content-Type"), "application/grpc") {
			next.ServeHTTP(w, r)
			return
		}

		rw := &smartResponseWriter{
			w:      w,
			header: make(http.Header),
		}

		next.ServeHTTP(rw, r)

		// If headers weren't written, default to 200 OK
		if !rw.committed {
			rw.WriteHeader(http.StatusOK)
		}

		// If we passed through, we are done
		if rw.passThrough {
			return
		}

		// If we are here, we have a buffered error response that needs rewriting
		rw.rewriteError()
	})
}

const maxErrorBufferSize = 32 * 1024 // 32KB limit for error buffering

type smartResponseWriter struct {
	w           http.ResponseWriter
	header      http.Header
	statusCode  int
	body        *bytes.Buffer
	committed   bool
	passThrough bool
}

// Header returns the header map that will be sent by WriteHeader.
//
// Summary: returns the header map that will be sent by WriteHeader.
//
// Parameters:
//   None.
//
// Returns:
//   - http.Header: The http.Header.
func (w *smartResponseWriter) Header() http.Header {
	return w.header
}

// WriteHeader sends an HTTP response header with the provided status code.
//
// Summary: sends an HTTP response header with the provided status code.
//
// Parameters:
//   - code: int. The code.
//
// Returns:
//   None.
func (w *smartResponseWriter) WriteHeader(code int) {
	if w.committed {
		return
	}
	w.committed = true
	w.statusCode = code

	// We pass through if:
	// 1. Success (code < 400)
	// 2. Already JSON (Content-Type contains application/json)
	contentType := w.header.Get("Content-Type")
	isJSON := strings.Contains(contentType, "application/json")

	if code < 400 || isJSON {
		w.passThrough = true
		w.flushHeader()
	} else {
		// Buffer for rewriting
		w.body = &bytes.Buffer{}
	}
}

// Write writes the data to the connection as part of an HTTP reply.
//
// Summary: writes the data to the connection as part of an HTTP reply.
//
// Parameters:
//   - b: []byte. The b.
//
// Returns:
//   - int: The int.
//   - error: An error if the operation fails.
//
// Throws/Errors:
//   Returns an error if the operation fails.
func (w *smartResponseWriter) Write(b []byte) (int, error) {
	if !w.committed {
		w.WriteHeader(http.StatusOK)
	}

	if w.passThrough {
		return w.w.Write(b)
	}

	// Check buffer limit
	if w.body.Len()+len(b) > maxErrorBufferSize {
		// Too large to rewrite, switch to pass-through
		logging.GetLogger().Warn("Error response too large for JSON-RPC rewrite, streaming raw response", "size", w.body.Len()+len(b))
		w.passThrough = true
		w.flushHeader()
		// Write what we have so far
		if w.body.Len() > 0 {
			_, _ = w.w.Write(w.body.Bytes())
			w.body.Reset()
		}
		return w.w.Write(b)
	}

	return w.body.Write(b)
}

func (w *smartResponseWriter) flushHeader() {
	// Copy headers
	for k, v := range w.header {
		for _, vv := range v {
			w.w.Header().Add(k, vv)
		}
	}
	w.w.WriteHeader(w.statusCode)
}

// Flush implements http.Flusher to support streaming.
//
// Summary: implements http.Flusher to support streaming.
//
// Parameters:
//   None.
//
// Returns:
//   None.
func (w *smartResponseWriter) Flush() {
	if w.passThrough {
		if f, ok := w.w.(http.Flusher); ok {
			f.Flush()
		}
	}
}

func (w *smartResponseWriter) rewriteError() {
	bodyStr := strings.TrimSpace(w.body.String())

	// Determine error code based on body content or status
	code := -32603 // Internal error default
	message := bodyStr
	if message == "" {
		message = http.StatusText(w.statusCode)
	}

	messageLower := strings.ToLower(message)
	switch {
	case strings.Contains(messageLower, "parse error") || strings.Contains(messageLower, "malformed") || strings.Contains(messageLower, "invalid character"):
		code = -32700
		message = "Parse error"
	case strings.Contains(messageLower, "invalid request") || message == "Bad Request" || strings.Contains(messageLower, "accept must contain"):
		code = -32600
		message = "Invalid Request"
	case strings.Contains(messageLower, "method not found") || strings.Contains(messageLower, "not handled") || strings.Contains(messageLower, "unsupported"):
		code = -32601
		message = "Method not found"
	case strings.Contains(messageLower, "invalid params"):
		code = -32602
		message = "Invalid params"
	}

	// Sentinel Security: Sanitize sensitive data for internal errors (5xx).
	// We prevent leaking stack traces or sensitive upstream errors in the Data field.
	var data any = bodyStr
	if w.statusCode >= 500 {
		message = "Internal error"
		data = nil
	}

	resp := JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      nil,
		Error: &JSONRPCError{
			Code:    code,
			Message: message,
			Data:    data,
		},
	}

	w.w.Header().Set("Content-Type", "application/json")
	w.w.WriteHeader(w.statusCode) // Keep original status code (e.g. 400)
	_ = json.NewEncoder(w.w).Encode(resp)
}
