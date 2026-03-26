// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package integration

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

// StartMockServer startMockServer start mock server.
//
// Summary: StartMockServer start mock server.
//
// Parameters:
//   - t (*testing.T): The t.
//   - handler (http.Handler): The handler.
//
// Returns:
//   - *httptest.Server: The result.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func StartMockServer(t *testing.T, handler http.Handler) *httptest.Server {
	t.Helper()
	server := httptest.NewServer(handler)
	t.Logf("Started mock server at %s", server.URL)
	return server
}

// DefaultMockHandler defaultMockHandler default mock handler.
//
// Summary: DefaultMockHandler default mock handler.
//
// Parameters:
//   - t (*testing.T): The t.
//   - responses (map[string]string): The responses.
//
// Returns:
//   - http.Handler: The result.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func DefaultMockHandler(t *testing.T, responses map[string]string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		bodyBytes, _ := io.ReadAll(r.Body)
		t.Logf("Mock server received request: %s %s Body: %s", r.Method, r.URL.RequestURI(), string(bodyBytes))

		// Check exact path match including query string
		if body, ok := responses[r.URL.RequestURI()]; ok {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(body))
			return
		}

		// Fallback to check path only match
		if body, ok := responses[r.URL.Path]; ok {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(body))
			return
		}

		t.Logf("Mock server: no response found for %s", r.URL.RequestURI())
		http.NotFound(w, r)
	})
}

// CreateMockServerWithResponses persists the mock server with responses.
//
// Summary: Persists the mock server with responses.
//
// Parameters:
//   - t (*testing.T): The t.
//   - responses (map[string]string): The responses.
//
// Returns:
//   - *httptest.Server: The result.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func CreateMockServerWithResponses(t *testing.T, responses map[string]string) *httptest.Server {
	return StartMockServer(t, DefaultMockHandler(t, responses))
}
