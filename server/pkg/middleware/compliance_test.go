// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware_test

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/mcpany/core/pkg/middleware"
	"github.com/stretchr/testify/assert"
)

func TestJSONRPCComplianceMiddleware(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		requestBody    string
		requestHeaders map[string]string
		handler        http.HandlerFunc
		expectedCode   int
		expectedBody   string // Partial match or exact match depending on needs
		checkJSONRPC   bool   // Whether to check if body is valid JSON-RPC error
		expectedRPC    *middleware.JSONRPCResponse
	}{
		{
			name:   "Non-POST request passed through",
			method: http.MethodGet,
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte("ok"))
			},
			expectedCode: http.StatusOK,
			expectedBody: "ok",
		},
		{
			name:   "SSE request passed through",
			method: http.MethodPost,
			requestHeaders: map[string]string{
				"Accept": "text/event-stream",
			},
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte("event: message\ndata: ok\n\n"))
			},
			expectedCode: http.StatusOK,
			expectedBody: "event: message\ndata: ok\n\n",
		},
		{
			name:   "Success response passed through",
			method: http.MethodPost,
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`{"jsonrpc":"2.0","result":"ok","id":1}`))
			},
			expectedCode: http.StatusOK,
			expectedBody: `{"jsonrpc":"2.0","result":"ok","id":1}`,
		},
		{
			name:   "JSON error response passed through",
			method: http.MethodPost,
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusBadRequest)
				_, _ = w.Write([]byte(`{"jsonrpc":"2.0","error":{"code":-32600,"message":"Invalid Request"},"id":null}`))
			},
			expectedCode: http.StatusBadRequest,
			expectedBody: `{"jsonrpc":"2.0","error":{"code":-32600,"message":"Invalid Request"},"id":null}`,
		},
		{
			name:   "Non-JSON error response rewritten (400 Bad Request)",
			method: http.MethodPost,
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "text/plain")
				http.Error(w, "Bad Request", http.StatusBadRequest)
			},
			expectedCode: http.StatusBadRequest,
			checkJSONRPC: true,
			expectedRPC: &middleware.JSONRPCResponse{
				JSONRPC: "2.0",
				ID:      nil,
				Error: &middleware.JSONRPCError{
					Code:    -32600,
					Message: "Invalid Request",
					Data:    "Bad Request", // http.Error adds newline, but we trim space in the middleware
				},
			},
		},
		{
			name:   "Non-JSON error response rewritten (404 Not Found)",
			method: http.MethodPost,
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "text/plain")
				http.Error(w, "Method not found", http.StatusNotFound)
			},
			expectedCode: http.StatusNotFound,
			checkJSONRPC: true,
			expectedRPC: &middleware.JSONRPCResponse{
				JSONRPC: "2.0",
				ID:      nil,
				Error: &middleware.JSONRPCError{
					Code:    -32601,
					Message: "Method not found",
					Data:    "Method not found",
				},
			},
		},
		{
			name:   "Non-JSON error response rewritten (500 Internal Server Error)",
			method: http.MethodPost,
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "text/plain")
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			},
			expectedCode: http.StatusInternalServerError,
			checkJSONRPC: true,
			expectedRPC: &middleware.JSONRPCResponse{
				JSONRPC: "2.0",
				ID:      nil,
				Error: &middleware.JSONRPCError{
					Code:    -32603,
					Message: "Internal Server Error\n",
					Data:    "Internal Server Error",
				},
			},
		},
		{
			name:   "Parse error detection",
			method: http.MethodPost,
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusBadRequest)
				_, _ = w.Write([]byte("invalid character 'x' looking for beginning of value"))
			},
			expectedCode: http.StatusBadRequest,
			checkJSONRPC: true,
			expectedRPC: &middleware.JSONRPCResponse{
				JSONRPC: "2.0",
				ID:      nil,
				Error: &middleware.JSONRPCError{
					Code:    -32700,
					Message: "Parse error",
					Data:    "invalid character 'x' looking for beginning of value",
				},
			},
		},
		{
			name:   "Invalid params detection",
			method: http.MethodPost,
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusBadRequest)
				_, _ = w.Write([]byte("Invalid Params: missing x"))
			},
			expectedCode: http.StatusBadRequest,
			checkJSONRPC: true,
			expectedRPC: &middleware.JSONRPCResponse{
				JSONRPC: "2.0",
				ID:      nil,
				Error: &middleware.JSONRPCError{
					Code:    -32602,
					Message: "Invalid params",
					Data:    "Invalid Params: missing x",
				},
			},
		},
		{
			name:   "Empty error body uses status text",
			method: http.MethodPost,
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusBadRequest)
			},
			expectedCode: http.StatusBadRequest,
			checkJSONRPC: true,
			expectedRPC: &middleware.JSONRPCResponse{
				JSONRPC: "2.0",
				ID:      nil,
				Error: &middleware.JSONRPCError{
					Code:    -32600, // Matches "Bad Request"
					Message: "Invalid Request",
					Data:    "Bad Request",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := middleware.JSONRPCComplianceMiddleware(tt.handler)
			req := httptest.NewRequest(tt.method, "/", strings.NewReader(tt.requestBody))
			for k, v := range tt.requestHeaders {
				req.Header.Set(k, v)
			}
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			assert.Equal(t, tt.expectedCode, rec.Code)

			if tt.checkJSONRPC {
				var resp middleware.JSONRPCResponse
				err := json.Unmarshal(rec.Body.Bytes(), &resp)
				assert.NoError(t, err, "Response should be valid JSON")
				assert.Equal(t, tt.expectedRPC.JSONRPC, resp.JSONRPC)
				assert.Equal(t, tt.expectedRPC.ID, resp.ID)
				assert.Equal(t, tt.expectedRPC.Error.Code, resp.Error.Code)
				assert.Equal(t, tt.expectedRPC.Error.Message, resp.Error.Message)

				// Data match: exact match
				assert.Equal(t, tt.expectedRPC.Error.Data, resp.Error.Data)
				assert.Contains(t, rec.Header().Get("Content-Type"), "application/json")
			} else {
				if tt.expectedBody != "" {
					assert.Equal(t, tt.expectedBody, rec.Body.String())
				}
			}
		})
	}
}

func TestBufferedResponseWriter(t *testing.T) {
	// Since BufferedResponseWriter is not exported, we test it via the middleware behavior

	handler := middleware.JSONRPCComplianceMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Custom", "value")
		w.WriteHeader(http.StatusTeapot)
		_, _ = w.Write([]byte("I'm a teapot"))
	}))

	req := httptest.NewRequest(http.MethodPost, "/", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusTeapot, rec.Code)
	// The current middleware implementation drops headers when rewriting the response.
	assert.Equal(t, "", rec.Header().Get("X-Custom"))

	assert.Contains(t, rec.Header().Get("Content-Type"), "application/json")
	var resp middleware.JSONRPCResponse
	err := json.Unmarshal(rec.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, -32603, resp.Error.Code) // Default error
	assert.Equal(t, "I'm a teapot", resp.Error.Data)
}

func TestJSONRPCComplianceMiddleware_WriterTypes(t *testing.T) {
	// Ensure that the middleware handles writing correctly
	handler := middleware.JSONRPCComplianceMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, "OK")
	}))

	req := httptest.NewRequest(http.MethodPost, "/", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "OK", rec.Body.String())
}
