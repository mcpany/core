package tool_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/mcpany/core/server/pkg/client"
	"github.com/mcpany/core/server/pkg/tool"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

type webhookMockMCPClient struct {
	client.MCPClient
	callToolResponse *mcp.CallToolResult
	callToolErr      error
}

func (m *webhookMockMCPClient) CallTool(_ context.Context, _ *mcp.CallToolParams) (*mcp.CallToolResult, error) {
	return m.callToolResponse, m.callToolErr
}

func TestMCPTool_Execute_InputTransformation_Webhook(t *testing.T) {
	t.Setenv("MCPANY_DANGEROUS_ALLOW_LOCAL_IPS", "true")

	webhookServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/cloudevents+json")
		responseEvent := `{
			"specversion": "1.0",
			"type": "com.mcpany.tool.transform_input.response",
			"source": "webhook-test",
			"id": "123",
			"data": {"transformed": "input"}
		}`
		w.Write([]byte(responseEvent))
	}))
	defer webhookServer.Close()

	mcpClient := &webhookMockMCPClient{
		callToolResponse: &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{
					Text: `{"status": "ok"}`,
				},
			},
		},
	}

	callDef := configv1.MCPCallDefinition_builder{
		InputTransformer: configv1.InputTransformer_builder{
			Webhook: configv1.WebhookConfig_builder{
				Url: webhookServer.URL,
			}.Build(),
		}.Build(),
	}.Build()

	mcpTool := v1.Tool_builder{
		Name:                proto.String("test-tool"),
	}.Build()

	toolInstance := tool.NewMCPTool(mcpTool, mcpClient, callDef)

	req := &tool.ExecutionRequest{
		ToolName:   "test-tool",
		ToolInputs: json.RawMessage(`{}`),
	}
	_, err := toolInstance.Execute(context.Background(), req)
	require.NoError(t, err)
}
