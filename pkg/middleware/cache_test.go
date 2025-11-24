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
	"time"

	"github.com/mcpany/core/pkg/middleware"
	"github.com/mcpany/core/pkg/tool"
	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/protobuf/types/known/durationpb"
)

type mockTool struct {
	toolProto *v1.Tool
	mock.Mock
}

func (m *mockTool) Execute(ctx context.Context, req *tool.ExecutionRequest) (any, error) {
	args := m.Called(ctx, req)
	return args.Get(0), args.Error(1)
}

func (m *mockTool) Tool() *v1.Tool {
	return m.toolProto
}

func (m *mockTool) GetCacheConfig() *configv1.CacheConfig {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*configv1.CacheConfig)
}

type mockToolManager struct {
	tool.ToolManagerInterface
	mock.Mock
}

func (m *mockToolManager) GetServiceInfo(serviceID string) (*tool.ServiceInfo, bool) {
	args := m.Called(serviceID)
	return args.Get(0).(*tool.ServiceInfo), args.Bool(1)
}

func TestCachingMiddleware(t *testing.T) {
	t.Run("cache hit", func(t *testing.T) {
		mockToolManager := &mockToolManager{}
		cachingMiddleware := middleware.NewCachingMiddleware(mockToolManager)

		toolProto := &v1.Tool{}
		toolProto.SetServiceId("service")
		mockTool := &mockTool{toolProto: toolProto}

		cacheConfig := &configv1.CacheConfig{}
		cacheConfig.SetIsEnabled(true)
		cacheConfig.SetTtl(durationpb.New(10 * time.Second))
		mockTool.On("GetCacheConfig").Return(cacheConfig)

		req := &tool.ExecutionRequest{
			ToolName:   "service.test-tool",
			ToolInputs: json.RawMessage(`{"input":"value"}`),
		}
		ctx := tool.NewContextWithTool(context.Background(), mockTool)

		// Prime the cache
		cachingMiddleware.Execute(ctx, req, func(ctx context.Context, req *tool.ExecutionRequest) (any, error) {
			return "cached result", nil
		})

		next := func(ctx context.Context, req *tool.ExecutionRequest) (any, error) {
			t.Fatal("next should not be called")
			return nil, nil
		}

		result, err := cachingMiddleware.Execute(ctx, req, next)
		assert.NoError(t, err)
		assert.Equal(t, "cached result", result)
	})

	t.Run("cache miss", func(t *testing.T) {
		mockToolManager := &mockToolManager{}
		cachingMiddleware := middleware.NewCachingMiddleware(mockToolManager)

		toolProto := &v1.Tool{}
		toolProto.SetServiceId("service")
		mockTool := &mockTool{toolProto: toolProto}

		cacheConfig := &configv1.CacheConfig{}
		cacheConfig.SetIsEnabled(true)
		cacheConfig.SetTtl(durationpb.New(10 * time.Second))
		mockTool.On("GetCacheConfig").Return(cacheConfig)

		req := &tool.ExecutionRequest{
			ToolName:   "service.test-tool",
			ToolInputs: json.RawMessage(`{"input":"value"}`),
		}
		ctx := tool.NewContextWithTool(context.Background(), mockTool)

		nextCalled := false
		next := func(ctx context.Context, req *tool.ExecutionRequest) (any, error) {
			nextCalled = true
			return "new result", nil
		}

		result, err := cachingMiddleware.Execute(ctx, req, next)
		assert.NoError(t, err)
		assert.Equal(t, "new result", result)
		assert.True(t, nextCalled)
	})

	t.Run("cache override", func(t *testing.T) {
		mockToolManager := &mockToolManager{}
		cachingMiddleware := middleware.NewCachingMiddleware(mockToolManager)

		toolProto := &v1.Tool{}
		toolProto.SetServiceId("service")
		mockTool := &mockTool{toolProto: toolProto}

		// Service-level cache config
		serviceCacheConfig := &configv1.CacheConfig{}
		serviceCacheConfig.SetIsEnabled(true)
		serviceCacheConfig.SetTtl(durationpb.New(20 * time.Second))
		serviceInfo := &tool.ServiceInfo{
			Config: &configv1.UpstreamServiceConfig{},
		}
		serviceInfo.Config.SetCache(serviceCacheConfig)
		mockToolManager.On("GetServiceInfo", "service").Return(serviceInfo, true)

		// Call-level cache config
		callCacheConfig := &configv1.CacheConfig{}
		callCacheConfig.SetIsEnabled(true)
		callCacheConfig.SetTtl(durationpb.New(5 * time.Second))
		mockTool.On("GetCacheConfig").Return(callCacheConfig)

		req := &tool.ExecutionRequest{
			ToolName:   "service.test-tool",
			ToolInputs: json.RawMessage(`{"input":"value"}`),
		}
		ctx := tool.NewContextWithTool(context.Background(), mockTool)

		// Prime the cache with the call-level TTL
		cachingMiddleware.Execute(ctx, req, func(ctx context.Context, req *tool.ExecutionRequest) (any, error) {
			return "cached result", nil
		})

		// Wait for the call-level cache to expire
		time.Sleep(6 * time.Second)

		nextCalled := false
		next := func(ctx context.Context, req *tool.ExecutionRequest) (any, error) {
			nextCalled = true
			return "new result", nil
		}

		result, err := cachingMiddleware.Execute(ctx, req, next)
		assert.NoError(t, err)
		assert.Equal(t, "new result", result)
		assert.True(t, nextCalled)
	})

	t.Run("cache disabled", func(t *testing.T) {
		mockToolManager := &mockToolManager{}
		cachingMiddleware := middleware.NewCachingMiddleware(mockToolManager)

		toolProto := &v1.Tool{}
		toolProto.SetServiceId("service")
		mockTool := &mockTool{toolProto: toolProto}

		cacheConfig := &configv1.CacheConfig{}
		cacheConfig.SetIsEnabled(false) // Cache is disabled
		mockTool.On("GetCacheConfig").Return(cacheConfig)

		req := &tool.ExecutionRequest{
			ToolName:   "service.test-tool",
			ToolInputs: json.RawMessage(`{"input":"value"}`),
		}
		ctx := tool.NewContextWithTool(context.Background(), mockTool)

		nextCalled := false
		next := func(ctx context.Context, req *tool.ExecutionRequest) (any, error) {
			nextCalled = true
			return "new result", nil
		}

		result, err := cachingMiddleware.Execute(ctx, req, next)
		assert.NoError(t, err)
		assert.Equal(t, "new result", result)
		assert.True(t, nextCalled, "next should be called when cache is disabled")
	})

	t.Run("no tool in context", func(t *testing.T) {
		mockToolManager := &mockToolManager{}
		cachingMiddleware := middleware.NewCachingMiddleware(mockToolManager)

		req := &tool.ExecutionRequest{
			ToolName:   "service.test-tool",
			ToolInputs: json.RawMessage(`{"input":"value"}`),
		}
		ctx := context.Background() // No tool in context

		nextCalled := false
		next := func(ctx context.Context, req *tool.ExecutionRequest) (any, error) {
			nextCalled = true
			return "new result", nil
		}

		result, err := cachingMiddleware.Execute(ctx, req, next)
		assert.NoError(t, err)
		assert.Equal(t, "new result", result)
		assert.True(t, nextCalled, "next should be called when no tool is in context")
	})

	t.Run("next returns error", func(t *testing.T) {
		mockToolManager := &mockToolManager{}
		cachingMiddleware := middleware.NewCachingMiddleware(mockToolManager)

		toolProto := &v1.Tool{}
		toolProto.SetServiceId("service")
		mockTool := &mockTool{toolProto: toolProto}

		cacheConfig := &configv1.CacheConfig{}
		cacheConfig.SetIsEnabled(true)
		cacheConfig.SetTtl(durationpb.New(10 * time.Second))
		mockTool.On("GetCacheConfig").Return(cacheConfig)

		req := &tool.ExecutionRequest{
			ToolName:   "service.test-tool",
			ToolInputs: json.RawMessage(`{"input":"value"}`),
		}
		ctx := tool.NewContextWithTool(context.Background(), mockTool)

		nextCalled := 0
		next := func(ctx context.Context, req *tool.ExecutionRequest) (any, error) {
			nextCalled++
			if nextCalled == 1 {
				return nil, assert.AnError
			}
			return "new result", nil
		}

		// First call, next returns an error
		_, err := cachingMiddleware.Execute(ctx, req, next)
		assert.ErrorIs(t, err, assert.AnError)

		// Second call, should not hit cache
		result, err := cachingMiddleware.Execute(ctx, req, next)
		assert.NoError(t, err)
		assert.Equal(t, "new result", result)
		assert.Equal(t, 2, nextCalled, "next should be called twice")
	})

	t.Run("service info not found", func(t *testing.T) {
		mockToolManager := &mockToolManager{}
		cachingMiddleware := middleware.NewCachingMiddleware(mockToolManager)

		toolProto := &v1.Tool{}
		toolProto.SetServiceId("service")
		mockTool := &mockTool{toolProto: toolProto}

		mockTool.On("GetCacheConfig").Return(nil)
		mockToolManager.On("GetServiceInfo", "service").Return((*tool.ServiceInfo)(nil), false)

		req := &tool.ExecutionRequest{
			ToolName:   "service.test-tool",
			ToolInputs: json.RawMessage(`{"input":"value"}`),
		}
		ctx := tool.NewContextWithTool(context.Background(), mockTool)

		nextCalled := false
		next := func(ctx context.Context, req *tool.ExecutionRequest) (any, error) {
			nextCalled = true
			return "new result", nil
		}

		result, err := cachingMiddleware.Execute(ctx, req, next)
		assert.NoError(t, err)
		assert.Equal(t, "new result", result)
		assert.True(t, nextCalled, "next should be called when service info is not found")
	})

	t.Run("multi-tool cache", func(t *testing.T) {
		mockToolManager := &mockToolManager{}
		cachingMiddleware := middleware.NewCachingMiddleware(mockToolManager)

		// Tool 1 with caching enabled
		tool1Proto := &v1.Tool{}
		tool1Proto.SetServiceId("service1")
		mockTool1 := &mockTool{toolProto: tool1Proto}
		cacheConfig1 := &configv1.CacheConfig{}
		cacheConfig1.SetIsEnabled(true)
		cacheConfig1.SetTtl(durationpb.New(10 * time.Second))
		mockTool1.On("GetCacheConfig").Return(cacheConfig1)

		// Tool 2 with caching disabled
		tool2Proto := &v1.Tool{}
		tool2Proto.SetServiceId("service2")
		mockTool2 := &mockTool{toolProto: tool2Proto}
		cacheConfig2 := &configv1.CacheConfig{}
		cacheConfig2.SetIsEnabled(false)
		mockTool2.On("GetCacheConfig").Return(cacheConfig2)

		req1 := &tool.ExecutionRequest{
			ToolName:   "service1.test-tool1",
			ToolInputs: json.RawMessage(`{"input":"value1"}`),
		}
		ctx1 := tool.NewContextWithTool(context.Background(), mockTool1)

		req2 := &tool.ExecutionRequest{
			ToolName:   "service2.test-tool2",
			ToolInputs: json.RawMessage(`{"input":"value2"}`),
		}
		ctx2 := tool.NewContextWithTool(context.Background(), mockTool2)

		// Prime cache for tool 1
		cachingMiddleware.Execute(ctx1, req1, func(ctx context.Context, req *tool.ExecutionRequest) (any, error) {
			return "cached result1", nil
		})

		// First call for tool 2
		nextCalled2 := 0
		next2 := func(ctx context.Context, req *tool.ExecutionRequest) (any, error) {
			nextCalled2++
			return "new result2", nil
		}
		cachingMiddleware.Execute(ctx2, req2, next2)

		// Verify cache hit for tool 1
		next1 := func(ctx context.Context, req *tool.ExecutionRequest) (any, error) {
			t.Fatal("next should not be called for tool1")
			return nil, nil
		}
		result1, err1 := cachingMiddleware.Execute(ctx1, req1, next1)
		assert.NoError(t, err1)
		assert.Equal(t, "cached result1", result1)

		// Verify cache miss for tool 2
		result2, err2 := cachingMiddleware.Execute(ctx2, req2, next2)
		assert.NoError(t, err2)
		assert.Equal(t, "new result2", result2)
		assert.Equal(t, 2, nextCalled2, "next should be called twice for tool2")
	})
}
