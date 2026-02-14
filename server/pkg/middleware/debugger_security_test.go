// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDebuggerRedactsSensitiveHeaders(t *testing.T) {
	debugger := NewDebugger(10)
	defer debugger.Close()

	handler := debugger.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Set-Cookie", "session=secret_session_id; Path=/; Secure; HttpOnly")
		w.WriteHeader(http.StatusOK)
	}))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/secure-endpoint", nil)

	// Add sensitive headers
	sensitiveHeaders := map[string]string{
		"Authorization": "Bearer sensitive_token_123",
		"Cookie":        "session=user_session_abc",
		"X-API-Key":     "mcp_sk_live_12345",
	}

	for k, v := range sensitiveHeaders {
		req.Header.Set(k, v)
	}

	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Check entries
	entries := waitForEntries(t, debugger, 1)
	assert.Len(t, entries, 1)
	entry := entries[0]

	// Verify request headers are redacted
	// Note: Initially, this test is expected to FAIL or PASS depending on whether I assert they ARE redacted or NOT.
	// To confirm the vulnerability, I should assert they are present.
	// But the plan says "asserts that the captured entry contains the raw sensitive values (this confirms the vulnerability)".
	// However, I will write the test to assert redaction, and expect it to FAIL first.
	// This is standard TDD.

	assert.Equal(t, "[REDACTED]", entry.RequestHeaders.Get("Authorization"), "Authorization header should be redacted")
	assert.Equal(t, "[REDACTED]", entry.RequestHeaders.Get("Cookie"), "Cookie header should be redacted")
	assert.Equal(t, "[REDACTED]", entry.RequestHeaders.Get("X-API-Key"), "X-API-Key header should be redacted")

	// Verify response headers are redacted
	assert.Equal(t, "[REDACTED]", entry.ResponseHeaders.Get("Set-Cookie"), "Set-Cookie header should be redacted")
}
