package middleware_test

import (
	"context"
	"testing"
	"time"

	"github.com/mcpany/core/pkg/middleware"
	"github.com/mcpany/core/pkg/tool"
	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/modelcontextprotocol/go-sdk/mcp"
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

func (m *mockTool) MCPTool() *mcp.Tool {
	t, _ := tool.ConvertProtoToMCPTool(m.tool)
	return t
}

// mockToolManager is a mock implementation of the tool.ManagerInterface.
type mockToolManager struct{}

func (m *mockToolManager) GetServiceInfo(_ string) (*tool.ServiceInfo, bool) {
	return &tool.ServiceInfo{
		Config: &configv1.UpstreamServiceConfig{},
	}, true
}
func (m *mockToolManager) AddTool(_ tool.Tool) error                { return nil }
func (m *mockToolManager) GetTool(_ string) (tool.Tool, bool)       { return nil, false }
func (m *mockToolManager) ListTools() []tool.Tool                   { return nil }
func (m *mockToolManager) ListServices() []*tool.ServiceInfo        { return nil }
func (m *mockToolManager) AddMiddleware(_ tool.ExecutionMiddleware) {}
func (m *mockToolManager) ExecuteTool(_ context.Context, _ *tool.ExecutionRequest) (interface{}, error) {
	return nil, nil
}
func (m *mockToolManager) SetMCPServer(_ tool.MCPServerProvider)        {}
func (m *mockToolManager) AddServiceInfo(_ string, _ *tool.ServiceInfo) {}
func (m *mockToolManager) SetProfiles(_ []string, _ []*configv1.ProfileDefinition) {}
func (m *mockToolManager) ClearToolsForService(_ string)                {}

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

func TestCachingMiddleware_ServiceInfoNotFound(t *testing.T) {
	// Setup
	tm := &mockToolManager{}
	cacheMiddleware := middleware.NewCachingMiddleware(tm)

	// Tool belonging to a service that is NOT known to the tool manager
	testTool := &mockTool{
		tool: v1.Tool_builder{
			Name:      proto.String(testToolName),
			ServiceId: proto.String("unknown-service"),
		}.Build(),
		cacheConfig: nil,
	}
	req := &tool.ExecutionRequest{ToolName: "unknown-service.test-tool"}
	ctx := tool.NewContextWithTool(context.Background(), testTool)
	nextFunc := func(ctx context.Context, req *tool.ExecutionRequest) (any, error) {
		t, _ := tool.GetFromContext(ctx)
		return t.Execute(ctx, req)
	}

	// Should proceed without caching because service info (and thus cache config) is missing
	res, err := cacheMiddleware.Execute(ctx, req, nextFunc)
	require.NoError(t, err)
	assert.Equal(t, successResult, res)
	assert.Equal(t, 1, testTool.executeCount)
}

func TestCachingMiddleware_ActionDeleteCache(t *testing.T) {
	// Setup
	tm := &mockToolManager{}
	cacheMiddleware := middleware.NewCachingMiddleware(tm)

	testTool := &mockTool{
		tool: v1.Tool_builder{
			Name:      proto.String(testToolName),
			ServiceId: proto.String(testServiceName),
		}.Build(),
		cacheConfig: configv1.CacheConfig_builder{
			IsEnabled: proto.Bool(true),
			Ttl:       durationpb.New(1 * time.Hour),
		}.Build(),
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
	// We expect ActionDeleteCache to SKIP cache lookup and force execution.
	res, err := cacheMiddleware.Execute(ctxWithDelete, req, nextFunc)
	require.NoError(t, err)
	assert.Equal(t, successResult, res)
	assert.Equal(t, 2, testTool.executeCount, "Tool should be executed again when ActionDeleteCache is used")

	// WAIT. If DeleteCache is set, we probably want to FORCE execution?
	// Current middleware logical flow:
	// 1. Check Cache.
	// 2. If hit, return.
	// 3. If miss, run next.
	// 4. If DeleteCache, delete.

	// If the user sets `DELETE_CACHE`, they probably intend "Run this tool and remove the old cache entry".
	// They might expect it to run fresh.
	// BUT per my implementation, if there is a cache entry, it returns it!
	// And then returns result.
	// `ActionDeleteCache` check is at step 4 (after execution).
	// If step 2 returns, step 4 is NOT reached!
	// This means `DELETE_CACHE` action is IGNORED if cache hit!

	// If `ActionDeleteCache` is present, we should probably SKIP cache lookup?
	// Or proceed to delete AFTER returning cached value? (Doesn't make sense to delete if we just used it).
	// "User converts parameter transformer to webhook based system... add SAVE_CACHE and DELETE_CACHE actions".
	// Usually DELETE means INVALIDATE.
	// If INVALIDATE, we should verify invalidation.
	// If I want to invalidate, I might call the tool?
	// If I want to invalidate WITHOUT execution, that's different. But this is a Call Policy on a Call.
	// If `DELETE_CACHE` is returned, we should probably SKIP CACHE and then DELETE IT (to ensure freshness next time)?
	// Or maybe "Execute, then delete"? meaning 1-time execution that clears cache?
	// If I want to force refresh, I would use `DELETE_CACHE`?
	// If so, I should SKIP cache check.

	// I will update Validated Logic in CacheMiddleware:
	// If Action == DeleteCache, SKIP cache check.
}
