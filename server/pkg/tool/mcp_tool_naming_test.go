package tool_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/mcpany/core/server/pkg/tool"
	"github.com/mcpany/core/server/pkg/util"
	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/require"
)

func TestMCPTool_Execute_Bug_OriginalName(t *testing.T) {
	t.Parallel()
	// This test demonstrates the fix where MCPTool uses the original name (from tool definition)
	// instead of the sanitized name (from request parsing) when calling the upstream service.

	originalName := "my.tool"
	// In util.go: nonWordChars = regexp.MustCompile(`[^a-zA-Z0-9_-]+`)
	// "." matches this regex, so it is replaced by "".
	// So "my.tool" -> "mytool" and since length changed, hash is appended.

	sanitizedName, err := util.SanitizeToolName(originalName)
	require.NoError(t, err)

	serviceID := "myservice"
	// The tool is registered as "myservice.sanitizedName".

	// Upstream expects "my.tool".

	mockClient := &mockMCPClient{
		callToolFunc: func(_ context.Context, params *mcp.CallToolParams) (*mcp.CallToolResult, error) {
			// This assertion confirms the fix.
			// It should send "my.tool", not "mytool_...".
			if params.Name != originalName {
				return nil, errors.New("wrong tool name: " + params.Name)
			}
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: `{"output":"result"}`},
				},
			}, nil
		},
	}

	toolProto := &v1.Tool{}
	toolProto.SetName(originalName)   // "my.tool"
	toolProto.SetServiceId(serviceID) // "myservice"
	mcpTool := tool.NewMCPTool(toolProto, mockClient, &configv1.MCPCallDefinition{})

	inputs := json.RawMessage(`{}`)
	// The request comes in with the registered name
	req := &tool.ExecutionRequest{
		ToolName:   serviceID + "." + sanitizedName,
		ToolInputs: inputs,
	}

	_, err = mcpTool.Execute(context.Background(), req)

	// We assert that it succeeds.
	require.NoError(t, err)
}
