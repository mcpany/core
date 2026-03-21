// Copyright 2025 Author(s) of MCP Any
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

func TestHandleSystemStatus_SecurityWarnings(t *testing.T) {
	t.Run("Warning when API Key missing", func(t *testing.T) {
		app := NewApplication()
		app.SettingsManager = NewGlobalSettingsManager("", nil, nil) // Empty API Key

		req := httptest.NewRequest(http.MethodGet, "/system/status", nil)
		w := httptest.NewRecorder()

		handler := http.HandlerFunc(app.handleSystemStatus)
		handler.ServeHTTP(w, req)

		resp := w.Result()
		defer resp.Body.Close()

		var status SystemStatusResponse
		err := json.NewDecoder(resp.Body).Decode(&status)
		require.NoError(t, err)

		require.NotEmpty(t, status.SecurityWarnings)
		assert.Contains(t, status.SecurityWarnings, "No API Key configured")
	})

	t.Run("No Warning when API Key present", func(t *testing.T) {
		app := NewApplication()
		app.SettingsManager = NewGlobalSettingsManager("secret", nil, nil)

		req := httptest.NewRequest(http.MethodGet, "/system/status", nil)
		w := httptest.NewRecorder()

		handler := http.HandlerFunc(app.handleSystemStatus)
		handler.ServeHTTP(w, req)

		resp := w.Result()
		defer resp.Body.Close()

		var status SystemStatusResponse
		err := json.NewDecoder(resp.Body).Decode(&status)
		require.NoError(t, err)

		assert.Empty(t, status.SecurityWarnings)
	})
}
