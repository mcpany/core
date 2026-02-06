package framework

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mcpany/core/server/tests/integration"
	"github.com/stretchr/testify/require"
)

// DefaultClaudeModel is the default Claude model to use.
const DefaultClaudeModel = "claude-3-5-sonnet-latest"

// ClaudeCLI handles interactions with the Claude CLI tool for testing.
type ClaudeCLI struct {
	t *testing.T
}

// NewClaudeCLI creates a new ClaudeCLI instance.
//
// t is the t.
//
// Returns the result.
func NewClaudeCLI(t *testing.T) *ClaudeCLI {
	return &ClaudeCLI{t: t}
}

// Install installs the Claude CLI tool.
func (c *ClaudeCLI) Install() {
	c.t.Helper()
	root, err := integration.GetProjectRoot()
	require.NoError(c.t, err)
	cmd := exec.CommandContext(context.Background(), "npm", "install")
	cmd.Dir = filepath.Join(root, "tests", "integration", "upstream")
	err = cmd.Run()
	require.NoError(c.t, err, "failed to install claude-code")
}

func (c *ClaudeCLI) claudeCommand(args ...string) *exec.Cmd {
	c.t.Helper()
	root, err := integration.GetProjectRoot()
	require.NoError(c.t, err)
	// Assuming the binary is 'claude'
	claudePath := filepath.Join(root, "tests", "integration", "upstream", "node_modules", ".bin", "claude")
	return exec.CommandContext(context.Background(), claudePath, args...) //nolint:gosec // Test utility
}

// AddMCP adds an MCP server to the Claude CLI configuration.
//
// name is the name of the resource.
// endpoint is the endpoint.
func (c *ClaudeCLI) AddMCP(name, endpoint string) {
	c.t.Helper()

	var args []string
	args = append(args, "mcp", "add")

	// Check if endpoint is HTTP
	if strings.HasPrefix(endpoint, "http") {
		args = append(args, "--transport", "http")
	}

	args = append(args, name, endpoint)

	cmd := c.claudeCommand(args...)
	err := cmd.Run()
	require.NoError(c.t, err, "failed to configure claude-cli")
}

// RemoveMCP removes an MCP server from the Claude CLI configuration.
//
// name is the name of the resource.
func (c *ClaudeCLI) RemoveMCP(name string) {
	c.t.Helper()
	cmd := c.claudeCommand("mcp", "remove", name)
	err := cmd.Run()
	if err != nil {
		c.t.Logf("failed to remove mcp server '%s': %v", name, err)
	}
}

// Run executes a prompt against the Claude CLI.
//
// apiKey is the apiKey.
// prompt is the prompt.
//
// Returns the result.
// Returns an error if the operation fails.
func (c *ClaudeCLI) Run(apiKey, prompt string) (string, error) {
	c.t.Helper()
	var outputBuffer strings.Builder
	// -p for prompt? -m for model?
	// Need to verify flags.
	// Assuming -p for prompt based on common conventions and Gemini.
	cmd := c.claudeCommand("-p", prompt)
	if apiKey != "" {
		cmd.Env = append(os.Environ(), "ANTHROPIC_API_KEY="+apiKey)
	}
	cmd.Stdout = &outputBuffer
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	return outputBuffer.String(), err
}
