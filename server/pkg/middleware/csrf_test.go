package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCSRFMiddleware(t *testing.T) {
	allowedOrigins := []string{"http://allowed.com", "http://localhost:3000"}
	m := NewCSRFMiddleware(allowedOrigins)
	handler := m.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	tests := []struct {
		name           string
		method         string
		headers        map[string]string
		expectedStatus int
	}{
		{
			name:           "Safe Method GET",
			method:         http.MethodGet,
			headers:        map[string]string{},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Safe Method OPTIONS",
			method:         http.MethodOptions,
			headers:        map[string]string{},
			expectedStatus: http.StatusOK,
		},
		{
			name:   "POST with X-API-Key",
			method: http.MethodPost,
			headers: map[string]string{
				"X-API-Key": "some-key",
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:   "POST with Authorization Bearer",
			method: http.MethodPost,
			headers: map[string]string{
				"Authorization": "Bearer token",
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:   "POST with JSON Content-Type",
			method: http.MethodPost,
			headers: map[string]string{
				"Content-Type": "application/json",
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:   "POST with Allowed Origin",
			method: http.MethodPost,
			headers: map[string]string{
				"Origin": "http://allowed.com",
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:   "POST with Allowed Referer",
			method: http.MethodPost,
			headers: map[string]string{
				"Referer": "http://allowed.com/some/page",
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:   "POST with Blocked Origin",
			method: http.MethodPost,
			headers: map[string]string{
				"Origin": "http://attacker.com",
			},
			expectedStatus: http.StatusForbidden,
		},
		{
			name:   "POST with Blocked Referer",
			method: http.MethodPost,
			headers: map[string]string{
				"Referer": "http://attacker.com/page",
			},
			expectedStatus: http.StatusForbidden,
		},
		{
			name:   "POST Form without Origin/Referer (allowed for CLI)",
			method: http.MethodPost,
			headers: map[string]string{
				"Content-Type": "application/x-www-form-urlencoded",
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:   "POST Form with Blocked Origin",
			method: http.MethodPost,
			headers: map[string]string{
				"Content-Type": "application/x-www-form-urlencoded",
				"Origin":       "http://attacker.com",
			},
			expectedStatus: http.StatusForbidden,
		},
		{
			name:   "POST Form with Same Origin (Host match)",
			method: http.MethodPost,
			headers: map[string]string{
				"Content-Type": "application/x-www-form-urlencoded",
				"Origin":       "http://same-server.com",
				"Host":         "same-server.com",
			},
			expectedStatus: http.StatusOK,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, "/", nil)
			// httptest.NewRequest sets Host to "example.com" by default.
			// We need to override it if specified in headers for our test logic.
			if host, ok := tc.headers["Host"]; ok {
				req.Host = host
			}
			for k, v := range tc.headers {
				if k != "Host" { // Host header is special in Go http.Request
					req.Header.Set(k, v)
				}
			}
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req)
			assert.Equal(t, tc.expectedStatus, w.Code, "Test case: %s", tc.name)
		})
	}
}

func TestCSRFMiddleware_EmptyConfig(t *testing.T) {
	// Initialize with empty allowed origins
	m := NewCSRFMiddleware([]string{})
	handler := m.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	tests := []struct {
		name           string
		headers        map[string]string
		expectedStatus int
	}{
		{
			name: "Localhost Origin Allowed by Default",
			headers: map[string]string{
				"Origin": "http://localhost:3000",
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "127.0.0.1 Origin Allowed by Default",
			headers: map[string]string{
				"Origin": "http://127.0.0.1:4000",
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "External Origin Blocked",
			headers: map[string]string{
				"Origin": "http://external.com",
			},
			expectedStatus: http.StatusForbidden,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/", nil)
			for k, v := range tc.headers {
				req.Header.Set(k, v)
			}
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req)
			assert.Equal(t, tc.expectedStatus, w.Code, "Test case: %s", tc.name)
		})
	}
}
