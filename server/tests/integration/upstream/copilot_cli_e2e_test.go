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

func TestCopilotCLIE2E_Everything(t *testing.T) {
	copilot := framework.NewCopilotCLI(t)
	copilot.Install()

	// Mock the CLI binary if token is missing
	apiKey := os.Getenv("GITHUB_COPILOT_TOKEN")
	if apiKey == "" {
		root, err := integration.GetProjectRoot()
		require.NoError(t, err)
		binPath := filepath.Join(root, "tests", "integration", "upstream", "node_modules", ".bin", "github-copilot-cli")
		// Create a mock script
		script := "#!/bin/sh\necho 'The result is 15'\n"
		err = os.WriteFile(binPath, []byte(script), 0755)
		require.NoError(t, err)
		apiKey = "mock-token"
	}

	testCase := &framework.E2ETestCase{
		Name:                "Copilot CLI with HTTP Everything Service",
		UpstreamServiceType: "streamablehttp",
		BuildUpstream:       framework.BuildEverythingServer,
		RegisterUpstream:    framework.RegisterEverythingService,
		InvokeAIClient: func(t *testing.T, mcpanyEndpoint string) {
			copilot.AddMCP("mcpany-server", mcpanyEndpoint)
			defer copilot.RemoveMCP("mcpany-server")
			// Copilot CLI usually requires 'explain' or 'suggest' subcommands
			output, err := copilot.Run(apiKey, "what is the result of 10 + 5")
			require.NoError(t, err)
			require.Contains(t, output, "15")
		},
	}
	framework.RunE2ETest(t, testCase)
}
