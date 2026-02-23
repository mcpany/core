//go:build e2e

// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package e2e_sequential

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/require"
)

// TestCUJ_Lifecycle_And_Config tests lifecycle events and config changes.
// Using Filesystem upstream to avoid dependency on external binaries or containers.
func TestCUJ_Lifecycle_And_Config(t *testing.T) {
	binPath := BuildServer(t)

	rootDir := findRootDir(t)
	configDir := filepath.Join(rootDir, "build", "e2e_config_lifecycle")
	err := os.MkdirAll(configDir, 0755)
	require.NoError(t, err)
	defer os.RemoveAll(configDir)

	// Create a dummy file to read/list
	dummyFile := filepath.Join(configDir, "hello.txt")
	err = os.WriteFile(dummyFile, []byte("world"), 0644)
	require.NoError(t, err)

	dataPath := configDir

	// Initial Config: Enable Filesystem Upstream
	// Note: We don't need mcp_listen_address here because StartServerProcess overrides it via flag.
	// But we do need it for the logic inside the test that updates config?
	// StartServerProcess uses "--mcp-listen-address" flag which takes precedence over config file usually.
	// Let's verify precedence. `server.go` says:
	// bindAddress := opts.JSONRPCPort (from flag)
	// if cfg.GlobalSettings.McpListenAddress != "" { bindAddress = ... }
	// So Config File OVERRIDES flag!
	// This means we must ensure config file does NOT set it, OR sets it to what we want.
	// But StartServerProcess picks a random port.
	// If config file has a fixed port, it might conflict.
	// If config file has "127.0.0.1:0", server might pick random, but we won't know which one unless we parse logs.
	//
	// Strategy:
	// StartServerProcess sets the port via flag.
	// We should OMIT mcp_listen_address from config file so flag is used.
	// OR we update config file with the port chosen by StartServerProcess?
	// StartServerProcess picks port -> runs command.
	// If config file overrides it, we are in trouble.
	//
	// Let's omit mcp_listen_address from config yaml.

	config1 := fmt.Sprintf(`
global_settings:
  # mcp_listen_address omitted to use flag
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
    filesystem_service:
      root_paths:
        "/data": "%s"
      os: {}
`, dataPath)
	configPath := filepath.Join(configDir, "config.yaml")
	err = os.WriteFile(configPath, []byte(config1), 0644)
	require.NoError(t, err)

	sp := StartServerProcess(t, binPath, "--config-path", configPath, "--debug", "--api-key", "test-key")
	defer sp.Stop()
	baseURL := sp.BaseURL

	// CUJ 1: Health
	verifyEndpoint(t, fmt.Sprintf("%s/healthz", baseURL), 200, 60*time.Second)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	client := mcp.NewClient(&mcp.Implementation{Name: "cuj-client", Version: "1.0"}, nil)
	transport := &mcp.StreamableClientTransport{Endpoint: baseURL + "/mcp?api_key=test-key"}
	session, err := client.Connect(ctx, transport, nil)
	require.NoError(t, err)
	defer session.Close()

	// Initial Check
	list, err := session.ListTools(ctx, nil)
	require.NoError(t, err)
	t.Logf("Initial tools: %d", len(list.Tools))

	foundFS := false
	for _, tool := range list.Tools {
		if strings.Contains(tool.Name, "list_directory") {
			foundFS = true
			break
		}
	}
	if !foundFS {
		var names []string
		for _, t := range list.Tools {
			names = append(names, t.Name)
		}
		t.Logf("Tools found: %v", names)
	}

	// CUJ 2: Hot-Reload / Restart
	// We stop the previous process and start a new one to simulate restart/reload behavior with new config
	sp.Stop()

	config2 := strings.Replace(config1, "enabled: true", "enabled: true\n        \"second-service\":\n          enabled: true", 1) + fmt.Sprintf(`
  - id: "second-service"
    name: "Second Service"
    auto_discover_tool: true
    filesystem_service:
      root_paths:
        "/data": "%s"
      os: {}
`, dataPath)
	err = os.WriteFile(configPath, []byte(config2), 0644)
	require.NoError(t, err)

	sp = StartServerProcess(t, binPath, "--config-path", configPath, "--debug", "--api-key", "test-key")
	defer sp.Stop()
	baseURL = sp.BaseURL

	// Wait for health
	verifyEndpoint(t, fmt.Sprintf("%s/healthz", baseURL), 200, 60*time.Second)

	// Re-connect
	transport = &mcp.StreamableClientTransport{Endpoint: baseURL + "/mcp?api_key=test-key"}
	session, err = client.Connect(ctx, transport, nil)
	require.NoError(t, err)
	defer session.Close()

	require.Eventually(t, func() bool {
		l, e := session.ListTools(ctx, nil)
		if e != nil {
			return false
		}
		for _, tool := range l.Tools {
			if strings.Contains(tool.Name, "read_file") && strings.Contains(tool.Name, "SecondService") {
				return true
			}
		}
		return false
	}, 15*time.Second, 1*time.Second, "New tool 'read_file' should appear")

	// CUJ 3: Disable
	sp.Stop()

	config3 := fmt.Sprintf(`
global_settings:
  # Omitted port
upstream_services:
  - id: "fs-service"
    name: "Filesystem Service"
    disable: true
    filesystem_service:
      root_paths:
        "/data": "%s"
      os: {}
      tools:
        - name: "list_files"
`, dataPath)
	err = os.WriteFile(configPath, []byte(config3), 0644)
	require.NoError(t, err)

	sp = StartServerProcess(t, binPath, "--config-path", configPath, "--debug", "--api-key", "test-key")
	defer sp.Stop()
	baseURL = sp.BaseURL

	// Wait for health
	verifyEndpoint(t, fmt.Sprintf("%s/healthz", baseURL), 200, 60*time.Second)

	// Re-connect
	transport = &mcp.StreamableClientTransport{Endpoint: baseURL + "/mcp?api_key=test-key"}
	session, err = client.Connect(ctx, transport, nil)
	require.NoError(t, err)
	defer session.Close()

	require.Eventually(t, func() bool {
		l, e := session.ListTools(ctx, nil)
		if e != nil {
			return false
		}
		for _, tool := range l.Tools {
			if strings.Contains(tool.Name, "list_directory") && strings.Contains(tool.Name, "FilesystemService") {
				return false
			}
		}
		return true
	}, 15*time.Second, 1*time.Second, "Tool 'list_directory' should disappear")

	// CUJ 4: Validating Topology
	topoResp, err := http.Get(fmt.Sprintf("%s/api/v1/topology?api_key=test-key", baseURL))
	require.NoError(t, err)
	defer topoResp.Body.Close()
	require.Equal(t, 200, topoResp.StatusCode)
}
