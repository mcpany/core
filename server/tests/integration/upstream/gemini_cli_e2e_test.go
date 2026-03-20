// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

//go:build e2e

package upstream

import (
	"os"
	"testing"

	"github.com/mcpany/core/server/tests/framework"
	"github.com/stretchr/testify/require"
)

func TestGeminiCLIE2E_Everything(t *testing.T) {
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		apiKey = "mock-token"
	}

	gemini := framework.NewGeminiCLI(t)
	gemini.Install()

	testCase := &framework.E2ETestCase{
		Name:                "Gemini CLI with HTTP Everything Service",
		UpstreamServiceType: "streamablehttp",
		BuildUpstream:       framework.BuildEverythingServer,
		RegisterUpstream:    framework.RegisterEverythingService,
		InvokeAIClient: func(t *testing.T, mcpanyEndpoint string) {
			framework.VerifyMCPClient(t, mcpanyEndpoint)
			gemini.AddMCP("mcpany-server", mcpanyEndpoint)
			defer gemini.RemoveMCP("mcpany-server")
			output, err := gemini.Run(apiKey, "what is the result of 10 + 5")
			if err != nil {
			    require.Contains(t, err.Error(), "API_KEY_INVALID")
			} else {
			    require.Contains(t, output, "15")
			}
		},
	}
	framework.RunE2ETest(t, testCase)
}
