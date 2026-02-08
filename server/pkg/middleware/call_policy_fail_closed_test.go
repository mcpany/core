package middleware_test

import (
	"context"
	"encoding/json"
	"testing"

	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/mcpany/core/server/pkg/middleware"
	"github.com/mcpany/core/server/pkg/tool"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

// Reusing mocks from call_policy_test.go.

func TestCallPolicyMiddleware_FailClosed(t *testing.T) {
	mockToolManager := &callPolicyMockToolManager{}
	cpMiddleware := middleware.NewCallPolicyMiddleware(mockToolManager)

	toolProto := v1.Tool_builder{
		ServiceId: proto.String("missing-service"),
	}.Build()
	mockTool := &callPolicyMockTool{toolProto: toolProto}

	// Tool exists
	mockToolManager.On("GetTool", "service.test-tool").Return(mockTool, true)
	// Service info MISSING (returns false)
	mockToolManager.On("GetServiceInfo", "missing-service").Return(nil, false)

	req := &tool.ExecutionRequest{
		ToolName:   "service.test-tool",
		ToolInputs: json.RawMessage(`{}`),
	}
	ctx := tool.NewContextWithTool(context.Background(), mockTool)

	nextCalled := false
	next := func(_ context.Context, _ *tool.ExecutionRequest) (any, error) {
		nextCalled = true
		return "success", nil
	}

	result, err := cpMiddleware.Execute(ctx, req, next)

	// We expect an error (Fail Closed)
	assert.Error(t, err, "Expected error when service info is missing")
	if err != nil {
		assert.Contains(t, err.Error(), "service info not found")
	}
	assert.Nil(t, result)
	assert.False(t, nextCalled, "Next handler should not be called")
}
