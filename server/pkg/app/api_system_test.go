// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/mcpany/core/server/pkg/appconsts"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHandleSystemStatus(t *testing.T) {
	app := NewApplication()
	app.startTime = time.Now().Add(-1 * time.Hour) // Started 1 hour ago
	atomic.StoreInt64(&app.activeConnections, 5)
	app.BoundHTTPPort = 8080
	app.BoundGRPCPort = 50051
	// Mock settings manager with no API Key
	app.SettingsManager = NewGlobalSettingsManager("", nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/system/status", nil)
	w := httptest.NewRecorder()

	app.handleSystemStatus(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode)

	var status map[string]any
	err := json.NewDecoder(resp.Body).Decode(&status)
	require.NoError(t, err)

	assert.Equal(t, appconsts.Version, status["version"])
	assert.Equal(t, float64(5), status["active_connections"])
	assert.Equal(t, float64(8080), status["bound_http_port"])
	assert.Equal(t, float64(50051), status["bound_grpc_port"])

	// Uptime should be around 3600 seconds
	uptime := status["uptime_seconds"].(float64)
	assert.True(t, uptime >= 3599 && uptime <= 3601)

	// Check warnings
	warnings := status["security_warnings"].([]any)
	assert.NotEmpty(t, warnings)
	assert.Contains(t, warnings[0], "No API Key configured")
}

func TestHandleSystemStatus_MethodNotAllowed(t *testing.T) {
	app := NewApplication()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/system/status", nil)
	w := httptest.NewRecorder()

	app.handleSystemStatus(w, req)
	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
}
