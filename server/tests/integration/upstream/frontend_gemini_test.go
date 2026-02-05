// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

//go:build e2e

package upstream

import (
	"os"
	"testing"

	"github.com/mcpany/core/server/tests/framework"
	"github.com/mcpany/core/server/tests/integration"
	"github.com/stretchr/testify/require"
)

func TestGeminiCLIE2E_FrontendReact(t *testing.T) {
	apiKey := os.Getenv("GEMINI_API_KEY")
	// Check if apiKey is set, otherwise assume system is configured (as per user request)
	if apiKey == "" {
		t.Log("GEMINI_API_KEY not set, assuming system-level configuration (e.g. gcloud auth)")
	}

	gemini := framework.NewGeminiCLI(t)
	gemini.Install()

	configPath := "../../../../marketplace/upstream_service_collection/frontend_react.yaml"
	configContent, err := os.ReadFile(configPath)
	require.NoError(t, err)

	testCase := &framework.E2ETestCase{
		Name: "Gemini CLI with Frontend React Collection",
		StartMCPANYServer: func(t *testing.T, testName string, extraArgs ...string) *integration.MCPANYTestServerInfo {
			return integration.StartMCPANYServerWithConfig(t, testName, string(configContent))
		},
		InvokeAIClient: func(t *testing.T, mcpanyEndpoint string) {
			// MCP server takes a moment to be fully ready for tools?
			// StartMCPANYServerWithConfig waits for port.

			// Configure Gemini CLI to use our MCP server
			gemini.AddMCP("mcpany-frontend", mcpanyEndpoint)
			defer gemini.RemoveMCP("mcpany-frontend")

			// Ask Gemini to check npm version
			// We hope it picks "npm_exec" or "npm_run" based on description.
			// "npm_exec" handles arbitrary args but is restricted.
			// "npm_version" is not explicit, but "npm_exec" with args=["--version"] works if safe.
			// Wait, we blocked "-" flags in the previous test?
			// The previous test check for argument injection failed on "--version".
			// So "npm --version" via "npm_exec" might fail if it uses "--version".
			// We should ask for "npm root" to be safe and consistent with previous verification.

			prompt := "Check the npm root directory using the available npm tool."
			output, err := gemini.Run(apiKey, prompt)
			require.NoError(t, err)
			t.Logf("Gemini Output: %s", output)

			// Verify output contains expected content
			require.Contains(t, output, "node_modules", "Output should contain node_modules path")
		},
	}
	framework.RunE2ETest(t, testCase)
}
