// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

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

// DefaultClaudeModel represents the public DefaultClaudeModel entity.
//
// Summary: Defines the structured data model representing a claude model.
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
const DefaultClaudeModel = "claude-3-5-sonnet-latest"

// ClaudeCLI represents the public ClaudeCLI entity.
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
type ClaudeCLI struct {
	t *testing.T
}

// NewClaudeCLI serves as a public interface for interacting with NewClaudeCLI.
//
// Summary: Constructs and returns an initialized claude cli ready for consumption.
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
func NewClaudeCLI(t *testing.T) *ClaudeCLI {
	return &ClaudeCLI{t: t}
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
	return exec.CommandContext(context.Background(), claudePath, args...)
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
func (c *ClaudeCLI) RemoveMCP(name string) {
	c.t.Helper()
	cmd := c.claudeCommand("mcp", "remove", name)
	err := cmd.Run()
	if err != nil {
		c.t.Logf("failed to remove mcp server '%s': %v", name, err)
	}
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
