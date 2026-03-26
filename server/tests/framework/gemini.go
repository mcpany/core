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
const DefaultModel = "gemini-2.5-flash"

// GeminiCLI handles interactions with the Gemini CLI tool for testing.
// GeminiCLI handles interactions with the Gemini CLI tool for testing.
// Summary: GeminiCLI
	t *testing.T
}

// NewGeminiCLI creates a new GeminiCLI instance.
//
// t is the t.
//
// Returns the result.
// NewGeminiCLI creates a new GeminiCLI instance.
// Summary: NewGeminiCLI
//
// t is the t.
//
// Returns the result.
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	return &GeminiCLI{t: t}
}

// Install installs the Gemini CLI tool.
// Install installs the Gemini CLI tool.
// Summary: Install
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
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
// name is the name of the resource.
// endpoint is the endpoint.
// AddMCP adds an MCP server to the Gemini CLI configuration.
// Summary: AddMCP
//
// name is the name of the resource.
// endpoint is the endpoint.
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	g.t.Helper()
	cmd := g.geminiCommand("mcp", "add", "--transport", "http", name, endpoint)
	err := cmd.Run()
	require.NoError(g.t, err, "failed to configure gemini-cli")
}

// RemoveMCP removes an MCP server from the Gemini CLI configuration.
//
// name is the name of the resource.
// RemoveMCP removes an MCP server from the Gemini CLI configuration.
// Summary: RemoveMCP
//
// name is the name of the resource.
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	g.t.Helper()
	cmd := g.geminiCommand("mcp", "remove", name)
	err := cmd.Run()
	if err != nil {
		g.t.Logf("failed to remove mcp server '%s': %v", name, err)
	}
}

// Run executes a prompt against the Gemini CLI using the provided API key.
//
// apiKey is the apiKey.
// prompt is the prompt.
//
// Returns the result.
// Returns an error if the operation fails.
// Run executes a prompt against the Gemini CLI using the provided API key.
// Summary: Run
//
// apiKey is the apiKey.
// prompt is the prompt.
//
// Returns the result.
// Returns an error if the operation fails.
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
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
