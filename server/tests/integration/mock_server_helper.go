// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package integration

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

// StartMockServer starts a new mock server with the provided handler.
//
// Summary: Starts a new mock server with the provided handler.
//
// Parameters:
//   - t (*testing.T): Parameter.
//   - handler (http.Handler): Parameter.
//
// Returns:
//   - *httptest.Server: Return value.
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

// DefaultMockHandler provides a simple way to define responses for specific paths.
//
// Summary: Provides a simple way to define responses for specific paths.
//
// Parameters:
//   - t (*testing.T): Parameter.
//   - responses (map[string]string): Parameter.
//
// Returns:
//   - http.Handler: Return value.
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

// CreateMockServerWithResponses is a convenience function to start a server with static responses.
//
// Summary: Is a convenience function to start a server with static responses.
//
// Parameters:
//   - t (*testing.T): Parameter.
//   - responses (map[string]string): Parameter.
//
// Returns:
//   - *httptest.Server: Return value.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.

func CreateMockServerWithResponses(t *testing.T, responses map[string]string) *httptest.Server {
	return StartMockServer(t, DefaultMockHandler(t, responses))
}
