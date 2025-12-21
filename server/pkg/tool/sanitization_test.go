package tool

import (
	"context"
	"testing"

	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

func TestToolManager_ToolNameSanitization(t *testing.T) {
	tm := NewManager(nil)

	// Use a tool name that requires sanitization (contains space)
	rawToolName := "my tool"
	serviceID := "test-service"

	mockTool := &MockTool{
		ToolFunc: func() *v1.Tool {
			return &v1.Tool{
				ServiceId: proto.String(serviceID),
				Name:      proto.String(rawToolName),
			}
		},
		ExecuteFunc: func(_ context.Context, _ *ExecutionRequest) (any, error) {
			return "success", nil
		},
	}

	err := tm.AddTool(mockTool)
	assert.NoError(t, err)

	// Derive the name as MCP would see it (using the converter used by AddTool)
	pbTool := mockTool.Tool()
	mcpTool, err := ConvertProtoToMCPTool(pbTool)
	assert.NoError(t, err)

	// Attempt to execute the tool using the name that was registered with MCP
	req := &ExecutionRequest{
		ToolName: mcpTool.Name,
	}

	result, err := tm.ExecuteTool(context.Background(), req)
	assert.NoError(t, err, "ExecuteTool should succeed with the name registered with MCP")
	assert.Equal(t, "success", result)
}
