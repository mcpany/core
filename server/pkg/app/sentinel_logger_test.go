// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"bytes"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/mcpany/core/server/pkg/logging"
)

func TestSentinel_APIKeyLeak(t *testing.T) {
	// Reset logger to capture output
	logging.ForTestsOnlyResetLogger()

	var buf bytes.Buffer
	// Initialize logger to write to buf
	logging.Init(slog.LevelInfo, &buf)

	app := NewApplication()
	// Use a dummy key
	apiKey := "sentinel-secret-key-12345"
	// Ensure SettingsManager is initialized
	app.SettingsManager = NewGlobalSettingsManager(apiKey, nil, nil)

	// Create middleware
	mw := app.createAuthMiddleware(false, false)

	// Create a dummy handler
	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// Make a request with the API Key
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("X-API-Key", apiKey)

	handler.ServeHTTP(httptest.NewRecorder(), req)

	// Check logs
	logOutput := buf.String()

	// Verification: The log output should NOT contain the secret
	if strings.Contains(logOutput, apiKey) {
		t.Errorf("Security Leak! Log contains API Key: %s", logOutput)
	} else {
        // Just for debug
        // t.Logf("Log output: %s", logOutput)
    }
}
