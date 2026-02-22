//go:build e2e

// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package e2e_sequential

import (
	"bytes"
	"fmt"
	"net"
	"encoding/json"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"github.com/stretchr/testify/assert"
	"time"

	"github.com/stretchr/testify/require"
)

// buildServer builds the mcpany server binary and returns its path.
// It caches the build to avoid rebuilding for every test in the same run if possible,
// but for simplicity here we just build it.
func buildServer(t *testing.T, rootDir string) string {
	t.Helper()
	binDir := filepath.Join(rootDir, "build", "bin")
	err := os.MkdirAll(binDir, 0755)
	require.NoError(t, err)

	binPath := filepath.Join(binDir, "server")
	if runtimeOS := os.Getenv("GOOS"); runtimeOS == "windows" {
		binPath += ".exe"
	}

	// We use "go build" directly
	cmd := exec.Command("go", "build", "-o", binPath, filepath.Join(rootDir, "server/cmd/server"))
	// cmd.Env = append(os.Environ(), "CGO_ENABLED=1") // Ensure CGO if needed, usually needed for sqlite
	// Actually we should rely on environment.
	out, err := cmd.CombinedOutput()
	require.NoError(t, err, "Failed to build server: %s", string(out))

	return binPath
}

// startServer starts the server process.
// Returns the cmd object (for killing later) and the base URL.
func startServer(t *testing.T, binPath string, configPath string, port int) (*exec.Cmd, string) {
	t.Helper()
	baseURL := fmt.Sprintf("http://127.0.0.1:%d", port)

	args := []string{
		"run",
		"--config-path", configPath,
		"--mcp-listen-address", fmt.Sprintf("127.0.0.1:%d", port),
		"--api-key", "test-key",
		"--debug",
	}

	cmd := exec.Command(binPath, args...)
	// Set env to allow unsafe config for tests
	cmd.Env = append(os.Environ(),
		"MCPANY_ALLOW_UNSAFE_CONFIG=true",
		"MCPANY_ALLOW_PRIVATE_NETWORK_RESOURCES=true",
		"MCPANY_DANGEROUS_ALLOW_LOCAL_IPS=true",
	)

	// Capture output for debugging if test fails
	// We can use a pipe or temp file
	// For now, let's write to a log file in the config dir
	logFile, err := os.Create(configPath + ".log")
	require.NoError(t, err)
	cmd.Stdout = logFile
	cmd.Stderr = logFile

	err = cmd.Start()
	require.NoError(t, err, "Failed to start server process")

	// Wait for health
	if !assert.Eventually(t, func() bool {
		resp, err := http.Get(baseURL + "/healthz")
		if err != nil {
			return false
		}
		defer resp.Body.Close()
		return resp.StatusCode == 200
	}, 10*time.Second, 100*time.Millisecond, "Server did not become healthy at %s", baseURL) {
		// Read log file
		logs, _ := os.ReadFile(configPath + ".log")
		t.Logf("Server logs:\n%s", string(logs))
		t.FailNow()
	}

	return cmd, baseURL
}

// findFreePort finds a free TCP port on localhost.
func findFreePort(t *testing.T) int {
	t.Helper()
	addr, err := net.ResolveTCPAddr("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to resolve tcp addr: %v", err)
	}
	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		t.Fatalf("failed to listen on tcp addr: %v", err)
	}
	defer func() { _ = l.Close() }()
	return l.Addr().(*net.TCPAddr).Port
}

// seedData calls the seed endpoint with data from the provided map/struct.
func seedData(t *testing.T, baseURL string, data interface{}) {
	t.Helper()
	jsonData, err := json.Marshal(data)
    // Need to import encoding/json
	require.NoError(t, err)

	req, err := http.NewRequest("POST", baseURL+"/api/v1/debug/seed", bytes.NewBuffer(jsonData))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", "test-key")

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode, "Seed request failed")
}

// verifyEndpoint checks if an endpoint returns the expected status code within a timeout.
func verifyEndpoint(t *testing.T, url string, expectedStatus int, timeout time.Duration) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		//nolint:gosec // G107: Url is constructed internally in test
		resp, err := http.Get(url)
		if err == nil {
			_ = resp.Body.Close()
			if resp.StatusCode == expectedStatus {
				return
			}
		}
		time.Sleep(1 * time.Second)
	}
	t.Fatalf("Failed to verify endpoint %s within %v", url, timeout)
}
