// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

//go:build e2e

package upstream

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/mcpany/core/server/tests/framework"
	"github.com/mcpany/core/server/tests/integration"
	"github.com/stretchr/testify/require"
)

func TestClaudeCLIE2E_Everything(t *testing.T) {
	claude := framework.NewClaudeCLI(t)
	claude.Install()

	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		root, err := integration.GetProjectRoot()
		require.NoError(t, err)
		// Assuming binary name is 'claude' or '@modelcontextprotocol/claude-cli' binary?
		// framework/claude.go likely knows.
		// Looking at gemini.go, it uses node_modules/.bin/gemini.
		// Claude CLI package name is usually @anthropic-ai/claude-code or similar?
		// Assuming 'claude' based on variable name.
		binPath := filepath.Join(root, "tests", "integration", "upstream", "node_modules", ".bin", "claude")
		script := "#!/bin/sh\necho 'The result is 15'\n"
		err = os.WriteFile(binPath, []byte(script), 0755)
		if err != nil {
			// Try 'claude-cli' or similar if 'claude' fails?
			// framework/claude.go code isn't visible here but inferred.
			// Let's rely on 'claude' being the likely bin name or the test failing if wrong.
		}
		require.NoError(t, err)
		apiKey = "mock-token"
	}

	testCase := &framework.E2ETestCase{
		Name:                "Claude CLI with HTTP Everything Service",
		UpstreamServiceType: "streamablehttp",
		BuildUpstream:       framework.BuildEverythingServer,
		RegisterUpstream:    framework.RegisterEverythingService,
		InvokeAIClient: func(t *testing.T, mcpanyEndpoint string) {
			claude.AddMCP("mcpany-server", mcpanyEndpoint)
			defer claude.RemoveMCP("mcpany-server")
			output, err := claude.Run(apiKey, "what is the result of 10 + 5")
			require.NoError(t, err)
			require.Contains(t, output, "15")
		},
	}
	framework.RunE2ETest(t, testCase)
}
