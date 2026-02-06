// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestHandleSystemStatus(t *testing.T) {
	// Setup Application state
	app := &Application{
		startTime:         time.Now().Add(-1 * time.Hour), // Started 1 hour ago
		activeConnections: 5,
	}
	app.BoundHTTPPort.Store(8080)
	app.BoundGRPCPort.Store(50051)
	app.SettingsManager = NewGlobalSettingsManager("test-api-key", nil, nil)

	// Create request
	req := httptest.NewRequest(http.MethodGet, "/system/status", nil)
	w := httptest.NewRecorder()

	// Call handler directly (it's a method on Application, accessing private fields)
	app.handleSystemStatus(w, req)

	// Verify response code
	assert.Equal(t, http.StatusOK, w.Code)

	// Verify response body
	var resp SystemStatusResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)

	assert.Equal(t, int64(3600), resp.UptimeSeconds, "Uptime should be approx 3600 seconds")
	assert.Equal(t, int32(5), resp.ActiveConnections)
	assert.Equal(t, 8080, resp.BoundHTTPPort)
	assert.Equal(t, 50051, resp.BoundGRPCPort)
	assert.NotEmpty(t, resp.Version)
	assert.Empty(t, resp.SecurityWarnings, "Should have no security warnings with API key")
}

func TestHandleSystemStatus_SecurityWarning(t *testing.T) {
	// Setup Application state without API Key
	app := &Application{
		startTime: time.Now(),
	}
	app.SettingsManager = NewGlobalSettingsManager("", nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/system/status", nil)
	w := httptest.NewRecorder()

	app.handleSystemStatus(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp SystemStatusResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)

	assert.Contains(t, resp.SecurityWarnings, "No API Key configured")
}
