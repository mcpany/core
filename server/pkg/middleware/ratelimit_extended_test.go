// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware_test

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/mcpany/core/pkg/auth"
	"github.com/mcpany/core/pkg/middleware"
	"github.com/mcpany/core/pkg/tool"
	"github.com/mcpany/core/pkg/util"
	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/protobuf/proto"
)

func TestRateLimitMiddleware_Partitioning(t *testing.T) {
	const successResult = "success"

	// Helper to create middleware and mocks
	setup := func(t *testing.T, keyBy configv1.RateLimitConfig_KeyBy, rps float64, burst int64) (*middleware.RateLimitMiddleware, *rateLimitMockToolManager, *rateLimitMockTool) {
		mockToolManager := &rateLimitMockToolManager{}
		rlMiddleware := middleware.NewRateLimitMiddleware(mockToolManager)

		toolProto := v1.Tool_builder{
			ServiceId: proto.String("service"),
		}.Build()
		mockTool := &rateLimitMockTool{toolProto: toolProto}

		rlConfig := configv1.RateLimitConfig_builder{
			IsEnabled:         proto.Bool(true),
			RequestsPerSecond: proto.Float64(rps),
			Burst:             proto.Int64(burst),
			KeyBy:             &keyBy,
		}.Build()

		serviceInfo := &tool.ServiceInfo{
			Name: "test-service",
			Config: configv1.UpstreamServiceConfig_builder{
				RateLimit: rlConfig,
			}.Build(),
		}

		mockToolManager.On("GetTool", mock.Anything).Return(mockTool, true)
		mockToolManager.On("GetServiceInfo", "service").Return(serviceInfo, true)

		return rlMiddleware, mockToolManager, mockTool
	}

	t.Run("Partition by IP", func(t *testing.T) {
		// RPS 1, Burst 1.
		rlMiddleware, _, mockTool := setup(t, configv1.RateLimitConfig_KEY_BY_IP, 1, 1)

		req := &tool.ExecutionRequest{
			ToolName:   "service.test-tool",
			ToolInputs: json.RawMessage(`{}`),
		}

		// Request 1 from IP A
		ctxA := tool.NewContextWithTool(context.Background(), mockTool)
		ctxA = util.ContextWithRemoteIP(ctxA, "1.1.1.1")

		next := func(_ context.Context, _ *tool.ExecutionRequest) (any, error) { return successResult, nil }

		// IP A consumes 1 token
		_, err := rlMiddleware.Execute(ctxA, req, next)
		assert.NoError(t, err)

		// IP A should be blocked now
		_, err = rlMiddleware.Execute(ctxA, req, next)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "rate limit exceeded")

		// Request 2 from IP B (should be allowed)
		ctxB := tool.NewContextWithTool(context.Background(), mockTool)
		ctxB = util.ContextWithRemoteIP(ctxB, "2.2.2.2")

		_, err = rlMiddleware.Execute(ctxB, req, next)
		assert.NoError(t, err)
	})

	t.Run("Partition by IP - Unknown", func(t *testing.T) {
		rlMiddleware, _, mockTool := setup(t, configv1.RateLimitConfig_KEY_BY_IP, 1, 1)
		req := &tool.ExecutionRequest{ToolName: "service.test-tool", ToolInputs: json.RawMessage(`{}`)}

		// Context without IP
		ctx := tool.NewContextWithTool(context.Background(), mockTool)
		next := func(_ context.Context, _ *tool.ExecutionRequest) (any, error) { return successResult, nil }

		// Consumes token for "ip:unknown"
		_, err := rlMiddleware.Execute(ctx, req, next)
		assert.NoError(t, err)

		// Should be blocked
		_, err = rlMiddleware.Execute(ctx, req, next)
		assert.Error(t, err)
	})

	t.Run("Partition by User ID", func(t *testing.T) {
		rlMiddleware, _, mockTool := setup(t, configv1.RateLimitConfig_KEY_BY_USER_ID, 1, 1)
		req := &tool.ExecutionRequest{ToolName: "service.test-tool", ToolInputs: json.RawMessage(`{}`)}

		ctxA := tool.NewContextWithTool(context.Background(), mockTool)
		ctxA = auth.ContextWithUser(ctxA, "userA")

		next := func(_ context.Context, _ *tool.ExecutionRequest) (any, error) { return successResult, nil }

		_, err := rlMiddleware.Execute(ctxA, req, next)
		assert.NoError(t, err)

		_, err = rlMiddleware.Execute(ctxA, req, next)
		assert.Error(t, err)

		// User B
		ctxB := tool.NewContextWithTool(context.Background(), mockTool)
		ctxB = auth.ContextWithUser(ctxB, "userB")
		_, err = rlMiddleware.Execute(ctxB, req, next)
		assert.NoError(t, err)
	})

	t.Run("Partition by User ID - Anonymous", func(t *testing.T) {
		rlMiddleware, _, mockTool := setup(t, configv1.RateLimitConfig_KEY_BY_USER_ID, 1, 1)
		req := &tool.ExecutionRequest{ToolName: "service.test-tool", ToolInputs: json.RawMessage(`{}`)}

		ctx := tool.NewContextWithTool(context.Background(), mockTool)
		next := func(_ context.Context, _ *tool.ExecutionRequest) (any, error) { return successResult, nil }

		_, err := rlMiddleware.Execute(ctx, req, next)
		assert.NoError(t, err)

		_, err = rlMiddleware.Execute(ctx, req, next)
		assert.Error(t, err)
	})

	t.Run("Partition by API Key (Context)", func(t *testing.T) {
		rlMiddleware, _, mockTool := setup(t, configv1.RateLimitConfig_KEY_BY_API_KEY, 1, 1)
		req := &tool.ExecutionRequest{ToolName: "service.test-tool", ToolInputs: json.RawMessage(`{}`)}

		ctxA := tool.NewContextWithTool(context.Background(), mockTool)
		ctxA = auth.ContextWithAPIKey(ctxA, "keyA")

		next := func(_ context.Context, _ *tool.ExecutionRequest) (any, error) { return successResult, nil }

		_, err := rlMiddleware.Execute(ctxA, req, next)
		assert.NoError(t, err)

		_, err = rlMiddleware.Execute(ctxA, req, next)
		assert.Error(t, err)

		ctxB := tool.NewContextWithTool(context.Background(), mockTool)
		ctxB = auth.ContextWithAPIKey(ctxB, "keyB")
		_, err = rlMiddleware.Execute(ctxB, req, next)
		assert.NoError(t, err)
	})

	t.Run("Partition by API Key (Header X-API-Key)", func(t *testing.T) {
		rlMiddleware, _, mockTool := setup(t, configv1.RateLimitConfig_KEY_BY_API_KEY, 1, 1)
		req := &tool.ExecutionRequest{ToolName: "service.test-tool", ToolInputs: json.RawMessage(`{}`)}

		httpReq, _ := http.NewRequest("POST", "/", nil)
		httpReq.Header.Set("X-API-Key", "keyHeader")

		ctx := tool.NewContextWithTool(context.Background(), mockTool)
		ctx = context.WithValue(ctx, "http.request", httpReq)

		next := func(_ context.Context, _ *tool.ExecutionRequest) (any, error) { return successResult, nil }

		_, err := rlMiddleware.Execute(ctx, req, next)
		assert.NoError(t, err)

		_, err = rlMiddleware.Execute(ctx, req, next)
		assert.Error(t, err)
	})

	t.Run("Partition by API Key (Header Authorization)", func(t *testing.T) {
		rlMiddleware, _, mockTool := setup(t, configv1.RateLimitConfig_KEY_BY_API_KEY, 1, 1)
		req := &tool.ExecutionRequest{ToolName: "service.test-tool", ToolInputs: json.RawMessage(`{}`)}

		httpReq, _ := http.NewRequest("POST", "/", nil)
		httpReq.Header.Set("Authorization", "Bearer token")

		ctx := tool.NewContextWithTool(context.Background(), mockTool)
		ctx = context.WithValue(ctx, "http.request", httpReq)

		next := func(_ context.Context, _ *tool.ExecutionRequest) (any, error) { return successResult, nil }

		_, err := rlMiddleware.Execute(ctx, req, next)
		assert.NoError(t, err)

		_, err = rlMiddleware.Execute(ctx, req, next)
		assert.Error(t, err)
	})
}
