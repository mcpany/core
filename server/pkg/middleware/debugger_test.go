// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDebuggerMiddleware(t *testing.T) {
	debugger := NewDebugger(10)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	wrapped := debugger.Middleware(handler)

	// Make a request
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	wrapped.ServeHTTP(w, req)

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
		wrapped.ServeHTTP(w, req)
	}

	entries = debugger.Entries()
	assert.Len(t, entries, 10) // Should only keep last 10
}

func TestDebuggerBodyCapture(t *testing.T) {
	debugger := NewDebugger(10)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(body)
	})

	wrapped := debugger.Middleware(handler)

	payload := map[string]string{"foo": "bar"}
	payloadBytes, _ := json.Marshal(payload)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/echo", bytes.NewBuffer(payloadBytes))
	req.Header.Set("Content-Type", "application/json")
	wrapped.ServeHTTP(w, req)

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

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Handler should still receive full body
		body, _ := io.ReadAll(r.Body)
		w.WriteHeader(http.StatusOK)
		w.Write(body)
	})

	wrapped := debugger.Middleware(handler)

	longString := "This is a very long string that should be truncated"
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/echo", bytes.NewBufferString(longString))
	wrapped.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, longString, w.Body.String()) // Handler got full body

	entries := debugger.Entries()
	assert.Len(t, entries, 1)
	assert.Contains(t, entries[0].RequestBody, "This is a ")
	assert.Contains(t, entries[0].RequestBody, "... [truncated]")
	assert.Contains(t, entries[0].ResponseBody, "This is a ")
	assert.Contains(t, entries[0].ResponseBody, "... [truncated]")
}
