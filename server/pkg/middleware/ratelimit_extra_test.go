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
	"google.golang.org/protobuf/proto"
)

func TestRateLimitMiddleware_Partitioning(t *testing.T) {
	const successResult = "success"

	setupMiddleware := func(keyBy configv1.RateLimitConfig_KeyBy) (*middleware.RateLimitMiddleware, *rateLimitMockToolManager, *rateLimitMockTool) {
		mockToolManager := &rateLimitMockToolManager{}
		rlMiddleware := middleware.NewRateLimitMiddleware(mockToolManager)

		toolProto := v1.Tool_builder{
			ServiceId: proto.String("service"),
		}.Build()
		mockTool := &rateLimitMockTool{toolProto: toolProto}

		rlConfig := configv1.RateLimitConfig_builder{
			IsEnabled:         proto.Bool(true),
			RequestsPerSecond: proto.Float64(1),
			Burst:             proto.Int64(1),
			KeyBy:             &keyBy,
		}.Build()

		serviceInfo := &tool.ServiceInfo{
			Name: "test-service",
			Config: configv1.UpstreamServiceConfig_builder{
				RateLimit: rlConfig,
			}.Build(),
		}
		mockToolManager.On("GetTool", "service.test-tool").Return(mockTool, true)
		mockToolManager.On("GetServiceInfo", "service").Return(serviceInfo, true)

        return rlMiddleware, mockToolManager, mockTool
    }

	t.Run("partition by ip", func(t *testing.T) {
        rlMiddleware, _, mockTool := setupMiddleware(configv1.RateLimitConfig_KEY_BY_IP)

        execute := func(ip string) error {
            req := &tool.ExecutionRequest{
                ToolName:   "service.test-tool",
                ToolInputs: json.RawMessage(`{}`),
            }
            ctx := context.Background()
            if ip != "" {
                ctx = util.ContextWithRemoteIP(ctx, ip)
            }
            ctx = tool.NewContextWithTool(ctx, mockTool)

            next := func(_ context.Context, _ *tool.ExecutionRequest) (any, error) {
                return successResult, nil
            }
            _, err := rlMiddleware.Execute(ctx, req, next)
            return err
        }

        // IP 1: Request 1 -> Allow
        assert.NoError(t, execute("1.1.1.1"))

        // IP 1: Request 2 -> Block (1 RPS exceeded)
        assert.Error(t, execute("1.1.1.1"))

        // IP 2: Request 1 -> Allow (Different bucket)
        assert.NoError(t, execute("2.2.2.2"))
	})

    t.Run("partition by user", func(t *testing.T) {
        rlMiddleware, _, mockTool := setupMiddleware(configv1.RateLimitConfig_KEY_BY_USER_ID)

        execute := func(user string) error {
            req := &tool.ExecutionRequest{
                ToolName:   "service.test-tool",
                ToolInputs: json.RawMessage(`{}`),
            }
            ctx := context.Background()
            if user != "" {
                ctx = auth.ContextWithUser(ctx, user)
            }
            ctx = tool.NewContextWithTool(ctx, mockTool)

            next := func(_ context.Context, _ *tool.ExecutionRequest) (any, error) {
                return successResult, nil
            }
            _, err := rlMiddleware.Execute(ctx, req, next)
            return err
        }

        // User A: Request 1 -> Allow
        assert.NoError(t, execute("userA"))

        // User A: Request 2 -> Block
        assert.Error(t, execute("userA"))

        // User B: Request 1 -> Allow
        assert.NoError(t, execute("userB"))
	})

    t.Run("partition by api key header", func(t *testing.T) {
        rlMiddleware, _, mockTool := setupMiddleware(configv1.RateLimitConfig_KEY_BY_API_KEY)

        execute := func(apiKey string) error {
            req := &tool.ExecutionRequest{
                ToolName:   "service.test-tool",
                ToolInputs: json.RawMessage(`{}`),
            }
            ctx := context.Background()

            // Inject HTTP request with header
            httpReq, _ := http.NewRequest("POST", "/", nil)
            if apiKey != "" {
                httpReq.Header.Set("X-API-Key", apiKey)
            }
            ctx = context.WithValue(ctx, "http.request", httpReq)

            ctx = tool.NewContextWithTool(ctx, mockTool)

            next := func(_ context.Context, _ *tool.ExecutionRequest) (any, error) {
                return successResult, nil
            }
            _, err := rlMiddleware.Execute(ctx, req, next)
            return err
        }

        // Key A: Request 1 -> Allow
        assert.NoError(t, execute("keyA"))

        // Key A: Request 2 -> Block
        assert.Error(t, execute("keyA"))

        // Key B: Request 1 -> Allow
        assert.NoError(t, execute("keyB"))
	})
}
