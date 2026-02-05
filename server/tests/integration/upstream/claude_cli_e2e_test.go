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

func TestClaudeCLIE2E_Everything(t *testing.T) {
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		t.Skip("ANTHROPIC_API_KEY not set, skipping test")
	}

	claude := framework.NewClaudeCLI(t)
	claude.Install()

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
