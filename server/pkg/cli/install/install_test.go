// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package install

import (
	"encoding/json"
	"runtime"
	"testing"

	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRun(t *testing.T) {
	fs := afero.NewMemMapFs()

	// Create a dummy config file
	configContent := []byte("services: []")
	absConfigPath := "/tmp/mcp-config.yaml"
	if runtime.GOOS == "windows" {
		absConfigPath = `C:\tmp\mcp-config.yaml`
	}
	require.NoError(t, afero.WriteFile(fs, absConfigPath, configContent, 0644))

	opts := &Options{
		Client:      "claude",
		ConfigPath:  absConfigPath,
		LocalBinary: false, // Use Docker
		Name:        "mcpany",
		Fs:          fs,
	}

	cmd := &cobra.Command{}

	// Determine expected config path
	claudePath, err := getClaudeConfigPath()
	require.NoError(t, err)

	// --- Test 1: New Install ---
	err = Run(cmd, opts)
	require.NoError(t, err)

	exists, err := afero.Exists(fs, claudePath)
	require.NoError(t, err)
	assert.True(t, exists)

	data, err := afero.ReadFile(fs, claudePath)
	require.NoError(t, err)

	var config map[string]interface{}
	err = json.Unmarshal(data, &config)
	require.NoError(t, err)

	mcpServers, ok := config["mcpServers"].(map[string]interface{})
	assert.True(t, ok)
	assert.Contains(t, mcpServers, "mcpany")

	mcpanyServer := mcpServers["mcpany"].(map[string]interface{})
	assert.Equal(t, "docker", mcpanyServer["command"])

	// --- Test 2: Update Install & Data Preservation ---
	// Add another server and a random top-level field
	mcpServers["other"] = map[string]interface{}{"command": "echo"}
	config["globalShortcut"] = "Ctrl+Space"
	config["theme"] = "dark"

	data, _ = json.Marshal(config)
	require.NoError(t, afero.WriteFile(fs, claudePath, data, 0644))

	// Run install again with local binary
	opts.LocalBinary = true

	err = Run(cmd, opts)
	require.NoError(t, err)

	data, err = afero.ReadFile(fs, claudePath)
	require.NoError(t, err)
	err = json.Unmarshal(data, &config)
	require.NoError(t, err)

	// Check if top-level fields are preserved
	assert.Equal(t, "Ctrl+Space", config["globalShortcut"])
	assert.Equal(t, "dark", config["theme"])

	mcpServers, ok = config["mcpServers"].(map[string]interface{})
	assert.True(t, ok)

	// Check if other server is preserved
	assert.Contains(t, mcpServers, "other")

	// Check if mcpany is updated
	mcpanyServer = mcpServers["mcpany"].(map[string]interface{})
	// Since os.Executable returns the test binary path, we expect it to be used.
	assert.NotEqual(t, "docker", mcpanyServer["command"])
}

func TestRun_Errors(t *testing.T) {
	fs := afero.NewMemMapFs()
	cmd := &cobra.Command{}

	// Test 1: Missing Config Path
	opts := &Options{
		Client:     "claude",
		ConfigPath: "",
		Fs:         fs,
	}
	err := Run(cmd, opts)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "config-path is required")

	// Test 2: Non-existent Config File
	opts.ConfigPath = "/non/existent/path.yaml"
	err = Run(cmd, opts)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "configuration file not found")

	// Test 3: Unsupported Client
	// Create a dummy config file so we pass the file check
	configPath := "/tmp/valid.yaml"
	if runtime.GOOS == "windows" {
		configPath = `C:\tmp\valid.yaml`
	}
	_ = afero.WriteFile(fs, configPath, []byte("{}"), 0644)

	opts.ConfigPath = configPath
	opts.Client = "unsupported"
	err = Run(cmd, opts)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported client")
}

func TestBuildServerConfig_Docker(t *testing.T) {
	opts := &Options{
		LocalBinary: false,
	}
	absPath := "/path/to/config.yaml"
	if runtime.GOOS == "windows" {
		absPath = `C:\path\to\config.yaml`
	}

	cfg, err := buildServerConfig(opts, absPath)
	require.NoError(t, err)

	assert.Equal(t, "docker", cfg.Command)

	// Check for volume mount
	// Expected args structure: run, -i, --rm, -v, HOST:CONTAINER, ghcr..., run, ...
	foundVolume := false
	for _, arg := range cfg.Args {
		if runtime.GOOS == "windows" {
			// Window paths are tricky in tests if not careful, but our input was explicit.
			// However filepath.Dir might change separators.
			// Let's just check for ":/etc/mcpany" suffix
			if len(arg) > 12 && arg[len(arg)-12:] == ":/etc/mcpany" {
				foundVolume = true
			}
		} else {
			if arg == "/path/to:/etc/mcpany" {
				foundVolume = true
			}
		}
	}
	assert.True(t, foundVolume, "Volume mount not found in args: %v", cfg.Args)
}
