//go:build e2e

package e2e

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStartup_FailsOnMalformedConfig(t *testing.T) {
	// Create a malformed config file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "malformed.yaml")
	// "invalid yaml content: : :" is invalid YAML mapping
	err := os.WriteFile(configPath, []byte("key: value: :"), 0644)
	require.NoError(t, err)

	// Run the server command
	// We assume we are running from server/tests/e2e/, so main is at ../../cmd/server/main.go
	dbPath := filepath.Join(tmpDir, "test.db")
	cmd := exec.Command("go", "run", "../../cmd/server/main.go", "run", "--config-path", configPath, "--mcp-listen-address", "127.0.0.1:0", "--grpc-port", "127.0.0.1:0", "--metrics-listen-address", "127.0.0.1:0")
	cmd.Env = append(os.Environ(), "MCPANY__GLOBAL_SETTINGS__DB_PATH="+dbPath, "MCPANY_ENABLE_FILE_CONFIG=true")

	// Set a timeout or rely on go test timeout
	// But usually startup failure is fast.

	output, err := cmd.CombinedOutput()

	// Expect failure
	assert.Error(t, err, "Server should have failed to start with malformed config")

	// Check if it's an exit error
	if exitErr, ok := err.(*exec.ExitError); ok {
		assert.NotEqual(t, 0, exitErr.ExitCode(), "Exit code should be non-zero")
	} else {
		// If it's not an ExitError but an error occurred, that's also fine (e.g. executable not found),
		// but we want to ensure it failed because of the config.
	}

	// Verify error output contains relevant message
	outputStr := string(output)
	// The error message from yaml unmarshal often contains "yaml: " or "cannot unmarshal"
	// Our logger logs: "Failed to parse config file, skipping ... error=..." (OLD)
	// NEW: it should return the error up the stack.
	// "Application failed" ... "error" ...
	// The underlying yaml error usually says something like "mapping values are not allowed in this context"

	// Also check for our specific log message if possible, or just the error propagation
	assert.True(t, strings.Contains(outputStr, "yaml:") || strings.Contains(outputStr, "unmarshal"),
		"Output should contain yaml/unmarshal error. Got:\n%s", outputStr)
}
