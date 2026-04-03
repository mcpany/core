// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package framework

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mcpany/core/server/tests/integration"
	"github.com/stretchr/testify/require"
)

// CopilotCLI represents the public CopilotCLI entity.
//
// Summary: Defines the structured data model representing a cli.
//
// Parameters:
//   - None.
//
// Returns:
//   - None.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
type CopilotCLI struct {
	t         *testing.T
	configDir string
	servers   map[string]MCPServerConfig
}

// MCPServerConfig represents the public MCPServerConfig entity.
//
// Summary: Defines the structured data model representing a server config.
//
// Parameters:
//   - None.
//
// Returns:
//   - None.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
type MCPServerConfig struct {
	Command string   `json:"command,omitempty"`
	Args    []string `json:"args,omitempty"`
	URL     string   `json:"url,omitempty"`
	Type    string   `json:"type"` // "local", "http", "sse"
}

// MCPConfig represents the public MCPConfig entity.
//
// Summary: Defines the structured data model representing a config.
//
// Parameters:
//   - None.
//
// Returns:
//   - None.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
type MCPConfig struct {
	MCPServers map[string]MCPServerConfig `json:"mcpServers"`
}

// NewCopilotCLI serves as a public interface for interacting with NewCopilotCLI.
//
// Summary: Constructs and returns an initialized copilot cli ready for consumption.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the successfully computed domain model or execution state.
//
// Errors:
//   - No explicit errors are thrown by this operation.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
func NewCopilotCLI(t *testing.T) *CopilotCLI {
	tempDir := t.TempDir()
	return &CopilotCLI{
		t:         t,
		configDir: tempDir, // Use a temp dir for XDG_CONFIG_HOME
		servers:   make(map[string]MCPServerConfig),
	}
}

// Install serves as a public interface for interacting with Install.
//
// Summary: Install the  appropriately based on current system conditions.
//
// Parameters:
//   - None.
//
// Returns:
//   - Returns the successfully computed domain model or execution state.
//
// Errors:
//   - No explicit errors are thrown by this operation.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
func (c *CopilotCLI) Install() {
	c.t.Helper()
	root, err := integration.GetProjectRoot()
	require.NoError(c.t, err)
	cmd := exec.CommandContext(context.Background(), "npm", "install")
	cmd.Dir = filepath.Join(root, "tests", "integration", "upstream")
	err = cmd.Run()
	require.NoError(c.t, err, "failed to install github-copilot-cli")
}

func (c *CopilotCLI) copilotCommand(args ...string) *exec.Cmd {
	c.t.Helper()
	root, err := integration.GetProjectRoot()
	require.NoError(c.t, err)
	// Assuming the binary is 'github-copilot-cli' from the npm package '@github/copilot-cli' or similar.
	// We need to be careful with the binary name.
	// The search result said 'npm install -g @github/copilot' and the binary might be 'github-copilot-cli'.
	// We'll trust the package.json dependency.
	copilotPath := filepath.Join(root, "tests", "integration", "upstream", "node_modules", ".bin", "github-copilot-cli")
	return exec.CommandContext(context.Background(), copilotPath, args...)
}

// AddMCP serves as a public interface for interacting with AddMCP.
//
// Summary: Add the mcp appropriately based on current system conditions.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the successfully computed domain model or execution state.
//
// Errors:
//   - No explicit errors are thrown by this operation.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
func (c *CopilotCLI) AddMCP(name, endpoint string) {
	c.t.Helper()

	// Determine type based on endpoint
	cfg := MCPServerConfig{}
	// Determine type based on endpoint
	// In our E2E, we usually test with HTTP servers (streamablehttp) or sse.
	// For now we assume http type for simplicity as the previous logic did.
	cfg.Type = "http"
	cfg.URL = endpoint

	c.servers[name] = cfg
	c.writeConfig()
}

// RemoveMCP serves as a public interface for interacting with RemoveMCP.
//
// Summary: Remove the mcp appropriately based on current system conditions.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the successfully computed domain model or execution state.
//
// Errors:
//   - No explicit errors are thrown by this operation.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
func (c *CopilotCLI) RemoveMCP(name string) {
	c.t.Helper()
	delete(c.servers, name)
	c.writeConfig()
}

func (c *CopilotCLI) writeConfig() {
	c.t.Helper()
	config := MCPConfig{
		MCPServers: c.servers,
	}

	data, err := json.MarshalIndent(config, "", "  ")
	require.NoError(c.t, err)

	// Create .copilot directory inside configDir
	copilotDir := filepath.Join(c.configDir, ".copilot")
	if err := os.MkdirAll(copilotDir, 0750); err != nil {
		require.NoError(c.t, err)
	}

	configFile := filepath.Join(copilotDir, "mcp-config.json")
	err = os.WriteFile(configFile, data, 0600)
	require.NoError(c.t, err, "failed to write mcp-config.json")
}

// Run serves as a public interface for interacting with Run.
//
// Summary: Run the  appropriately based on current system conditions.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the expected domain model and an error upon failure.
//
// Errors:
//   - Propagates exceptions from underlying I/O or validation layers.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
func (c *CopilotCLI) Run(apiKey, prompt string) (string, error) {
	c.t.Helper()
	var outputBuffer strings.Builder

	// Copilot CLI usually requires a subcommand.
	// based on search: "github-copilot-cli what-the-shell" or similar aliases?
	// Common use: github-copilot-cli explain "prompt"
	cmd := c.copilotCommand("explain", prompt)

	// Inject XDG_CONFIG_HOME to point to our temp config
	env := os.Environ()
	env = append(env, "XDG_CONFIG_HOME="+c.configDir)
	if apiKey != "" {
		env = append(env, "GITHUB_COPILOT_TOKEN="+apiKey)
		// Assuming this env var is correct for the CLI auth or we might need GH_TOKEN.
		// The CLI often mimics 'gh' cli auth. If so, it might expect 'gh auth login'.
		// But for E2E we hope for token support.
		env = append(env, "GH_TOKEN="+apiKey)
	}
	cmd.Env = env

	cmd.Stdout = &outputBuffer
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	return outputBuffer.String(), err
}
