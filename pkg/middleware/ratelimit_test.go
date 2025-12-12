/*
 * Copyright 2025 Author(s) of MCP Any
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package middleware_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/mcpany/core/pkg/middleware"
	"github.com/mcpany/core/pkg/tool"
	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type rateLimitMockTool struct {
	toolProto *v1.Tool
	mock.Mock
}

func (m *rateLimitMockTool) Execute(ctx context.Context, req *tool.ExecutionRequest) (any, error) {
	args := m.Called(ctx, req)
	return args.Get(0), args.Error(1)
}

func (m *rateLimitMockTool) Tool() *v1.Tool {
	return m.toolProto
}

func (m *rateLimitMockTool) GetCacheConfig() *configv1.CacheConfig {
	return nil
}

type rateLimitMockToolManager struct {
	tool.ToolManagerInterface
	mock.Mock
}

func (m *rateLimitMockToolManager) GetServiceInfo(serviceID string) (*tool.ServiceInfo, bool) {
	args := m.Called(serviceID)
	if args.Get(0) == nil {
		return nil, args.Bool(1)
	}
	return args.Get(0).(*tool.ServiceInfo), args.Bool(1)
}

func TestRateLimitMiddleware(t *testing.T) {
	t.Run("rate limit allowed", func(t *testing.T) {
		mockToolManager := &rateLimitMockToolManager{}
		rlMiddleware := middleware.NewRateLimitMiddleware(mockToolManager)

		toolProto := &v1.Tool{}
		toolProto.SetServiceId("service")
		mockTool := &rateLimitMockTool{toolProto: toolProto}

		rlConfig := &configv1.RateLimitConfig{}
		rlConfig.SetIsEnabled(true)
		rlConfig.SetRequestsPerSecond(10)
		rlConfig.SetBurst(1)

		serviceInfo := &tool.ServiceInfo{
			Name:   "test-service",
			Config: &configv1.UpstreamServiceConfig{},
		}
		serviceInfo.Config.SetRateLimit(rlConfig)
		mockToolManager.On("GetServiceInfo", "service").Return(serviceInfo, true)

		req := &tool.ExecutionRequest{
			ToolName:   "service.test-tool",
			ToolInputs: json.RawMessage(`{}`),
		}
		ctx := tool.NewContextWithTool(context.Background(), mockTool)

		nextCalled := false
		next := func(ctx context.Context, req *tool.ExecutionRequest) (any, error) {
			nextCalled = true
			return "success", nil
		}

		result, err := rlMiddleware.Execute(ctx, req, next)
		assert.NoError(t, err)
		assert.Equal(t, "success", result)
		assert.True(t, nextCalled)
	})

	t.Run("rate limit exceeded", func(t *testing.T) {
		mockToolManager := &rateLimitMockToolManager{}
		rlMiddleware := middleware.NewRateLimitMiddleware(mockToolManager)

		toolProto := &v1.Tool{}
		toolProto.SetServiceId("service")
		mockTool := &rateLimitMockTool{toolProto: toolProto}

		rlConfig := &configv1.RateLimitConfig{}
		rlConfig.SetIsEnabled(true)
		rlConfig.SetRequestsPerSecond(1) // 1 RPS
		rlConfig.SetBurst(1)

		serviceInfo := &tool.ServiceInfo{
			Name:   "test-service",
			Config: &configv1.UpstreamServiceConfig{},
		}
		serviceInfo.Config.SetRateLimit(rlConfig)
		mockToolManager.On("GetServiceInfo", "service").Return(serviceInfo, true)

		req := &tool.ExecutionRequest{
			ToolName:   "service.test-tool",
			ToolInputs: json.RawMessage(`{}`),
		}
		ctx := tool.NewContextWithTool(context.Background(), mockTool)

		next := func(ctx context.Context, req *tool.ExecutionRequest) (any, error) {
			return "success", nil
		}

		// First request should succeed
		result, err := rlMiddleware.Execute(ctx, req, next)
		assert.NoError(t, err)
		assert.Equal(t, "success", result)

		// Immediate second request should fail
		_, err = rlMiddleware.Execute(ctx, req, next)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "rate limit exceeded")
	})

	t.Run("rate limit config update", func(t *testing.T) {
		mockToolManager := &rateLimitMockToolManager{}
		rlMiddleware := middleware.NewRateLimitMiddleware(mockToolManager)

		toolProto := &v1.Tool{}
		toolProto.SetServiceId("service")
		mockTool := &rateLimitMockTool{toolProto: toolProto}

		// Initial Config: 10 RPS, Burst 10
		rlConfig1 := &configv1.RateLimitConfig{}
		rlConfig1.SetIsEnabled(true)
		rlConfig1.SetRequestsPerSecond(10)
		rlConfig1.SetBurst(10)

		serviceInfo1 := &tool.ServiceInfo{
			Name:   "test-service",
			Config: &configv1.UpstreamServiceConfig{},
		}
		serviceInfo1.Config.SetRateLimit(rlConfig1)

		// Updated Config: 1 RPS, Burst 1
		rlConfig2 := &configv1.RateLimitConfig{}
		rlConfig2.SetIsEnabled(true)
		rlConfig2.SetRequestsPerSecond(1)
		rlConfig2.SetBurst(1)

		serviceInfo2 := &tool.ServiceInfo{
			Name:   "test-service",
			Config: &configv1.UpstreamServiceConfig{},
		}
		serviceInfo2.Config.SetRateLimit(rlConfig2)

		// Mock sequence:
		mockToolManager.On("GetServiceInfo", "service").Return(serviceInfo1, true).Once()
		mockToolManager.On("GetServiceInfo", "service").Return(serviceInfo2, true).Twice()

		req := &tool.ExecutionRequest{
			ToolName:   "service.test-tool",
			ToolInputs: json.RawMessage(`{}`),
		}
		ctx := tool.NewContextWithTool(context.Background(), mockTool)

		next := func(ctx context.Context, req *tool.ExecutionRequest) (any, error) {
			return "success", nil
		}

		// 1. Allowed (Config 1: 10 RPS)
		result, err := rlMiddleware.Execute(ctx, req, next)
		assert.NoError(t, err)
		assert.Equal(t, "success", result)

		// 2. Allowed (Config 2: 1 RPS, Burst 1)
		// Since we updated burst to 1, and we have tokens (from 10), but burst cap reduces tokens to 1.
		// Wait, if I had 9 tokens. New Burst 1. Tokens -> 1.
		// Allow() consumes 1. Tokens -> 0. Success.
		result, err = rlMiddleware.Execute(ctx, req, next)
		assert.NoError(t, err)
		assert.Equal(t, "success", result)

		// 3. Blocked (Config 2: 1 RPS, Burst 1, Tokens 0)
		_, err = rlMiddleware.Execute(ctx, req, next)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "rate limit exceeded")
	})

	t.Run("rate limit disabled", func(t *testing.T) {
		mockToolManager := &rateLimitMockToolManager{}
		rlMiddleware := middleware.NewRateLimitMiddleware(mockToolManager)

		toolProto := &v1.Tool{}
		toolProto.SetServiceId("service")
		mockTool := &rateLimitMockTool{toolProto: toolProto}

		rlConfig := &configv1.RateLimitConfig{}
		rlConfig.SetIsEnabled(false)

		serviceInfo := &tool.ServiceInfo{
			Name:   "test-service",
			Config: &configv1.UpstreamServiceConfig{},
		}
		serviceInfo.Config.SetRateLimit(rlConfig)
		mockToolManager.On("GetServiceInfo", "service").Return(serviceInfo, true)

		req := &tool.ExecutionRequest{
			ToolName:   "service.test-tool",
			ToolInputs: json.RawMessage(`{}`),
		}
		ctx := tool.NewContextWithTool(context.Background(), mockTool)

		nextCalled := false
		next := func(ctx context.Context, req *tool.ExecutionRequest) (any, error) {
			nextCalled = true
			return "success", nil
		}

		result, err := rlMiddleware.Execute(ctx, req, next)
		assert.NoError(t, err)
		assert.Equal(t, "success", result)
		assert.True(t, nextCalled)
	})

	t.Run("service info not found", func(t *testing.T) {
		mockToolManager := &rateLimitMockToolManager{}
		rlMiddleware := middleware.NewRateLimitMiddleware(mockToolManager)

		toolProto := &v1.Tool{}
		toolProto.SetServiceId("service")
		mockTool := &rateLimitMockTool{toolProto: toolProto}

		mockToolManager.On("GetServiceInfo", "service").Return(nil, false)

		req := &tool.ExecutionRequest{
			ToolName:   "service.test-tool",
			ToolInputs: json.RawMessage(`{}`),
		}
		ctx := tool.NewContextWithTool(context.Background(), mockTool)

		nextCalled := false
		next := func(ctx context.Context, req *tool.ExecutionRequest) (any, error) {
			nextCalled = true
			return "success", nil
		}

		result, err := rlMiddleware.Execute(ctx, req, next)
		assert.NoError(t, err)
		assert.Equal(t, "success", result)
		assert.True(t, nextCalled)
	})
}
