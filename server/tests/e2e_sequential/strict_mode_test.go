//go:build e2e

package e2e_sequential

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestStrictFlag(t *testing.T) {
	// Locate the server binary
	rootDir, err := os.Getwd()
	require.NoError(t, err)

    // Helper to find root (repo root)
    for {
        if _, err := os.Stat(filepath.Join(rootDir, ".git")); err == nil {
            break
        }
        parent := filepath.Dir(rootDir)
        if parent == rootDir {
            // Fallback: if .git not found (e.g. in some CI/Docker envs), try to find "server" and "ui" dirs
            if _, err := os.Stat(filepath.Join(rootDir, "server")); err == nil {
                if _, err := os.Stat(filepath.Join(rootDir, "ui")); err == nil {
                     break
                }
            }
            t.Fatal("Could not find repository root")
        }
        rootDir = parent
    }

	serverBin := filepath.Join(rootDir, "build", "bin", "server")
    t.Logf("Looking for server binary at: %s", serverBin)
	if _, err := os.Stat(serverBin); os.IsNotExist(err) {
		t.Skip("Server binary not found, skipping strict mode test. Run 'make build' first.")
	}

	// Create a broken config
	configDir := t.TempDir()
	brokenConfigPath := filepath.Join(configDir, "broken.yaml")
	brokenConfig := `
global_settings:
  mcp_listen_address: "127.0.0.1:0" # Random port
upstream_services:
  - name: "broken_service"
    http_service:
      address: "https://this-domain-does-not-exist-12345.com"
      tools:
        - name: "test"
          call_id: "test"
      calls:
        test:
          endpoint_path: "/"
          method: "HTTP_METHOD_GET"
`
	err = os.WriteFile(brokenConfigPath, []byte(brokenConfig), 0644)
	require.NoError(t, err)

	// Test 1: Run with --strict and broken config -> Should Fail (Exit Code 1)
	cmd := exec.Command(serverBin, "run", "--config-path", brokenConfigPath, "--strict")
	output, err := cmd.CombinedOutput()

    // We expect an error
	require.Error(t, err, "Expected server to fail in strict mode with broken config")

    // Verify exit code if possible (ExitError)
    if exitErr, ok := err.(*exec.ExitError); ok {
        require.NotEqual(t, 0, exitErr.ExitCode(), "Exit code should be non-zero")
    }
    require.Contains(t, string(output), "strict mode validation failed", "Output should contain failure message")

	// Test 2: Run without --strict and broken config -> Should Start (we kill it)
	// We need to capture output to confirm it started or use a timeout
	cmd = exec.Command(serverBin, "run", "--config-path", brokenConfigPath)
    // Set a pipe to read output? Or just run and kill.
    // Ideally we want to see "HTTP server listening"

	err = cmd.Start()
	require.NoError(t, err, "Expected server to start without strict mode")

    done := make(chan error)
    go func() { done <- cmd.Wait() }()

    // Wait a bit to ensure it doesn't crash immediately
    select {
    case <-time.After(2 * time.Second):
        // Good, it's still running
        _ = cmd.Process.Kill()
    case err := <-done:
        t.Fatalf("Server exited unexpectedly without strict mode: %v", err)
    }
}
