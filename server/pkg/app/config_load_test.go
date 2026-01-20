// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"context"
	"os"
	"testing"
	"time"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestRun_ConfigPath_WithoutEnvVar(t *testing.T) {
	// Setup
	fs := afero.NewMemMapFs()
	// Create a config file with a typo
	configContent := `
upstream_services:
  - name: "weather-api"
    http_service:
      adress: "https://api.weather.com"
`
	_ = afero.WriteFile(fs, "config.yaml", []byte(configContent), 0644)

	// Ensure env var is NOT set
	os.Unsetenv("MCPANY_ENABLE_FILE_CONFIG")

	app := NewApplication()
	// Mock storage to avoid DB init errors
	mockStore := new(MockStore)
	// We need to return empty config from DB so it relies on file (if loaded)
	mockStore.On("ListServices", mock.Anything).Return(([]*configv1.UpstreamServiceConfig)(nil), nil)
	mockStore.On("GetGlobalSettings", mock.Anything).Return((*configv1.GlobalSettings)(nil), nil)
	// Save calls during init
	mockStore.On("SaveGlobalSettings", mock.Anything, mock.Anything).Return(nil)
	mockStore.On("SaveService", mock.Anything, mock.Anything).Return(nil)
	// Load call
	mockStore.On("Load", mock.Anything).Return(&configv1.McpAnyServerConfig{}, nil)
	mockStore.On("HasConfigSources").Return(false)

	app.Storage = mockStore

	// Run
	// We expect it to FAIL because of the typo in config.yaml.
	// If it succeeds (returns nil or timeout error later), it means it IGNORED the config.
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	err := app.Run(ctx, fs, false, "8080", "50051", []string{"config.yaml"}, "", 1*time.Second)

	// In the buggy state, err will be nil (started successfully with empty config)
	// or "context deadline exceeded" if it hung waiting for something.
	// We want to assert that it FAILS with validation error.
	if err == nil {
		t.Log("Bug detected: Server started successfully ignoring invalid config file")
		// Fail the test to prove the bug exists.
		// Wait, the prompt says: "Write a new test case that specifically fails before your fix and passes after it"
		// So if I assert "Error Should Be Validation Error", it will fail now.
		t.Fail()
		return
	}

	if err.Error() == "context deadline exceeded" {
		t.Log("Bug detected: Server started successfully (timeout) ignoring invalid config file")
		t.Fail()
		return
	}

	assert.Contains(t, err.Error(), "unknown field \"adress\"", "Should fail with validation error")
}
