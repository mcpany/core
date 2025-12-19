// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

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

func (m *mockTool) Execute(_ context.Context, _ *tool.ExecutionRequest) (any, error) {
	m.executeCount++
	return successResult, nil
}

func (m *mockTool) GetCacheConfig() *configv1.CacheConfig {
	return m.cacheConfig
}

func TestCachingMiddleware_ExecutionAndCacheHit(t *testing.T) {
	// Setup
	config := &configv1.CacheConfig{
		IsEnabled: proto.Bool(true),
		Ttl:       durationpb.New(100 * time.Millisecond),
	}
	cacheMiddleware := middleware.NewCachingMiddleware(config)

	testTool := &mockTool{
		tool: &v1.Tool{
			Name:      proto.String(testToolName),
			ServiceId: proto.String(testServiceName),
		},
		// No tool-level override
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
	ttl := 50 * time.Millisecond
	config := &configv1.CacheConfig{
		IsEnabled: proto.Bool(true),
		Ttl:       durationpb.New(ttl),
	}
	cacheMiddleware := middleware.NewCachingMiddleware(config)

	testTool := &mockTool{
		tool: &v1.Tool{
			Name:      proto.String(testToolName),
			ServiceId: proto.String(testServiceName),
		},
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
	config := &configv1.CacheConfig{
		IsEnabled: proto.Bool(false), // Cache is explicitly disabled
		Ttl:       durationpb.New(1 * time.Hour),
	}
	cacheMiddleware := middleware.NewCachingMiddleware(config)

	testTool := &mockTool{
		tool: &v1.Tool{
			Name:      proto.String(testToolName),
			ServiceId: proto.String(testServiceName),
		},
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

func TestCachingMiddleware_ToolOverride(t *testing.T) {
    // Service-level cache disabled
    config := &configv1.CacheConfig{IsEnabled: proto.Bool(false)}
    cacheMiddleware := middleware.NewCachingMiddleware(config)

    // Tool-level cache ENABLED
    testTool := &mockTool{
		tool: &v1.Tool{Name: proto.String(testToolName)},
		cacheConfig: &configv1.CacheConfig{
		    IsEnabled: proto.Bool(true),
		    Ttl: durationpb.New(100*time.Millisecond),
		},
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

	// 2. Second call (should be cached)
	_, err = cacheMiddleware.Execute(ctx, req, nextFunc)
	require.NoError(t, err)
	assert.Equal(t, 1, testTool.executeCount)
}

func TestCachingMiddleware_ActionDeleteCache(t *testing.T) {
	// Setup
	config := &configv1.CacheConfig{
		IsEnabled: proto.Bool(true),
		Ttl:       durationpb.New(1 * time.Hour),
	}
	cacheMiddleware := middleware.NewCachingMiddleware(config)

	testTool := &mockTool{
		tool: &v1.Tool{
			Name:      proto.String(testToolName),
			ServiceId: proto.String(testServiceName),
		},
	}
	req := &tool.ExecutionRequest{ToolName: testServiceToolName}
	ctx := tool.NewContextWithTool(context.Background(), testTool)

	nextFunc := func(ctx context.Context, req *tool.ExecutionRequest) (any, error) {
		t, _ := tool.GetFromContext(ctx)
		return t.Execute(ctx, req)
	}

	// 1. First call to populate cache
	_, err := cacheMiddleware.Execute(ctx, req, nextFunc)
	require.NoError(t, err)
	assert.Equal(t, 1, testTool.executeCount)

	// 2. Second call with DELETE_CACHE action
	// Inject CacheControl with Delete Action
	cacheControl := &tool.CacheControl{Action: tool.ActionDeleteCache}
	ctxWithDelete := tool.NewContextWithCacheControl(ctx, cacheControl)

	// This should run the tool AND delete the cache
	_, err = cacheMiddleware.Execute(ctxWithDelete, req, nextFunc)
	require.NoError(t, err)
    // Execution count increases because DELETE_CACHE implies skipping cache lookup (per new logic I added)
	assert.Equal(t, 2, testTool.executeCount)

	// 3. Third call - cache should be empty, so execute again
	// Reset context to normal
	_, err = cacheMiddleware.Execute(ctx, req, nextFunc)
	require.NoError(t, err)
	assert.Equal(t, 3, testTool.executeCount)
}
