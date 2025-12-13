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
	"testing"
	"time"

	"github.com/mcpany/core/pkg/middleware"
	"github.com/mcpany/core/pkg/tool"
	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/durationpb"
)

const (
	testToolName        = "test-tool"
	testServiceName     = "test-service"
	testServiceToolName = "test-service.test-tool"
	successResult       = "success"
)

// mockTool is a mock implementation of the tool.Tool interface for testing.
type mockTool struct {
	tool         *v1.Tool
	executeCount int
	cacheConfig  *configv1.CacheConfig
}

func (m *mockTool) Tool() *v1.Tool {
	return m.tool
}

func (m *mockTool) Execute(ctx context.Context, _ *tool.ExecutionRequest) (any, error) {
	m.executeCount++
	return successResult, nil
}

func (m *mockTool) GetCacheConfig() *configv1.CacheConfig {
	return m.cacheConfig
}

// mockToolManager is a mock implementation of the tool.ManagerInterface.
type mockToolManager struct{}

func (m *mockToolManager) GetServiceInfo(serviceID string) (*tool.ServiceInfo, bool) {
	return &tool.ServiceInfo{
		Config: &configv1.UpstreamServiceConfig{},
	}, true
}
func (m *mockToolManager) AddTool(_ tool.Tool) error                             { return nil }
func (m *mockToolManager) GetTool(_ string) (tool.Tool, bool)             { return nil, false }
func (m *mockToolManager) ListTools() []tool.Tool                                { return nil }
func (m *mockToolManager) AddMiddleware(_ tool.ExecutionMiddleware) {}
func (m *mockToolManager) ExecuteTool(_ context.Context, _ *tool.ExecutionRequest) (interface{}, error) {
	return nil, nil
}
func (m *mockToolManager) SetMCPServer(_ tool.MCPServerProvider)           {}
func (m *mockToolManager) AddServiceInfo(_ string, _ *tool.ServiceInfo) {}
func (m *mockToolManager) ClearToolsForService(_ string)                  {}

func TestCachingMiddleware_ExecutionAndCacheHit(t *testing.T) {
	// Setup
	tm := &mockToolManager{}
	cacheMiddleware := middleware.NewCachingMiddleware(tm)

	testTool := &mockTool{
		tool: v1.Tool_builder{
			Name:      proto.String("test-tool"),
			ServiceId: proto.String("test-service"),
		}.Build(),
		cacheConfig: configv1.CacheConfig_builder{
			IsEnabled: proto.Bool(true),
			Ttl:       durationpb.New(100 * time.Millisecond),
		}.Build(),
	}

	req := &tool.ExecutionRequest{
		ToolName: testServiceToolName,
	}

	// Create a context with the tool
	ctx := tool.NewContextWithTool(context.Background(), testTool)

	// Define the "next" function in the middleware chain
	nextFunc := func(ctx context.Context, req *tool.ExecutionRequest) (any, error) {
		t, _ := tool.GetFromContext(ctx)
		return t.Execute(ctx, req)
	}

	// 1. First call - should execute the tool
	res, err := cacheMiddleware.Execute(ctx, req, nextFunc)
	require.NoError(t, err)
	assert.Equal(t, successResult, res)
	assert.Equal(t, 1, testTool.executeCount, "Tool should have been executed on the first call")

	// 2. Second call - should hit the cache
	res, err = cacheMiddleware.Execute(ctx, req, nextFunc)
	require.NoError(t, err)
	assert.Equal(t, successResult, res)
	assert.Equal(t, 1, testTool.executeCount, "Tool should not have been executed again; result should come from cache")
}

func TestCachingMiddleware_CacheExpiration(t *testing.T) {
	// Setup
	tm := &mockToolManager{}
	cacheMiddleware := middleware.NewCachingMiddleware(tm)
	ttl := 50 * time.Millisecond

	testTool := &mockTool{
		tool: v1.Tool_builder{
			Name:      proto.String(testToolName),
			ServiceId: proto.String(testServiceName),
		}.Build(),
		cacheConfig: configv1.CacheConfig_builder{
			IsEnabled: proto.Bool(true),
			Ttl:       durationpb.New(ttl),
		}.Build(),
	}
	req := &tool.ExecutionRequest{ToolName: "test-service.test-tool"}
	ctx := tool.NewContextWithTool(context.Background(), testTool)
	nextFunc := func(ctx context.Context, req *tool.ExecutionRequest) (any, error) {
		t, _ := tool.GetFromContext(ctx)
		return t.Execute(ctx, req)
	}

	// 1. First call to populate cache
	_, err := cacheMiddleware.Execute(ctx, req, nextFunc)
	require.NoError(t, err)
	require.Equal(t, 1, testTool.executeCount)

	// 2. Wait for the cache to expire
	time.Sleep(ttl + 10*time.Millisecond)

	// 3. Third call - should execute the tool again
	res, err := cacheMiddleware.Execute(ctx, req, nextFunc)
	require.NoError(t, err)
	assert.Equal(t, successResult, res)
	assert.Equal(t, 2, testTool.executeCount, "Tool should be executed again after cache expiry")
}

func TestCachingMiddleware_CacheDisabled(t *testing.T) {
	// Setup
	tm := &mockToolManager{}
	cacheMiddleware := middleware.NewCachingMiddleware(tm)
	testTool := &mockTool{
		tool: v1.Tool_builder{
			Name:      proto.String(testToolName),
			ServiceId: proto.String(testServiceName),
		}.Build(),
		cacheConfig: configv1.CacheConfig_builder{
			IsEnabled: proto.Bool(false), // Cache is explicitly disabled
			Ttl:       durationpb.New(1 * time.Hour),
		}.Build(),
	}
	req := &tool.ExecutionRequest{ToolName: "test-service.test-tool"}
	ctx := tool.NewContextWithTool(context.Background(), testTool)
	nextFunc := func(ctx context.Context, req *tool.ExecutionRequest) (any, error) {
		t, _ := tool.GetFromContext(ctx)
		return t.Execute(ctx, req)
	}

	// 1. First call
	_, err := cacheMiddleware.Execute(ctx, req, nextFunc)
	require.NoError(t, err)
	assert.Equal(t, 1, testTool.executeCount)

	// 2. Second call
	_, err = cacheMiddleware.Execute(ctx, req, nextFunc)
	require.NoError(t, err)
	assert.Equal(t, 2, testTool.executeCount, "Tool should be executed every time when cache is disabled")
}

func TestCachingMiddleware_NoCacheConfig(t *testing.T) {
	// Setup
	tm := &mockToolManager{}
	cacheMiddleware := middleware.NewCachingMiddleware(tm)
	testTool := &mockTool{
		tool: v1.Tool_builder{
			Name:      proto.String(testToolName),
			ServiceId: proto.String(testServiceName),
		}.Build(),
		cacheConfig: nil, // No cache config provided for the tool
	}
	req := &tool.ExecutionRequest{ToolName: testServiceToolName}
	ctx := tool.NewContextWithTool(context.Background(), testTool)
	nextFunc := func(ctx context.Context, req *tool.ExecutionRequest) (any, error) {
		t, _ := tool.GetFromContext(ctx)
		return t.Execute(ctx, req)
	}

	// 1. First call
	_, err := cacheMiddleware.Execute(ctx, req, nextFunc)
	require.NoError(t, err)
	assert.Equal(t, 1, testTool.executeCount)

	// 2. Second call
	_, err = cacheMiddleware.Execute(ctx, req, nextFunc)
	require.NoError(t, err)
	assert.Equal(t, 2, testTool.executeCount, "Tool should be executed every time when there is no cache config")
}
