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

// DefaultModel is the default Gemini model to use.
//
// Summary: DefaultModel is the default Gemini model to use.
const DefaultModel = "gemini-2.5-flash"

// GeminiCLI handles interactions with the Gemini CLI tool for testing.
//
// Summary: GeminiCLI handles interactions with the Gemini CLI tool for testing.
type GeminiCLI struct {
	t *testing.T
}

// NewGeminiCLI creates a new GeminiCLI instance.
//
// Summary: NewGeminiCLI creates a new GeminiCLI instance.
//
// Parameters:
//   - t (*testing.T): The t parameter.
//
// Returns:
//   - *GeminiCLI: The *GeminiCLI result.
//
// Errors:
//   - None.
//
// Side Effects:
//   - May modify internal state or perform external calls.
func NewGeminiCLI(t *testing.T) *GeminiCLI {
	return &GeminiCLI{t: t}
}

// Install installs the Gemini CLI tool.
//
// Summary: Install installs the Gemini CLI tool.
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
//   - May modify internal state or perform external calls.
func (g *GeminiCLI) Install() {
	g.t.Helper()
	root, err := integration.GetProjectRoot()
	require.NoError(g.t, err)
	cmd := exec.CommandContext(context.Background(), "npm", "install")
	cmd.Dir = filepath.Join(root, "tests", "integration", "upstream")
	err = cmd.Run()
	require.NoError(g.t, err, "failed to install gemini-cli")
}

func (g *GeminiCLI) geminiCommand(args ...string) *exec.Cmd {
	g.t.Helper()
	root, err := integration.GetProjectRoot()
	require.NoError(g.t, err)
	geminiPath := filepath.Join(root, "tests", "integration", "upstream", "node_modules", ".bin", "gemini")
	return exec.CommandContext(context.Background(), geminiPath, args...)
}

// AddMCP adds an MCP server to the Gemini CLI configuration.
//
// Summary: AddMCP adds an MCP server to the Gemini CLI configuration.
//
// Parameters:
//   - name (string): The name parameter.
//   - endpoint (string): The endpoint parameter.
//
// Returns:
//   - None.
//
// Errors:
//   - None.
//
// Side Effects:
//   - May modify internal state or perform external calls.
func (g *GeminiCLI) AddMCP(name, endpoint string) {
	g.t.Helper()
	cmd := g.geminiCommand("mcp", "add", "--transport", "http", name, endpoint)
	err := cmd.Run()
	require.NoError(g.t, err, "failed to configure gemini-cli")
}

// RemoveMCP removes an MCP server from the Gemini CLI configuration.
//
// Summary: RemoveMCP removes an MCP server from the Gemini CLI configuration.
//
// Parameters:
//   - name (string): The name parameter.
//
// Returns:
//   - None.
//
// Errors:
//   - None.
//
// Side Effects:
//   - May modify internal state or perform external calls.
func (g *GeminiCLI) RemoveMCP(name string) {
	g.t.Helper()
	cmd := g.geminiCommand("mcp", "remove", name)
	err := cmd.Run()
	if err != nil {
		g.t.Logf("failed to remove mcp server '%s': %v", name, err)
	}
}

// Run executes a prompt against the Gemini CLI using the provided API key.
//
// Summary: Run executes a prompt against the Gemini CLI using the provided API key.
//
// Parameters:
//   - apiKey (string): The apiKey parameter.
//   - prompt (string): The prompt parameter.
//
// Returns:
//   - string: The string result.
//   - error: An error if the operation fails.
//
// Errors:
//   - Returns an error if the operation fails.
//
// Side Effects:
//   - May modify internal state or perform external calls.
func (g *GeminiCLI) Run(apiKey, prompt string) (string, error) {
	g.t.Helper()
	var outputBuffer strings.Builder
	cmd := g.geminiCommand("-m", DefaultModel, "-p", prompt)
	if apiKey != "" {
		cmd.Env = append(os.Environ(), "GEMINI_API_KEY="+apiKey)
	}
	cmd.Stdout = &outputBuffer
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	return outputBuffer.String(), err
}
