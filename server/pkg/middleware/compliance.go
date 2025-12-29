// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/mcpany/core/pkg/logging"
)

// JSONRPCError represents a JSON-RPC 2.0 error object.
type JSONRPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

// JSONRPCResponse represents a JSON-RPC 2.0 response object.
type JSONRPCResponse struct {
	JSONRPC string        `json:"jsonrpc"`
	ID      any           `json:"id"`
	Error   *JSONRPCError `json:"error,omitempty"`
}

// JSONRPCComplianceMiddleware ensures that errors are returned as valid JSON-RPC responses.
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

		// We need to buffer the response to check if it's an error
		// Note: This might have performance impact for large responses, but errors are usually small.

		// Use a separate buffer for our capture, but we need to prevent writing to the real writer
		// if we are going to modify the response.
		// So we can't use the simple wrapper above if we want to REPLACE the body.
		// We have to NOT write to underlying writer until we decide.

		bufRec := &bufferedResponseWriter{
			ResponseWriter: w,
			statusCode:     200, // Default
			header:         make(http.Header),
			body:           &bytes.Buffer{},
		}

		next.ServeHTTP(bufRec, r)

		// Check if we need to modify the response
		if bufRec.statusCode >= 400 {
			contentType := bufRec.header.Get("Content-Type")
			// If it's not JSON, or if it's text/plain, we might need to wrap it.
			if !strings.Contains(contentType, "application/json") {
				bodyStr := strings.TrimSpace(bufRec.body.String())

				// Determine error code based on body content or status
				// This is heuristic-based because we lost the original context.
				code := -32603 // Internal error default
				message := bodyStr
				if message == "" {
					message = http.StatusText(bufRec.statusCode)
				}

				if strings.Contains(strings.ToLower(message), "parse error") || strings.Contains(strings.ToLower(message), "malformed") || strings.Contains(strings.ToLower(message), "invalid character") {
					code = -32700
					message = "Parse error"
				} else if strings.Contains(strings.ToLower(message), "invalid request") || message == "Bad Request" {
					code = -32600
					message = "Invalid Request"
				} else if strings.Contains(strings.ToLower(message), "method not found") || strings.Contains(strings.ToLower(message), "not handled") || strings.Contains(strings.ToLower(message), "unsupported") {
					code = -32601
					message = "Method not found"
				} else if strings.Contains(strings.ToLower(message), "invalid params") {
					code = -32602
					message = "Invalid params"
				}

				// Try to extract ID from request if possible (best effort)
				// We can read the request body again if we buffered it, but we didn't.
				// So ID is null.

				resp := JSONRPCResponse{
					JSONRPC: "2.0",
					ID:      nil,
					Error: &JSONRPCError{
						Code:    code,
						Message: message,
						Data:    bodyStr, // Include original message as data
					},
				}

				logging.GetLogger().Info("Rewriting error response to JSON-RPC", "original_status", bufRec.statusCode, "original_body", bodyStr, "new_code", code)

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(bufRec.statusCode) // Keep original status code (e.g. 400)
				json.NewEncoder(w).Encode(resp)
				return
			}
		}

		// Write original response
		for k, v := range bufRec.header {
			for _, vv := range v {
				w.Header().Add(k, vv)
			}
		}
		w.WriteHeader(bufRec.statusCode)
		w.Write(bufRec.body.Bytes())
	})
}

type bufferedResponseWriter struct {
	http.ResponseWriter
	statusCode int
	header     http.Header
	body       *bytes.Buffer
}

func (w *bufferedResponseWriter) Header() http.Header {
	return w.header
}

func (w *bufferedResponseWriter) WriteHeader(code int) {
	w.statusCode = code
}

func (w *bufferedResponseWriter) Write(b []byte) (int, error) {
	return w.body.Write(b)
}
