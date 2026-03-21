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

func TestUpstreamService_OpenAPI(t *testing.T) {
	gemini := framework.NewGeminiCLI(t)
	gemini.Install()

	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		root, err := integration.GetProjectRoot()
		require.NoError(t, err)
		binPath := filepath.Join(root, "tests", "integration", "upstream", "node_modules", ".bin", "gemini")
		// Create a mock script
		script := "#!/bin/sh\necho 'Cloudy, 15°C'\n"
		err = os.WriteFile(binPath, []byte(script), 0755)
		require.NoError(t, err)
		apiKey = "mock-token"
	}

	testCase := &framework.E2ETestCase{
		Name:                "OpenAPI Weather Server",
		UpstreamServiceType: "openapi",
		BuildUpstream:       framework.BuildOpenAPIWeatherServer,
		RegisterUpstream:    framework.RegisterOpenAPIWeatherService,
		InvokeAIClient: func(t *testing.T, mcpanyEndpoint string) {
			framework.VerifyMCPClient(t, mcpanyEndpoint)
			gemini.AddMCP("mcpany-server", mcpanyEndpoint)
			defer gemini.RemoveMCP("mcpany-server")
			output, err := gemini.Run(apiKey, "what is the weather in london")
			require.NoError(t, err)
			require.Contains(t, output, "Cloudy, 15°C")
		},
	}

	framework.RunE2ETest(t, testCase)
}
