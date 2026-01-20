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

	var config Config
	err = json.Unmarshal(data, &config)
	require.NoError(t, err)

	assert.Contains(t, config.MCPServers, "mcpany")
	assert.Equal(t, "docker", config.MCPServers["mcpany"].Command)

	// --- Test 2: Update Install ---
	// Add another server
	config.MCPServers["other"] = MCPServer{Command: "echo"}
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

	assert.Contains(t, config.MCPServers, "mcpany")
	assert.Contains(t, config.MCPServers, "other") // Should preserve existing

	// Since os.Executable returns the test binary path, we expect it to be used.
	assert.NotEqual(t, "docker", config.MCPServers["mcpany"].Command)
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
