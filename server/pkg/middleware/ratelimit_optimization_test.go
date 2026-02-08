package middleware_test

import (
	"context"
	"encoding/json"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/mcpany/core/server/pkg/middleware"
	"github.com/mcpany/core/server/pkg/tool"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

func TestRateLimitMiddleware_ArgumentsPopulated(t *testing.T) {
	mockToolManager := &rateLimitMockToolManager{}
	rlMiddleware := middleware.NewRateLimitMiddleware(mockToolManager)

	toolProto := v1.Tool_builder{
		ServiceId: proto.String("service"),
	}.Build()
	mockTool := &rateLimitMockTool{toolProto: toolProto}

	rlConfig := configv1.RateLimitConfig_builder{
		IsEnabled:         true,
		RequestsPerSecond: 10,
		Burst:             10,
		CostMetric:        configv1.RateLimitConfig_COST_METRIC_TOKENS, // Force parsing
	}.Build()

	serviceInfo := &tool.ServiceInfo{
		Name: "test-service",
		Config: configv1.UpstreamServiceConfig_builder{
			RateLimit: rlConfig,
		}.Build(),
	}
	mockToolManager.On("GetTool", "service.test-tool").Return(mockTool, true)
	mockToolManager.On("GetServiceInfo", "service").Return(serviceInfo, true)

	// Create request with ToolInputs but no Arguments
	toolInputs := map[string]interface{}{"foo": "bar"}
	toolInputsBytes, _ := json.Marshal(toolInputs)

	req := &tool.ExecutionRequest{
		ToolName:   "service.test-tool",
		ToolInputs: json.RawMessage(toolInputsBytes),
		Arguments:  nil,
	}
	ctx := tool.NewContextWithTool(context.Background(), mockTool)

	next := func(_ context.Context, req *tool.ExecutionRequest) (any, error) {
		// Verify Arguments are populated in next middleware
		assert.NotNil(t, req.Arguments)
		assert.Equal(t, "bar", req.Arguments["foo"])
		return "success", nil
	}

	result, err := rlMiddleware.Execute(ctx, req, next)
	assert.NoError(t, err)
	assert.Equal(t, "success", result)

	// Verify Arguments are populated in original request object (since it's a pointer)
	assert.NotNil(t, req.Arguments)
	assert.Equal(t, "bar", req.Arguments["foo"])
}
