// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDebuggerLatestEntries(t *testing.T) {
	debugger := NewDebugger(5)
	handler := debugger.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// Populate with 10 requests (overflows 5)
	for i := 0; i < 10; i++ {
		w := httptest.NewRecorder()
		path := fmt.Sprintf("/req-%d", i)
		req, _ := http.NewRequest("GET", path, nil)
		handler.ServeHTTP(w, req)
	}

	// Case 1: Fetch all (default)
	entries := debugger.Entries()
	assert.Len(t, entries, 5)
	// Entries returns Oldest -> Newest
	// Ring buffer (size 5) should contain: /req-5, /req-6, /req-7, /req-8, /req-9
	assert.Equal(t, "/req-5", entries[0].Path)
	assert.Equal(t, "/req-9", entries[4].Path)

	// Case 2: LatestEntries(3)
	latest := debugger.LatestEntries(3)
	assert.Len(t, latest, 3)
	// LatestEntries returns Newest -> Oldest
	// Should contain: /req-9, /req-8, /req-7
	assert.Equal(t, "/req-9", latest[0].Path)
	assert.Equal(t, "/req-8", latest[1].Path)
	assert.Equal(t, "/req-7", latest[2].Path)

	// Case 3: LatestEntries(10) (more than available)
	latestFull := debugger.LatestEntries(10)
	assert.Len(t, latestFull, 5) // Capped at size
	assert.Equal(t, "/req-9", latestFull[0].Path)
	assert.Equal(t, "/req-5", latestFull[4].Path)
}

func TestDebuggerAPIHandlerLimit(t *testing.T) {
	debugger := NewDebugger(5)
	handler := debugger.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// Populate
	for i := 0; i < 5; i++ {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", fmt.Sprintf("/req-%d", i), nil)
		handler.ServeHTTP(w, req)
	}

	// API Request with limit=2
	apiHandler := debugger.APIHandler()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/debug?limit=2", nil)
	apiHandler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var entries []DebugEntry
	err := json.NewDecoder(w.Body).Decode(&entries)
	assert.NoError(t, err)
	assert.Len(t, entries, 2)
	assert.Equal(t, "/req-4", entries[0].Path)
}
