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

// CopilotCLI handles interactions with the GitHub Copilot CLI tool for testing.
type CopilotCLI struct {
	t         *testing.T
	configDir string
	servers   map[string]MCPServerConfig
}

// MCPServerConfig defines the configuration for an MCP server.
type MCPServerConfig struct {
	Command string   `json:"command,omitempty"`
	Args    []string `json:"args,omitempty"`
	URL     string   `json:"url,omitempty"`
	Type    string   `json:"type"` // "local", "http", "sse"
}

// MCPConfig defines the configuration file structure.
type MCPConfig struct {
	MCPServers map[string]MCPServerConfig `json:"mcpServers"`
}

// NewCopilotCLI creates a new CopilotCLI instance.
//
// t is the t.
//
// Returns the result.
func NewCopilotCLI(t *testing.T) *CopilotCLI {
	tempDir := t.TempDir()
	return &CopilotCLI{
		t:         t,
		configDir: tempDir, // Use a temp dir for XDG_CONFIG_HOME
		servers:   make(map[string]MCPServerConfig),
	}
}

// Install installs the Copilot CLI tool.
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
	return exec.CommandContext(context.Background(), copilotPath, args...) //nolint:gosec // Test utility
}

// AddMCP adds an MCP server to the Copilot CLI configuration by writing to mcp-config.json.
//
// name is the name of the resource.
// endpoint is the endpoint.
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

// RemoveMCP removes an MCP server.
//
// name is the name of the resource.
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

// Run executes a prompt.
//
// apiKey is the apiKey.
// prompt is the prompt.
//
// Returns the result.
// Returns an error if the operation fails.
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
