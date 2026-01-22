// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package integration

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
)

// MockServerInfo contains info about the mock server.
type MockServerInfo struct {
	URL         string
	CleanupFunc func()
}

// StartMockServer starts a configurable mock server.
// handlers is a map of path to handler function.
func StartMockServer(t *testing.T, handlers map[string]http.HandlerFunc) *MockServerInfo {
	t.Helper()

	mu := sync.Mutex{}
	requestCounts := make(map[string]int)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		requestCounts[r.URL.Path]++
		mu.Unlock()

		if handler, ok := handlers[r.URL.Path]; ok {
			handler(w, r)
			return
		}

		http.NotFound(w, r)
	}))

	t.Logf("Started Mock Server at %s", server.URL)

	return &MockServerInfo{
		URL: server.URL,
		CleanupFunc: func() {
			server.Close()
		},
	}
}

// CatFactsHandler returns a mock handler for the cat facts API.
func CatFactsHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, `{"fact": "Cats are amazing mock animals.", "length": 28}`)
	}
}

// OpenNotifyHandler returns a mock handler for the Open Notify API.
func OpenNotifyHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, `{
			"message": "success",
			"number": 3,
			"people": [
				{"name": "Mock Astronaut 1", "craft": "ISS"},
				{"name": "Mock Astronaut 2", "craft": "ISS"},
				{"name": "Mock Astronaut 3", "craft": "Tiangong"}
			]
		}`)
	}
}
