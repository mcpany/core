//go:build e2e

// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package e2e_sequential

import (
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// verifyEndpointLifecycle checks if a URL returns the expected status code within a timeout
func verifyEndpointLifecycle(t *testing.T, url string, expectedStatus int, timeout time.Duration) {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		resp, err := http.Get(url)
		if err == nil {
			defer resp.Body.Close()
			if resp.StatusCode == expectedStatus {
				return
			}
		}
		time.Sleep(500 * time.Millisecond)
	}
	t.Fatalf("Endpoint %s did not return %d within %v", url, expectedStatus, timeout)
}

// TestCUJ_Lifecycle_And_Config tests lifecycle events and config changes.
// Using Filesystem upstream to avoid dependency on external binaries or containers.
func TestCUJ_Lifecycle_And_Config(t *testing.T) {
	// Enable running local if Docker is not available
	// useLocal := os.Getenv("E2E_DOCKER") != "true"

	rootDir, err := os.Getwd()
	require.NoError(t, err)
	// Adjust rootDir resolution based on where the test is run from
	if strings.Contains(rootDir, "tests/e2e_sequential") {
		rootDir = filepath.Join(rootDir, "../../..")
	} else if strings.Contains(rootDir, "server") {
		rootDir = filepath.Join(rootDir, "..")
	}
	rootDir, err = filepath.Abs(rootDir)
	require.NoError(t, err)

	configDir := filepath.Join(rootDir, "build", "e2e_config_lifecycle")
	err = os.MkdirAll(configDir, 0755)
	require.NoError(t, err)
	defer os.RemoveAll(configDir)

	// Create a dummy file to read/list
	dummyFile := filepath.Join(configDir, "hello.txt")
	err = os.WriteFile(dummyFile, []byte("world"), 0644)
	require.NoError(t, err)

	// In local mode, paths must be absolute on the host
	// dataPath := "/config_data"
	// if useLocal {
	// 	dataPath = configDir
	// }

	// Use ports that are likely free
	port1 := "50055"
	port2 := "50056"
	port3 := "50057"

	// Initial Config: Enable Filesystem Upstream
	// Note: We use the mock_server binary instead of a real filesystem server if possible
	// But let's assume we want to test the CONFIG lifecycle, not the tool execution specifically.
	// However, to "ListTools", we need a valid upstream.
	// If "filesystem_service" relies on "npx", it might fail in restricted envs.
	// Let's use "command_line_service" with "ls" or similar? No, MCP protocol.
	// Let's use the `mock_server` binary we built!
	mockServerBin := filepath.Join(rootDir, "build/bin/mock_server")
	if _, err := os.Stat(mockServerBin); os.IsNotExist(err) {
		t.Skip("Mock server binary not found at " + mockServerBin + ". Skipping test.")
	}

	config1 := fmt.Sprintf(`
global_settings:
  mcp_listen_address: "127.0.0.1:%s"
  profile_definitions:
    - name: "default"
      selector:
        tags: ["default"]
      service_config:
        "fs-service":
          enabled: true
upstream_services:
  - id: "fs-service"
    name: "Filesystem Service"
    disable: false
    auto_discover_tool: true
    mcp_service:
      stdio_connection:
        command: "%s"
        args: []
`, port1, mockServerBin)

	configPath := filepath.Join(configDir, "config.yaml")
	err = os.WriteFile(configPath, []byte(config1), 0644)
	require.NoError(t, err)

	serverBin := filepath.Join(rootDir, "build/bin/server")
	// Ensure binary exists or build it
	if _, err := os.Stat(serverBin); os.IsNotExist(err) {
		t.Log("Server binary not found, attempting to build...")
		buildCmd := exec.Command("make", "build")
		buildCmd.Dir = rootDir
		out, err := buildCmd.CombinedOutput()
		require.NoError(t, err, "Failed to build server: %s", string(out))
	}

	cmd := exec.Command(serverBin, "run", "--config-path", configPath, "--debug", "--api-key", "test-key")
	// Redirect output for debugging
	logFile, err := os.Create(filepath.Join(configDir, "server.log"))
	require.NoError(t, err)
	cmd.Stdout = logFile
	cmd.Stderr = logFile
	err = cmd.Start()
	require.NoError(t, err)

	baseURL := fmt.Sprintf("http://127.0.0.1:%s", port1)

	defer func() {
		if cmd != nil && cmd.Process != nil {
			cmd.Process.Kill()
		}
	}()

	// CUJ 1: Health
	verifyEndpointLifecycle(t, fmt.Sprintf("%s/healthz", baseURL), 200, 60*time.Second)

	// Wait for tools to be available via API first to be safe
	verifyEndpointLifecycle(t, fmt.Sprintf("%s/api/v1/tools", baseURL), 200, 30*time.Second)

	// CUJ 2: Restart with new config
	// Create Config 2 (Add another service)
	config2 := fmt.Sprintf(`
global_settings:
  mcp_listen_address: "127.0.0.1:%s"
  profile_definitions:
    - name: "default"
      selector:
        tags: ["default"]
      service_config:
        "fs-service":
          enabled: true
        "second-service":
          enabled: true
upstream_services:
  - id: "fs-service"
    name: "Filesystem Service"
    disable: false
    auto_discover_tool: true
    mcp_service:
      stdio_connection:
        command: "%s"
        args: []
  - id: "second-service"
    name: "Second Service"
    disable: false
    auto_discover_tool: true
    mcp_service:
      stdio_connection:
        command: "%s"
        args: []
`, port2, mockServerBin, mockServerBin)

	err = os.WriteFile(configPath, []byte(config2), 0644)
	require.NoError(t, err)

	// Kill and Restart
	cmd.Process.Kill()
	cmd.Wait()

	cmd = exec.Command(serverBin, "run", "--config-path", configPath, "--debug", "--api-key", "test-key")
	cmd.Stdout = logFile
	cmd.Stderr = logFile
	err = cmd.Start()
	require.NoError(t, err)
	baseURL = fmt.Sprintf("http://127.0.0.1:%s", port2)

	verifyEndpointLifecycle(t, fmt.Sprintf("%s/healthz", baseURL), 200, 60*time.Second)
	verifyEndpointLifecycle(t, fmt.Sprintf("%s/api/v1/tools", baseURL), 200, 30*time.Second)

	// CUJ 3: Disable Service
	config3 := fmt.Sprintf(`
global_settings:
  mcp_listen_address: "127.0.0.1:%s"
upstream_services:
  - id: "fs-service"
    name: "Filesystem Service"
    disable: true
    auto_discover_tool: true
    mcp_service:
      stdio_connection:
        command: "%s"
        args: []
`, port3, mockServerBin)

	err = os.WriteFile(configPath, []byte(config3), 0644)
	require.NoError(t, err)

	cmd.Process.Kill()
	cmd.Wait()

	cmd = exec.Command(serverBin, "run", "--config-path", configPath, "--debug", "--api-key", "test-key")
	cmd.Stdout = logFile
	cmd.Stderr = logFile
	err = cmd.Start()
	require.NoError(t, err)
	baseURL = fmt.Sprintf("http://127.0.0.1:%s", port3)

	verifyEndpointLifecycle(t, fmt.Sprintf("%s/healthz", baseURL), 200, 60*time.Second)

	// CUJ 4: Validating Topology
	topoResp, err := http.Get(fmt.Sprintf("%s/api/v1/topology?api_key=test-key", baseURL))
	require.NoError(t, err)
	defer topoResp.Body.Close()
	require.Equal(t, 200, topoResp.StatusCode)
}
