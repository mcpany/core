/*
 * Copyright 2025 Author(s) of MCP-XY
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

	"github.com/mcpxy/core/pkg/common/clock"
	"github.com/mcpxy/core/pkg/middleware"
	"github.com/mcpxy/core/pkg/tool"
	configv1 "github.com/mcpxy/core/proto/config/v1"
	v1 "github.com/mcpxy/core/proto/mcp_router/v1"
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
		cachingMiddleware := middleware.NewCachingMiddleware(mockToolManager, clock.New())

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
		cachingMiddleware := middleware.NewCachingMiddleware(mockToolManager, clock.New())

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
		fakeClock := clock.NewFake(time.Now())
		cachingMiddleware := middleware.NewCachingMiddleware(mockToolManager, fakeClock)

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
		fakeClock.Advance(6 * time.Second)

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

	t.Run("cache key consistency", func(t *testing.T) {
		req1 := &tool.ExecutionRequest{
			ToolName:   "test-tool",
			ToolInputs: json.RawMessage(`{"a":"1", "b":"2"}`),
		}
		req2 := &tool.ExecutionRequest{
			ToolName:   "test-tool",
			ToolInputs: json.RawMessage(`{"b":"2", "a":"1"}`),
		}

		cachingMiddleware := middleware.NewCachingMiddleware(nil, clock.New())
		key1 := cachingMiddleware.GetCacheKey(req1)
		key2 := cachingMiddleware.GetCacheKey(req2)

		assert.Equal(t, key1, key2, "Cache keys should be equal")
	})
}
