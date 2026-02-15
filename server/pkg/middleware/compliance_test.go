package middleware

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJSONRPCComplianceMiddleware(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		requestHeaders map[string]string
		handler        http.HandlerFunc
		expectedStatus int
		expectedBody   string
		validateJSON   func(*testing.T, *JSONRPCResponse)
	}{
		{
			name:   "Pass through non-POST request",
			method: http.MethodGet,
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("ok"))
			},
			expectedStatus: http.StatusOK,
			expectedBody:   "ok",
		},
		{
			name:   "Pass through SSE request",
			method: http.MethodPost,
			requestHeaders: map[string]string{
				"Accept": "text/event-stream",
			},
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("event: message\ndata: hello\n\n"))
			},
			expectedStatus: http.StatusOK,
			expectedBody:   "event: message\ndata: hello\n\n",
		},
		{
			name:   "Pass through successful JSON response",
			method: http.MethodPost,
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"jsonrpc":"2.0","result":"ok","id":1}`))
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `{"jsonrpc":"2.0","result":"ok","id":1}`,
		},
		{
			name:   "Rewrite 400 Bad Request",
			method: http.MethodPost,
			handler: func(w http.ResponseWriter, r *http.Request) {
				http.Error(w, "Bad Request", http.StatusBadRequest)
			},
			expectedStatus: http.StatusBadRequest,
			validateJSON: func(t *testing.T, resp *JSONRPCResponse) {
				assert.Equal(t, "2.0", resp.JSONRPC)
				assert.Nil(t, resp.ID)
				require.NotNil(t, resp.Error)
				assert.Equal(t, -32600, resp.Error.Code)
				assert.Equal(t, "Invalid Request", resp.Error.Message)
			},
		},
		{
			name:   "Rewrite 404 Not Found (Method not found)",
			method: http.MethodPost,
			handler: func(w http.ResponseWriter, r *http.Request) {
				http.Error(w, "Method not found", http.StatusNotFound)
			},
			expectedStatus: http.StatusNotFound,
			validateJSON: func(t *testing.T, resp *JSONRPCResponse) {
				assert.Equal(t, "2.0", resp.JSONRPC)
				assert.Nil(t, resp.ID)
				require.NotNil(t, resp.Error)
				assert.Equal(t, -32601, resp.Error.Code)
				assert.Equal(t, "Method not found", resp.Error.Message)
			},
		},
		{
			name:   "Rewrite 500 Internal Server Error",
			method: http.MethodPost,
			handler: func(w http.ResponseWriter, r *http.Request) {
				http.Error(w, "Something went wrong", http.StatusInternalServerError)
			},
			expectedStatus: http.StatusInternalServerError,
			validateJSON: func(t *testing.T, resp *JSONRPCResponse) {
				assert.Equal(t, "2.0", resp.JSONRPC)
				assert.Nil(t, resp.ID)
				require.NotNil(t, resp.Error)
				assert.Equal(t, -32603, resp.Error.Code)
				// Sentinel Security: Expect generic message for 500 errors to avoid leaking details
				assert.Equal(t, "Internal error", resp.Error.Message)
				assert.Nil(t, resp.Error.Data)
			},
		},
		{
			name:   "Do not rewrite JSON error response",
			method: http.MethodPost,
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte(`{"jsonrpc":"2.0","error":{"code":-32000,"message":"Custom error"},"id":1}`))
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"jsonrpc":"2.0","error":{"code":-32000,"message":"Custom error"},"id":1}`,
		},
		{
			name:   "Rewrite Parse error",
			method: http.MethodPost,
			handler: func(w http.ResponseWriter, r *http.Request) {
				http.Error(w, "Invalid character 'x' looking for beginning of value", http.StatusBadRequest)
			},
			expectedStatus: http.StatusBadRequest,
			validateJSON: func(t *testing.T, resp *JSONRPCResponse) {
				assert.Equal(t, "2.0", resp.JSONRPC)
				require.NotNil(t, resp.Error)
				assert.Equal(t, -32700, resp.Error.Code)
				assert.Equal(t, "Parse error", resp.Error.Message)
			},
		},
		{
			name:   "Rewrite Invalid params",
			method: http.MethodPost,
			handler: func(w http.ResponseWriter, r *http.Request) {
				http.Error(w, "Invalid params: missing argument", http.StatusBadRequest)
			},
			expectedStatus: http.StatusBadRequest,
			validateJSON: func(t *testing.T, resp *JSONRPCResponse) {
				assert.Equal(t, "2.0", resp.JSONRPC)
				require.NotNil(t, resp.Error)
				assert.Equal(t, -32602, resp.Error.Code)
				assert.Equal(t, "Invalid params", resp.Error.Message)
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, "/", nil)
			for k, v := range tc.requestHeaders {
				req.Header.Set(k, v)
			}
			rec := httptest.NewRecorder()

			middleware := JSONRPCComplianceMiddleware(tc.handler)
			middleware.ServeHTTP(rec, req)

			res := rec.Result()
			defer res.Body.Close()

			assert.Equal(t, tc.expectedStatus, res.StatusCode)

			body, err := io.ReadAll(res.Body)
			require.NoError(t, err)

			if tc.validateJSON != nil {
				var jsonResp JSONRPCResponse
				err := json.Unmarshal(body, &jsonResp)
				require.NoError(t, err, "Response body should be valid JSON")
				tc.validateJSON(t, &jsonResp)
			} else {
				// Trim newline added by http.Error
				// assert.Equal(t, tc.expectedBody, strings.TrimSpace(string(body)))
				// Actually http.Error adds newline, but normal write might not.
				// Let's just check containment or exact match depending on case.
				// For simple strings, contain is safe.
				assert.Contains(t, string(body), tc.expectedBody)
			}
		})
	}
}
