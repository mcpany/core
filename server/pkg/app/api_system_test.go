// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mcpany/core/server/pkg/appconsts"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSystemStatus(t *testing.T) {
	app := NewApplication()
	// Mock SettingsManager to avoid nil pointer dereference
	app.SettingsManager = NewGlobalSettingsManager("", nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/system/status", nil)
	w := httptest.NewRecorder()

	// Call the handler directly
	app.handleSystemStatus(w, req)

	resp := w.Result()
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var status SystemStatusResponse
	err := json.NewDecoder(resp.Body).Decode(&status)
	require.NoError(t, err)

	assert.Equal(t, appconsts.Version, status.Version)
	assert.GreaterOrEqual(t, status.UptimeSeconds, int64(0))
	assert.Equal(t, int32(0), status.ActiveConnections)
	// Since we didn't start servers, ports are 0
	assert.Equal(t, 0, status.BoundHTTPPort)
	assert.Equal(t, 0, status.BoundGRPCPort)

	// API Key is empty, so we expect a warning
	assert.Contains(t, status.SecurityWarnings, "No API Key configured")
}
