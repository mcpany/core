//go:build e2e

// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package e2e_sequential

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/exec"
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
	// Prioritize local execution
	useLocal := true

	rootDir, err := os.Getwd()
	require.NoError(t, err)
	if strings.HasSuffix(rootDir, "tests/e2e_sequential") {
		rootDir = filepath.Join(rootDir, "../../..")
	} else if strings.HasSuffix(rootDir, "server") {
		rootDir = filepath.Join(rootDir, "..")
	}
	rootDir, err = filepath.Abs(rootDir)
	require.NoError(t, err)

	configDir := filepath.Join(rootDir, "build", "e2e_config_lifecycle")
	err = os.MkdirAll(configDir, 0755)
	require.NoError(t, err)
	defer os.RemoveAll(configDir)

	// Build Server
	binPath := buildServer(t, rootDir)

	// Create a dummy file to read/list
	dummyFile := filepath.Join(configDir, "hello.txt")
	err = os.WriteFile(dummyFile, []byte("world"), 0644)
	require.NoError(t, err)

	// In local mode, paths must be absolute on the host
	dataPath := configDir

	// Initial Config: Enable Filesystem Upstream
	port := findFreePort(t)
	config1 := fmt.Sprintf(`
global_settings:
  mcp_listen_address: "127.0.0.1:%d"
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
`, port, dataPath)
	configPath := filepath.Join(configDir, "config.yaml")
	err = os.WriteFile(configPath, []byte(config1), 0644)
	require.NoError(t, err)

	var cmd *exec.Cmd
	var baseURL string

	if useLocal {
		cmd, baseURL = startServer(t, binPath, configPath, port)
	}

	defer func() {
		if cmd != nil && cmd.Process != nil {
			cmd.Process.Kill()
		}
	}()

	// CUJ 1: Health
	verifyEndpoint(t, fmt.Sprintf("%s/healthz", baseURL), 200, 60*time.Second)

	// Seed Data (Optional here as we use config file, but good practice)
	// seedData(t, baseURL, map[string]any{}) // Empty seed to ensure clean state if not relying on restart

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
	// Local process restart is just killing and starting again
    // Update config
    port2 := findFreePort(t)
    config2 := strings.Replace(config1, fmt.Sprintf("127.0.0.1:%d", port), fmt.Sprintf("127.0.0.1:%d", port2), 1)
	config2 = strings.Replace(config2, "enabled: true", "enabled: true\n        \"second-service\":\n          enabled: true", 1) + fmt.Sprintf(`
  - id: "second-service"
    name: "Second Service"
    auto_discover_tool: true
    filesystem_service:
      root_paths:
        "/data": "%s"
      os: {}
`, dataPath)
    t.Logf("Config 2 Content:\n%s", config2)
	err = os.WriteFile(configPath, []byte(config2), 0644)
	require.NoError(t, err)

    if useLocal {
        cmd.Process.Kill()
        cmd.Wait()
		cmd, baseURL = startServer(t, binPath, configPath, port2)
    }

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
    port3 := findFreePort(t)
	config3 := fmt.Sprintf(`
global_settings:
  mcp_listen_address: "127.0.0.1:%d"
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
`, port3, dataPath)
	err = os.WriteFile(configPath, []byte(config3), 0644)
	require.NoError(t, err)

    if useLocal {
        cmd.Process.Kill()
        cmd.Wait()
		cmd, baseURL = startServer(t, binPath, configPath, port3)
    }

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
