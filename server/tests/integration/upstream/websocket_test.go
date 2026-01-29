//go:build e2e

package upstream

import (
	"os"
	"testing"

	"github.com/mcpany/core/server/tests/framework"
	"github.com/stretchr/testify/require"
)

func TestUpstreamService_Websocket(t *testing.T) {
	gemini := framework.NewGeminiCLI(t)
	gemini.Install()

	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		// t.Skip("GEMINI_API_KEY is not set")
	}

	testCase := &framework.E2ETestCase{
		Name:                "Websocket Weather Server",
		UpstreamServiceType: "websocket",
		BuildUpstream:       framework.BuildWebsocketWeatherServer,
		RegisterUpstream:    framework.RegisterWebsocketWeatherService,
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
