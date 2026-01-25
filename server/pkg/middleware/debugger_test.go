// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDebuggerMiddleware(t *testing.T) {
	debugger := NewDebugger(10)

	handler := debugger.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// Make a request
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Check entries
	entries := debugger.Entries()
	assert.Len(t, entries, 1)
	assert.Equal(t, "/test", entries[0].Path)
	assert.Equal(t, http.StatusOK, entries[0].Status)
	assert.NotEmpty(t, entries[0].ID)

	// Test Ring Buffer Overflow
	for i := 0; i < 15; i++ {
		w = httptest.NewRecorder()
		req, _ = http.NewRequest("GET", "/test", nil)
		handler.ServeHTTP(w, req)
	}

	entries = debugger.Entries()
	assert.Len(t, entries, 10) // Should only keep last 10
}

func TestDebuggerBodyCapture(t *testing.T) {
	debugger := NewDebugger(10)

	handler := debugger.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(body)
	}))

	payload := map[string]string{"foo": "bar"}
	payloadBytes, _ := json.Marshal(payload)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/echo", bytes.NewBuffer(payloadBytes))
	req.Header.Set("Content-Type", "application/json")
	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, string(payloadBytes), w.Body.String())

	entries := debugger.Entries()
	assert.Len(t, entries, 1)
	assert.JSONEq(t, string(payloadBytes), entries[0].RequestBody)
	assert.JSONEq(t, string(payloadBytes), entries[0].ResponseBody)
}

func TestDebuggerLargeBodyTruncation(t *testing.T) {
	debugger := NewDebugger(10)
	debugger.maxBodySize = 10 // Very small limit for testing

	handler := debugger.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Handler should still receive full body
		body, _ := io.ReadAll(r.Body)
		w.WriteHeader(http.StatusOK)
		w.Write(body)
	}))

	longString := "This is a very long string that should be truncated"
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/echo", bytes.NewBufferString(longString))
	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, longString, w.Body.String()) // Handler got full body

	entries := debugger.Entries()
	assert.Len(t, entries, 1)
	assert.Contains(t, entries[0].RequestBody, "This is a ")
	assert.Contains(t, entries[0].RequestBody, "... [truncated]")
	assert.Contains(t, entries[0].ResponseBody, "This is a ")
	assert.Contains(t, entries[0].ResponseBody, "... [truncated]")
}

func TestDebuggerLatestEntries(t *testing.T) {
	debugger := NewDebugger(10)
	handler := debugger.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// Generate 15 requests
	for i := 0; i < 15; i++ {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test/"+strconv.Itoa(i), nil)
		handler.ServeHTTP(w, req)
	}

	// LatestEntries(5) should return 14, 13, 12, 11, 10
	latest := debugger.LatestEntries(5)
	assert.Len(t, latest, 5)
	assert.Equal(t, "/test/14", latest[0].Path)
	assert.Equal(t, "/test/13", latest[1].Path)

	// API Handler Limit
	apiHandler := debugger.APIHandler()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api?limit=3", nil)
	apiHandler.ServeHTTP(w, req)

	var apiEntries []DebugEntry
	_ = json.Unmarshal(w.Body.Bytes(), &apiEntries)
	assert.Len(t, apiEntries, 3)
	assert.Equal(t, "/test/14", apiEntries[0].Path)
}
