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

// DefaultModel represents the public DefaultModel entity.
//
// Summary: Defines the structured data model representing a model.
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
const DefaultModel = "gemini-2.5-flash"

// GeminiCLI represents the public GeminiCLI entity.
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
type GeminiCLI struct {
	t *testing.T
}

// NewGeminiCLI serves as a public interface for interacting with NewGeminiCLI.
//
// Summary: Constructs and returns an initialized gemini cli ready for consumption.
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
func NewGeminiCLI(t *testing.T) *GeminiCLI {
	return &GeminiCLI{t: t}
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
func (g *GeminiCLI) AddMCP(name, endpoint string) {
	g.t.Helper()
	cmd := g.geminiCommand("mcp", "add", "--transport", "http", name, endpoint)
	err := cmd.Run()
	require.NoError(g.t, err, "failed to configure gemini-cli")
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
func (g *GeminiCLI) RemoveMCP(name string) {
	g.t.Helper()
	cmd := g.geminiCommand("mcp", "remove", name)
	err := cmd.Run()
	if err != nil {
		g.t.Logf("failed to remove mcp server '%s': %v", name, err)
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
