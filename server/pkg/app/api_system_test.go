// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSystemStatus(t *testing.T) {
	app := NewApplication()
	app.SettingsManager = NewGlobalSettingsManager("test-key", nil, nil)

	// Mock configFileIgnored
	app.configFileIgnored = true

	req, err := http.NewRequest(http.MethodGet, "/api/v1/system/status", nil)
	require.NoError(t, err)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(app.handleSystemStatus)

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var resp SystemStatusResponse
	err = json.Unmarshal(rr.Body.Bytes(), &resp)
	require.NoError(t, err)

	assert.Contains(t, resp.SecurityWarnings, "Config files provided but ignored because MCPANY_ENABLE_FILE_CONFIG is not true.")
}
