//go:build e2e

package upstream

import (
	"os"
	"testing"

	"github.com/mcpany/core/server/tests/framework"
	"github.com/stretchr/testify/require"
)

func TestCopilotCLIE2E_Everything(t *testing.T) {
	apiKey := os.Getenv("GITHUB_COPILOT_TOKEN")
	if apiKey == "" {
		t.Skip("GITHUB_COPILOT_TOKEN not set, skipping test")
	}

	copilot := framework.NewCopilotCLI(t)
	copilot.Install()

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
