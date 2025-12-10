// Copyright (C) 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package e2e

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/mcpany/core/pkg/app"
	"github.com/mcpany/core/pkg/testutil"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestE2E_HealthCheck(t *testing.T) {
	// --- Test Setup ---
	healthy := true
	healthMux := http.NewServeMux()
	healthMux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		if healthy {
			w.WriteHeader(http.StatusOK)
			fmt.Fprint(w, "OK")
		} else {
			w.WriteHeader(http.StatusServiceUnavailable)
			fmt.Fprint(w, "Service Unavailable")
		}
	})
	healthMux.HandleFunc("/echo", func(w http.ResponseWriter, r *http.Request) {
		var body map[string]interface{}
		json.NewDecoder(r.Body).Decode(&body)
		json.NewEncoder(w).Encode(body)
	})

	upstreamServer := httptest.NewServer(healthMux)
	defer upstreamServer.Close()

	// --- MCP Any Server Setup ---
	configContent := fmt.Sprintf(`
upstreamServices:
  - name: "healthy-http-service"
    httpService:
      address: "%s"
      healthCheck:
        url: "%s/healthz"
        interval: "100ms"
        timeout: "50ms"
      calls:
        - operationId: "echo"
          endpointPath: "/echo"
          method: "HTTP_METHOD_POST"
`, upstreamServer.URL, upstreamServer.URL)

	fs := afero.NewMemMapFs()
	configDir := "/etc/mcpany"
	require.NoError(t, fs.MkdirAll(configDir, 0755))
	configPath := filepath.Join(configDir, "config.yaml")
	require.NoError(t, afero.WriteFile(fs, configPath, []byte(configContent), 0644))

	mcpApp := app.NewApplication()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Use random, available ports for the test server
	mcpAddr, err := testutil.GetFreePort()
	require.NoError(t, err)

	grpcAddr, err := testutil.GetFreePort()
	require.NoError(t, err)

	go func() {
		err := mcpApp.Run(ctx, fs, false, mcpAddr, grpcAddr, []string{configPath}, 5*time.Second)
		if err != nil && err != context.Canceled {
			t.Logf("MCP Any server exited with error: %v", err)
		}
	}()

	// Wait for the server to be ready
	require.NoError(t, testutil.WaitForServerReady(ctx, fmt.Sprintf("localhost%s", mcpAddr), 5*time.Second))

	// --- Test Execution ---
	mcpClient := &http.Client{}
	toolCallURL := fmt.Sprintf("http://localhost%s", mcpAddr)

	// 1. Initial state: service is healthy, tool call should succeed
	t.Run("ServiceHealthy_ToolCallSucceeds", func(t *testing.T) {
		body := `{"jsonrpc": "2.0", "method": "tools/call", "params": {"name": "healthy-http-service/-/echo", "arguments": {"message": "hello"}}, "id": 1}`
		resp, err := mcpClient.Post(toolCallURL, "application/json", strings.NewReader(body))
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)
		assert.NotContains(t, result, "error", "Tool call should not return an error when service is healthy")
	})

	// 2. Transition to unhealthy state
	t.Run("ServiceUnhealthy_ToolCallFails", func(t *testing.T) {
		healthy = false
		time.Sleep(250 * time.Millisecond) // Wait for the health check to fail

		body := `{"jsonrpc": "2.0", "method": "tools/call", "params": {"name": "healthy-http-service/-/echo", "arguments": {"message": "hello"}}, "id": 2}`
		resp, err := mcpClient.Post(toolCallURL, "application/json", strings.NewReader(body))
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode) // JSON-RPC itself returns 200

		var result map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&result)
		require.NoError(t, err)

		require.Contains(t, result, "error", "Tool call should return an error when service is unhealthy")
		errMap := result["error"].(map[string]interface{})
		assert.Contains(t, errMap["message"], "service is currently unavailable")
	})

	// 3. Transition back to healthy state
	t.Run("ServiceRecovers_ToolCallSucceeds", func(t *testing.T) {
		healthy = true
		time.Sleep(250 * time.Millisecond) // Wait for the health check to pass again

		body := `{"jsonrpc": "2.0", "method": "tools/call", "params": {"name": "healthy-http-service/-/echo", "arguments": {"message": "hello"}}, "id": 3}`
		resp, err := mcpClient.Post(toolCallURL, "application/json", strings.NewReader(body))
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)
		assert.NotContains(t, result, "error", "Tool call should succeed again after service recovers")
	})
}
