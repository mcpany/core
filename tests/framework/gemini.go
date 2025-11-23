// Copyright (C) 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package framework

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mcpany/core/tests/integration"
	"github.com/stretchr/testify/require"
)

const DefaultModel = "gemini-2.5-flash"

type GeminiCLI struct {
	t *testing.T
}

func NewGeminiCLI(t *testing.T) *GeminiCLI {
	return &GeminiCLI{t: t}
}

func (g *GeminiCLI) Install() {
	g.t.Helper()
	root, err := integration.GetProjectRoot()
	require.NoError(g.t, err)
	cmd := exec.Command("npm", "install")
	cmd.Dir = filepath.Join(root, "tests", "integration", "upstream")
	err = cmd.Run()
	require.NoError(g.t, err, "failed to install gemini-cli")
}

func (g *GeminiCLI) geminiCommand(args ...string) *exec.Cmd {
	g.t.Helper()
	root, err := integration.GetProjectRoot()
	require.NoError(g.t, err)
	geminiPath := filepath.Join(root, "tests", "integration", "upstream", "node_modules", ".bin", "gemini")
	return exec.Command(geminiPath, args...)
}

func (g *GeminiCLI) AddMCP(name, endpoint string) {
	g.t.Helper()
	cmd := g.geminiCommand("mcp", "add", "--transport", "http", name, endpoint)
	err := cmd.Run()
	require.NoError(g.t, err, "failed to configure gemini-cli")
}

func (g *GeminiCLI) RemoveMCP(name string) {
	g.t.Helper()
	cmd := g.geminiCommand("mcp", "remove", name)
	err := cmd.Run()
	if err != nil {
		g.t.Logf("failed to remove mcp server '%s': %v", name, err)
	}
}

func (g *GeminiCLI) Run(apiKey, prompt string) (string, error) {
	g.t.Helper()
	if apiKey == "" {
		g.t.Skip("GEMINI_API_KEY is not set. Please get one from AI Studio.")
	}

	var outputBuffer strings.Builder
	cmd := g.geminiCommand("-m", DefaultModel, "-p", prompt)
	cmd.Env = append(os.Environ(), "GEMINI_API_KEY="+apiKey)
	cmd.Stdout = &outputBuffer
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	return outputBuffer.String(), err
}
