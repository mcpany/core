// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware_test

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/mcpany/core/server/pkg/auth"
	"github.com/mcpany/core/server/pkg/middleware"
	"github.com/mcpany/core/server/pkg/tool"
	"github.com/mcpany/core/server/pkg/util"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

func TestRateLimitPartitioning(t *testing.T) {
	const successResult = "success"

	setupMiddleware := func(keyBy configv1.RateLimitConfig_KeyBy) (*middleware.RateLimitMiddleware, *rateLimitMockToolManager) {
		mockToolManager := &rateLimitMockToolManager{}
		rlMiddleware := middleware.NewRateLimitMiddleware(mockToolManager)

		toolProto := v1.Tool_builder{
			ServiceId: proto.String("service"),
		}.Build()
		mockTool := &rateLimitMockTool{toolProto: toolProto}

		rlConfig := configv1.RateLimitConfig_builder{
			IsEnabled:         true,
			RequestsPerSecond: 1,
			Burst:             1,
			KeyBy:             keyBy,
		}.Build()

		serviceInfo := &tool.ServiceInfo{
			Name: "test-service",
			Config: configv1.UpstreamServiceConfig_builder{
				RateLimit: rlConfig,
			}.Build(),
		}

		// Allow any number of calls
		mockToolManager.On("GetTool", "service.test-tool").Return(mockTool, true)
		mockToolManager.On("GetServiceInfo", "service").Return(serviceInfo, true)

		return rlMiddleware, mockToolManager
	}

	mockNext := func(_ context.Context, _ *tool.ExecutionRequest) (any, error) {
		return successResult, nil
	}

	req := &tool.ExecutionRequest{
		ToolName:   "service.test-tool",
		ToolInputs: json.RawMessage(`{}`),
	}

	toolProto := v1.Tool_builder{ServiceId: proto.String("service")}.Build()
	mockTool := &rateLimitMockTool{toolProto: toolProto}
	baseCtx := tool.NewContextWithTool(context.Background(), mockTool)

	t.Run("partition by IP", func(t *testing.T) {
		rlMiddleware, _ := setupMiddleware(configv1.RateLimitConfig_KEY_BY_IP)

		// Client A (IP 1.2.3.4)
		ctxA := util.ContextWithRemoteIP(baseCtx, "1.2.3.4")

		// 1. Success
		res, err := rlMiddleware.Execute(ctxA, req, mockNext)
		assert.NoError(t, err)
		assert.Equal(t, successResult, res)

		// 2. Fail (Limit exceeded for A)
		_, err = rlMiddleware.Execute(ctxA, req, mockNext)
		assert.Error(t, err)

		// Client B (IP 5.6.7.8)
		ctxB := util.ContextWithRemoteIP(baseCtx, "5.6.7.8")

		// 3. Success (Should not be affected by A)
		res, err = rlMiddleware.Execute(ctxB, req, mockNext)
		assert.NoError(t, err)
		assert.Equal(t, successResult, res)
	})

	t.Run("partition by User ID", func(t *testing.T) {
		rlMiddleware, _ := setupMiddleware(configv1.RateLimitConfig_KEY_BY_USER_ID)

		// User A
		ctxA := auth.ContextWithUser(baseCtx, "user-a")

		// 1. Success
		res, err := rlMiddleware.Execute(ctxA, req, mockNext)
		assert.NoError(t, err)
		assert.Equal(t, successResult, res)

		// 2. Fail
		_, err = rlMiddleware.Execute(ctxA, req, mockNext)
		assert.Error(t, err)

		// User B
		ctxB := auth.ContextWithUser(baseCtx, "user-b")

		// 3. Success
		res, err = rlMiddleware.Execute(ctxB, req, mockNext)
		assert.NoError(t, err)
		assert.Equal(t, successResult, res)
	})

	t.Run("partition by API Key (Context)", func(t *testing.T) {
		rlMiddleware, _ := setupMiddleware(configv1.RateLimitConfig_KEY_BY_API_KEY)

		// Key A
		ctxA := auth.ContextWithAPIKey(baseCtx, "key-a")

		// 1. Success
		res, err := rlMiddleware.Execute(ctxA, req, mockNext)
		assert.NoError(t, err)
		assert.Equal(t, successResult, res)

		// 2. Fail
		_, err = rlMiddleware.Execute(ctxA, req, mockNext)
		assert.Error(t, err)

		// Key B
		ctxB := auth.ContextWithAPIKey(baseCtx, "key-b")

		// 3. Success
		res, err = rlMiddleware.Execute(ctxB, req, mockNext)
		assert.NoError(t, err)
		assert.Equal(t, successResult, res)
	})

	t.Run("partition by API Key (Header X-API-Key)", func(t *testing.T) {
		rlMiddleware, _ := setupMiddleware(configv1.RateLimitConfig_KEY_BY_API_KEY)

		// Request A
		httpReqA, _ := http.NewRequest("POST", "http://example.com", nil)
		httpReqA.Header.Set("X-API-Key", "key-header-a")
		ctxA := context.WithValue(baseCtx, middleware.HTTPRequestContextKey, httpReqA)

		// 1. Success
		res, err := rlMiddleware.Execute(ctxA, req, mockNext)
		assert.NoError(t, err)
		assert.Equal(t, successResult, res)

		// 2. Fail
		_, err = rlMiddleware.Execute(ctxA, req, mockNext)
		assert.Error(t, err)

		// Request B
		httpReqB, _ := http.NewRequest("POST", "http://example.com", nil)
		httpReqB.Header.Set("X-API-Key", "key-header-b")
		ctxB := context.WithValue(baseCtx, middleware.HTTPRequestContextKey, httpReqB)

		// 3. Success
		res, err = rlMiddleware.Execute(ctxB, req, mockNext)
		assert.NoError(t, err)
		assert.Equal(t, successResult, res)
	})

	t.Run("partition by API Key (Header Authorization)", func(t *testing.T) {
		rlMiddleware, _ := setupMiddleware(configv1.RateLimitConfig_KEY_BY_API_KEY)

		// Request A
		httpReqA, _ := http.NewRequest("POST", "http://example.com", nil)
		httpReqA.Header.Set("Authorization", "Bearer token-a")
		ctxA := context.WithValue(baseCtx, middleware.HTTPRequestContextKey, httpReqA)

		// 1. Success
		res, err := rlMiddleware.Execute(ctxA, req, mockNext)
		assert.NoError(t, err)
		assert.Equal(t, successResult, res)

		// 2. Fail
		_, err = rlMiddleware.Execute(ctxA, req, mockNext)
		assert.Error(t, err)

		// Request B
		httpReqB, _ := http.NewRequest("POST", "http://example.com", nil)
		httpReqB.Header.Set("Authorization", "Bearer token-b")
		ctxB := context.WithValue(baseCtx, middleware.HTTPRequestContextKey, httpReqB)

		// 3. Success
		res, err = rlMiddleware.Execute(ctxB, req, mockNext)
		assert.NoError(t, err)
		assert.Equal(t, successResult, res)
	})

	t.Run("partition fallback (unknown)", func(t *testing.T) {
		// Test unknown IP
		rlMiddleware, _ := setupMiddleware(configv1.RateLimitConfig_KEY_BY_IP)

		// Context without IP
		// Should use "ip:unknown"

		// 1. Success
		res, err := rlMiddleware.Execute(baseCtx, req, mockNext)
		assert.NoError(t, err)
		assert.Equal(t, successResult, res)

		// 2. Fail (Limit exceeded for unknown)
		_, err = rlMiddleware.Execute(baseCtx, req, mockNext)
		assert.Error(t, err)
	})
}
