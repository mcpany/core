package upstream

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/mcpany/core/server/pkg/util"
	"github.com/mcpany/core/server/tests/framework"
	"github.com/mcpany/core/server/tests/integration"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/require"
)

func TestUpstreamService_GRPC_WithBearerAuth(t *testing.T) {
	testCase := &framework.E2ETestCase{
		Name:                "Authenticated gRPC Weather Server",
		UpstreamServiceType: "grpc",
		BuildUpstream:       framework.BuildGRPCAuthedWeatherServer,
		RegisterUpstream:    framework.RegisterGRPCAuthedWeatherService,
		InvokeAIClient: func(t *testing.T, mcpanyEndpoint string) {
			ctx, cancel := context.WithTimeout(context.Background(), integration.TestWaitTimeLong)
			defer cancel()

			testMCPClient := mcp.NewClient(&mcp.Implementation{Name: "test-mcp-client", Version: "v1.0.0"}, nil)
			cs, err := testMCPClient.Connect(ctx, &mcp.StreamableClientTransport{Endpoint: mcpanyEndpoint}, nil)
			require.NoError(t, err)
			defer func() { _ = cs.Close() }()

			const weatherServiceID = "e2e_grpc_authed_weather"
			serviceID, _ := util.SanitizeServiceName(weatherServiceID)
			sanitizedToolName, _ := util.SanitizeToolName("GetWeather")
			toolName := serviceID + "." + sanitizedToolName
			weatherArgs := `{"location": "london"}`
			res, err := cs.CallTool(ctx, &mcp.CallToolParams{Name: toolName, Arguments: json.RawMessage(weatherArgs)})
			require.NoError(t, err, "Error calling GetWeather tool with correct auth")
			require.NotNil(t, res, "Nil response from GetWeather tool with correct auth")
			switch content := res.Content[0].(type) {
			case *mcp.TextContent:
				require.JSONEq(t, `{"weather": "Cloudy, 15°C"}`, content.Text, "The weather is incorrect")
			default:
				t.Fatalf("Unexpected content type: %T", content)
			}
		},
	}

	framework.RunE2ETest(t, testCase)
}
