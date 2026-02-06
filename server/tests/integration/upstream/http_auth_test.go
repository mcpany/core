package upstream

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	apiv1 "github.com/mcpany/core/proto/api/v1"
	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/util"
	"github.com/mcpany/core/server/tests/framework"
	"github.com/mcpany/core/server/tests/integration"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func TestUpstreamService_HTTP_WithAPIKeyAuth(t *testing.T) {
	testCase := &framework.E2ETestCase{
		Name:                "Authenticated HTTP Echo Server",
		UpstreamServiceType: "http",
		BuildUpstream:       framework.BuildHTTPAuthedEchoServer,
		RegisterUpstream:    framework.RegisterHTTPAuthedEchoService,
		InvokeAIClient: func(t *testing.T, mcpanyEndpoint string) {
			ctx, cancel := context.WithTimeout(context.Background(), integration.TestWaitTimeLong)
			defer cancel()

			testMCPClient := mcp.NewClient(&mcp.Implementation{Name: "test-mcp-client", Version: "v1.0.0"}, nil)
			cs, err := testMCPClient.Connect(ctx, &mcp.StreamableClientTransport{Endpoint: mcpanyEndpoint}, nil)
			require.NoError(t, err)
			defer func() { _ = cs.Close() }()

			const echoServiceID = "e2e_http_authed_echo"
			serviceID, _ := util.SanitizeServiceName(echoServiceID)
			sanitizedToolName, _ := util.SanitizeToolName("echo")
			toolName := serviceID + "." + sanitizedToolName
			echoMessage := `{"message": "hello world from authed http"}`
			res, err := cs.CallTool(ctx, &mcp.CallToolParams{Name: toolName, Arguments: json.RawMessage(echoMessage)})
			require.NoError(t, err, "Error calling echo tool with correct auth")
			require.NotNil(t, res, "Nil response from echo tool with correct auth")
			switch content := res.Content[0].(type) {
			case *mcp.TextContent:
				require.JSONEq(t, echoMessage, content.Text, "The echoed message does not match the original")
			default:
				t.Fatalf("Unexpected content type: %T", content)
			}
		},
	}

	framework.RunE2ETest(t, testCase)
}

func TestUpstreamService_HTTP_WithIncorrectAPIKeyAuth(t *testing.T) {
	testCase := &framework.E2ETestCase{
		Name:                "Authenticated HTTP Echo Server with Incorrect API Key",
		UpstreamServiceType: "http",
		BuildUpstream:       framework.BuildHTTPAuthedEchoServer,
		RegisterUpstream: func(t *testing.T, registrationClient apiv1.RegistrationServiceClient, upstreamEndpoint string) {
			const wrongAuthServiceID = "e2e_http_wrong_auth_echo"
			secret := configv1.SecretValue_builder{
				PlainText: proto.String("wrong-key"),
			}.Build()
			wrongAuthConfig := configv1.Authentication_builder{
				ApiKey: configv1.APIKeyAuth_builder{
					ParamName: proto.String("X-Api-Key"),
					Value:     secret,
				}.Build(),
			}.Build()
			integration.RegisterHTTPService(t, registrationClient, wrongAuthServiceID, upstreamEndpoint, "echo", "/echo", http.MethodPost, wrongAuthConfig)
		},
		InvokeAIClient: func(t *testing.T, mcpanyEndpoint string) {
			ctx, cancel := context.WithTimeout(context.Background(), integration.TestWaitTimeLong)
			defer cancel()

			testMCPClient := mcp.NewClient(&mcp.Implementation{Name: "test-mcp-client", Version: "v1.0.0"}, nil)
			cs, err := testMCPClient.Connect(ctx, &mcp.StreamableClientTransport{Endpoint: mcpanyEndpoint}, nil)
			require.NoError(t, err)
			defer func() { _ = cs.Close() }()

			const wrongAuthServiceID = "e2e_http_wrong_auth_echo"
			wrongServiceKey, _ := util.SanitizeServiceName(wrongAuthServiceID)
			sanitizedToolName, _ := util.SanitizeToolName("echo")
			wrongToolName := wrongServiceKey + "." + sanitizedToolName
			echoMessage := `{"message": "hello world from authed http"}`
			_, err = cs.CallTool(ctx, &mcp.CallToolParams{Name: wrongToolName, Arguments: json.RawMessage(echoMessage)})
			require.Error(t, err, "Expected error when calling echo tool with incorrect auth")
			require.Contains(t, err.Error(), "Unauthorized", "Expected error message to contain 'Unauthorized'")
		},
	}

	framework.RunE2ETest(t, testCase)
}
