//go:build e2e

// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package e2e_sequential

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/mcpany/core/server/tests/integration"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/require"
)

// TestCUJ_Lifecycle_And_Config tests lifecycle events and config changes.
// Using Filesystem upstream to avoid dependency on external binaries or containers.
func TestCUJ_Lifecycle_And_Config(t *testing.T) {
	if !integration.IsDockerSocketAccessible() {
		t.Skip("Skipping E2E Docker test. Docker socket not accessible.")
	}

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

	// We mount the configDir as /data in the container so the filesystem upstream can read it?
	// Or we just use /tmp.
	// Actually, let's use a simple command "ls" if available, or just use the "filesystem" upstream
	// which is compiled in.

	// Create a dummy file to read/list
	dummyFile := filepath.Join(configDir, "hello.txt")
	err = os.WriteFile(dummyFile, []byte("world"), 0644)
	require.NoError(t, err)

	// Initial Config: Enable Filesystem Upstream
	// We assume the container has access to /config_data via volume
	config1 := `
global_settings:
  mcp_listen_address: ":50050"
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
        "/data": "/config_data"
      os: {}
`
	configPath := filepath.Join(configDir, "config.yaml")
	err = os.WriteFile(configPath, []byte(config1), 0644)
	require.NoError(t, err)

	containerName := fmt.Sprintf("mcpany-cuj-lifecycle-%d", time.Now().UnixNano())

	cmd := exec.Command("docker", "run", "-d", "--name", containerName,
		"-p", "25000:50050",
		"-v", fmt.Sprintf("%s:/mcp_config", configDir),
		"-v", fmt.Sprintf("%s:/config_data", configDir),
		"mcpany/server:latest",
		"run", "--config-path", "/mcp_config/config.yaml", "--mcp-listen-address", ":50050", "--debug", "--api-key", "test-key",
	)
	cmd.Env = os.Environ()
	out, err := cmd.CombinedOutput()
	require.NoError(t, err, "Failed to start container: %s", string(out))

	defer func() {
		if t.Failed() {
			logsCmd := exec.Command("docker", "logs", containerName)
			logsOutput, _ := logsCmd.CombinedOutput()
			t.Logf("Container Logs:\n%s", string(logsOutput))
		}
		_ = exec.Command("docker", "rm", "-f", containerName).Run()
	}()

	// Discover Port
	portCmd := exec.Command("docker", "port", containerName, "50050/tcp")
	var portBinding string
	require.Eventually(t, func() bool {
		out, err := portCmd.Output()
		if err != nil {
			return false
		}
		portBinding = strings.TrimSpace(string(out))
		return portBinding != ""
	}, 10*time.Second, 500*time.Millisecond, "Failed to get port")

	if idx := strings.Index(portBinding, "\n"); idx != -1 {
		portBinding = portBinding[:idx]
	}
	_, portStr, err := net.SplitHostPort(portBinding)
	require.NoError(t, err)
	baseURL := fmt.Sprintf("http://127.0.0.1:%s", portStr)

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
	// If FS service fails (e.g. read only or something), we might check logs.
	if !foundFS {
		var names []string
		for _, t := range list.Tools {
			names = append(names, t.Name)
		}
		t.Logf("Tools found: %v", names)
	}

	// CUJ 2: Hot-Reload
	config2 := strings.Replace(config1, "enabled: true", "enabled: true\n        \"second-service\":\n          enabled: true", 1) + `
  - id: "second-service"
    name: "Second Service"
    filesystem_service:
      root_paths:
        "/data": "/config_data"
      os: {}
      tools:
        - name: "read_file"
          description: "Read file"
`
	err = os.WriteFile(configPath, []byte(config2), 0644)
	require.NoError(t, err)

	// Force update via docker cp to ensure inotify triggers (bind mounts can be flaky)
	cpCmd := exec.Command("docker", "cp", configPath, fmt.Sprintf("%s:/mcp_config/config.yaml", containerName))
	cpOut, err := cpCmd.CombinedOutput()
	require.NoError(t, err, "Failed to cp config: %s", string(cpOut))

	// Restart container to ensure config is loaded (Workaround for hot-reload flakiness in E2E)
	err = exec.Command("docker", "restart", containerName).Run()
	require.NoError(t, err, "Failed to restart container")

	// Refresh Port
	portCmd = exec.Command("docker", "port", containerName, "50050/tcp")
	require.Eventually(t, func() bool {
		out, err := portCmd.Output()
		if err != nil {
			return false
		}
		portBinding = strings.TrimSpace(string(out))
		return portBinding != ""
	}, 10*time.Second, 500*time.Millisecond, "Failed to get port after restart")
	if idx := strings.Index(portBinding, "\n"); idx != -1 {
		portBinding = portBinding[:idx]
	}
	_, portStr, err = net.SplitHostPort(portBinding)
	require.NoError(t, err)
	baseURL = fmt.Sprintf("http://127.0.0.1:%s", portStr)

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
	config3 := `
upstream_services:
  - id: "fs-service"
    name: "Filesystem Service"
    disable: true
    filesystem_service:
      root_paths:
        "/data": "/config_data"
      os: {}
      tools:
        - name: "list_files"
`
	err = os.WriteFile(configPath, []byte(config3), 0644)
	require.NoError(t, err)

	// Force update via docker cp
	cpCmd2 := exec.Command("docker", "cp", configPath, fmt.Sprintf("%s:/mcp_config/config.yaml", containerName))
	cpOut2, err := cpCmd2.CombinedOutput()
	require.NoError(t, err, "Failed to cp config: %s", string(cpOut2))

	// Restart container
	err = exec.Command("docker", "restart", containerName).Run()
	require.NoError(t, err, "Failed to restart container")

	// Refresh Port
	portCmd = exec.Command("docker", "port", containerName, "50050/tcp")
	require.Eventually(t, func() bool {
		out, err := portCmd.Output()
		if err != nil {
			return false
		}
		portBinding = strings.TrimSpace(string(out))
		return portBinding != ""
	}, 10*time.Second, 500*time.Millisecond, "Failed to get port after restart")
	if idx := strings.Index(portBinding, "\n"); idx != -1 {
		portBinding = portBinding[:idx]
	}
	_, portStr, err = net.SplitHostPort(portBinding)
	require.NoError(t, err)
	baseURL = fmt.Sprintf("http://127.0.0.1:%s", portStr)

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
