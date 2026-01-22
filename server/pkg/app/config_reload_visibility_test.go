// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"testing"
	"time"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestConfigReloadVisibility(t *testing.T) {
	fs := afero.NewMemMapFs()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 1. Start with valid config
	configContent := `
global_settings:
  log_level: INFO
upstream_services: []
`
	err := afero.WriteFile(fs, "/config.yaml", []byte(configContent), 0o644)
	require.NoError(t, err)

	app := NewApplication()
	mockStore := new(MockStore)
	mockStore.On("Load", mock.Anything).Return(&configv1.McpAnyServerConfig{
		GlobalSettings: &configv1.GlobalSettings{},
	}, nil)
	mockStore.On("ListServices", mock.Anything).Return([]*configv1.UpstreamServiceConfig{}, nil)
	mockStore.On("GetGlobalSettings", mock.Anything).Return(&configv1.GlobalSettings{}, nil)
	mockStore.On("Close").Return(nil)
	app.Storage = mockStore

	errChan := make(chan error, 1)
	go func() {
		errChan <- app.Run(RunOptions{
			Ctx:             ctx,
			Fs:              fs,
			Stdio:           false,
			JSONRPCPort:     "127.0.0.1:0",
			GRPCPort:        "127.0.0.1:0",
			ConfigPaths:     []string{"/config.yaml"},
			APIKey:          "",
			ShutdownTimeout: 5 * time.Second,
		})
	}()

	require.NoError(t, app.WaitForStartup(ctx))
	baseURL := "http://127.0.0.1:" + strconv.Itoa(int(app.BoundHTTPPort.Load()))

	// 2. Check initial system status
	resp, err := http.Get(baseURL + "/api/v1/system/status")
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var statusMap map[string]any
	json.NewDecoder(resp.Body).Decode(&statusMap)

	// These fields should now be present (even if empty error)
	status, hasConfigStatus := statusMap["config_status"]
	assert.True(t, hasConfigStatus, "config_status should be present")
	assert.Equal(t, "ok", status)

	// 3. Trigger a bad reload
	// We overwrite the file with bad content
	err = afero.WriteFile(fs, "/config.yaml", []byte("malformed: : yaml"), 0o644)
	require.NoError(t, err)

	// Manually trigger reload since we don't want to wait for watcher
	err = app.ReloadConfig(ctx, fs, []string{"/config.yaml"})
	require.Error(t, err)

	// 4. Check system status again
	resp2, err := http.Get(baseURL + "/api/v1/system/status")
	require.NoError(t, err)
	defer resp2.Body.Close()

	var statusMap2 map[string]any
	json.NewDecoder(resp2.Body).Decode(&statusMap2)

	status2, hasStatus2 := statusMap2["config_status"]
	assert.True(t, hasStatus2)
	assert.Equal(t, "degraded", status2)

	reloadErr, hasError := statusMap2["last_reload_error"]
	assert.True(t, hasError, "last_reload_error should be present")
	assert.NotEmpty(t, reloadErr)
	assert.Contains(t, reloadErr, "yaml")

	// 5. Check Health endpoint
	resp3, err := http.Get(baseURL + "/api/v1/health")
	require.NoError(t, err)
	defer resp3.Body.Close()
	// It returns 200 OK currently
	assert.Equal(t, http.StatusOK, resp3.StatusCode)

	cancel()
	<-errChan
}
