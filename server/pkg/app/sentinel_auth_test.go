// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSentinelAuth_InsecureMode_PrivateNetwork(t *testing.T) {
	// Setup Application with no API Key
	app := NewApplication()
	// Need to initialize SettingsManager because createAuthMiddleware accesses it
	app.SettingsManager = NewGlobalSettingsManager("", nil, nil) // APIKey empty

	// Create Auth Middleware (forcePrivateIPOnly=false, trustProxy=false)
	middleware := app.createAuthMiddleware(false, false)

	// Create a dummy handler
	dummyHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	protectedHandler := middleware(dummyHandler)

	// Test Case 1: Loopback (Should be Allowed)
	req1 := httptest.NewRequest("GET", "/", nil)
	req1.RemoteAddr = "127.0.0.1:12345"
	rec1 := httptest.NewRecorder()
	protectedHandler.ServeHTTP(rec1, req1)
	assert.Equal(t, http.StatusOK, rec1.Code, "Loopback should be allowed in insecure mode")

	// Test Case 2: Private Network IP (Should be Blocked)
	// Currently, this passes (allowed), demonstrating the vulnerability.
	// We assert Forbidden to fail initially.
	req2 := httptest.NewRequest("GET", "/", nil)
	req2.RemoteAddr = "192.168.1.100:12345"
	rec2 := httptest.NewRecorder()
	protectedHandler.ServeHTTP(rec2, req2)

	// Expect failure until fixed
	assert.Equal(t, http.StatusForbidden, rec2.Code, "Private Network IP should be blocked in insecure mode")
}
