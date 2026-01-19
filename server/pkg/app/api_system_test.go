// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHandleSystemStatus(t *testing.T) {
	app := NewApplication()
	app.startTime = time.Now().Add(-10 * time.Second)
	app.activeConnections = 5
	app.BoundHTTPPort = 8080
	app.BoundGRPCPort = 9090

	app.SettingsManager = NewGlobalSettingsManager("", nil, nil)

	req, err := http.NewRequest(http.MethodGet, "/api/v1/system/status", nil)
	require.NoError(t, err)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(app.handleSystemStatus)

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var resp map[string]interface{}
	err = json.Unmarshal(rr.Body.Bytes(), &resp)
	require.NoError(t, err)

	assert.GreaterOrEqual(t, resp["uptime_seconds"].(float64), float64(10))
	assert.Equal(t, float64(5), resp["active_connections"])
	assert.Equal(t, float64(8080), resp["bound_http_port"])
	assert.Equal(t, float64(9090), resp["bound_grpc_port"])

	warnings := resp["security_warnings"].([]interface{})
	assert.Contains(t, warnings, "No API Key configured")
}
