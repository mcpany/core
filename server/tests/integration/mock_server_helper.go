// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package integration

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

// StartMockServer serves as a public interface for interacting with StartMockServer.
//
// Summary: Start the mock server appropriately based on current system conditions.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the successfully computed domain model or execution state.
//
// Errors:
//   - No explicit errors are thrown by this operation.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
func StartMockServer(t *testing.T, handler http.Handler) *httptest.Server {
	t.Helper()
	server := httptest.NewServer(handler)
	t.Logf("Started mock server at %s", server.URL)
	return server
}

// DefaultMockHandler serves as a public interface for interacting with DefaultMockHandler.
//
// Summary: Default the mock handler appropriately based on current system conditions.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the successfully computed domain model or execution state.
//
// Errors:
//   - No explicit errors are thrown by this operation.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
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

// CreateMockServerWithResponses serves as a public interface for interacting with CreateMockServerWithResponses.
//
// Summary: Create the mock server with responses appropriately based on current system conditions.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the successfully computed domain model or execution state.
//
// Errors:
//   - No explicit errors are thrown by this operation.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
func CreateMockServerWithResponses(t *testing.T, responses map[string]string) *httptest.Server {
	return StartMockServer(t, DefaultMockHandler(t, responses))
}
