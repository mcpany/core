// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware_test

import (
	"context"
	"testing"
	"time"

	"github.com/mcpany/core/server/pkg/middleware"
	"github.com/mcpany/core/server/pkg/tool"
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
func (m *mockToolManager) SetMCPServer(_ tool.MCPServerProvider)                   {}
func (m *mockToolManager) AddServiceInfo(_ string, _ *tool.ServiceInfo)            {}
func (m *mockToolManager) SetProfiles(_ []string, _ []*configv1.ProfileDefinition) {}
func (m *mockToolManager) ClearToolsForService(_ string)                           {}

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
}

func TestCachingMiddleware_DeterministicKeys(t *testing.T) {
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

	// Two requests with same content but different key order
	req1 := &tool.ExecutionRequest{
		ToolName:   testServiceToolName,
		ToolInputs: []byte(`{"a": 1, "b": 2}`),
	}
	req2 := &tool.ExecutionRequest{
		ToolName:   testServiceToolName,
		ToolInputs: []byte(`{"b": 2, "a": 1}`),
	}

	ctx := tool.NewContextWithTool(context.Background(), testTool)
	nextFunc := func(ctx context.Context, req *tool.ExecutionRequest) (any, error) {
		t, _ := tool.GetFromContext(ctx)
		return t.Execute(ctx, req)
	}

	// 1. Call with req1 - miss, execute
	res1, err := cacheMiddleware.Execute(ctx, req1, nextFunc)
	require.NoError(t, err)
	assert.Equal(t, successResult, res1)
	assert.Equal(t, 1, testTool.executeCount)

	// 2. Call with req2 - should hit cache
	res2, err := cacheMiddleware.Execute(ctx, req2, nextFunc)
	require.NoError(t, err)
	assert.Equal(t, successResult, res2)
	assert.Equal(t, 1, testTool.executeCount, "Should be cache hit despite different key order")
}

func TestCachingMiddleware_Clear(t *testing.T) {
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

	// 1. Populate cache
	_, err := cacheMiddleware.Execute(ctx, req, nextFunc)
	require.NoError(t, err)
	assert.Equal(t, 1, testTool.executeCount)

	// 2. Clear cache
	err = cacheMiddleware.Clear(context.Background())
	require.NoError(t, err)

	// 3. Call again - should execute (miss)
	_, err = cacheMiddleware.Execute(ctx, req, nextFunc)
	require.NoError(t, err)
	assert.Equal(t, 2, testTool.executeCount, "Tool should be executed again after cache clear")
}

func TestCachingMiddleware_ActionDeleteCache_VerifyDeletion(t *testing.T) {
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

	// 1. Populate cache
	_, err := cacheMiddleware.Execute(ctx, req, nextFunc)
	require.NoError(t, err)
	assert.Equal(t, 1, testTool.executeCount)

	// 2. Call with DeleteCache
	cacheControl := &tool.CacheControl{Action: tool.ActionDeleteCache}
	ctxWithDelete := tool.NewContextWithCacheControl(ctx, cacheControl)

	_, err = cacheMiddleware.Execute(ctxWithDelete, req, nextFunc)
	require.NoError(t, err)
	assert.Equal(t, 2, testTool.executeCount, "Should execute due to skip cache")

	// 3. Call again with Normal Allow
	// If cache was deleted in step 2, this should be a MISS -> Execute -> count=3
	_, err = cacheMiddleware.Execute(ctx, req, nextFunc)
	require.NoError(t, err)
	assert.Equal(t, 3, testTool.executeCount, "Should execute again because cache was deleted and not repopulated in step 2")
}

// MockProviderFactory mocks the EmbeddingProvider creation.
type MockProviderFactory struct {
	embeddings map[string][]float32
}

func (m *MockProviderFactory) Create(_ *configv1.SemanticCacheConfig, _ string) (middleware.EmbeddingProvider, error) {
	return &MockEmbeddingProvider{embeddings: m.embeddings}, nil
}

type MockEmbeddingProvider struct {
	embeddings map[string][]float32
}

func (m *MockEmbeddingProvider) Embed(ctx context.Context, text string) ([]float32, error) {
	if val, ok := m.embeddings[text]; ok {
		return val, nil
	}
	return []float32{0, 0, 0}, nil
}

func TestCachingMiddleware_SemanticCache(t *testing.T) {
	// Setup
	tm := &mockToolManager{}
	cacheMiddleware := middleware.NewCachingMiddleware(tm)

	// Override factory
	mockFactory := &MockProviderFactory{
		embeddings: map[string][]float32{
			"hello": {1.0, 0.0, 0.0},
			"hi":    {0.99, 0.05, 0.0},
		},
	}
	cacheMiddleware.SetProviderFactory(mockFactory.Create)

	testTool := &mockTool{
		tool: v1.Tool_builder{
			Name:      proto.String(testToolName),
			ServiceId: proto.String(testServiceName),
		}.Build(),
		cacheConfig: configv1.CacheConfig_builder{
			IsEnabled: proto.Bool(true),
			Strategy:  proto.String("semantic"),
			SemanticConfig: configv1.SemanticCacheConfig_builder{
				Provider: proto.String("openai"),
				ApiKey: configv1.SecretValue_builder{
					PlainText: proto.String("test-api-key"),
				}.Build(),
				Model:               proto.String("test-model"),
				SimilarityThreshold: proto.Float32(0.9),
			}.Build(),
			Ttl: durationpb.New(1 * time.Hour),
		}.Build(),
	}

	req := &tool.ExecutionRequest{
		ToolName:   testServiceToolName,
		ToolInputs: []byte("hello"),
	}

	ctx := tool.NewContextWithTool(context.Background(), testTool)
	nextFunc := func(ctx context.Context, req *tool.ExecutionRequest) (any, error) {
		t, _ := tool.GetFromContext(ctx)
		return t.Execute(ctx, req)
	}

	// 1. First call - should execute (miss, but sets cache)
	res1, err := cacheMiddleware.Execute(ctx, req, nextFunc)
	require.NoError(t, err)
	assert.Equal(t, successResult, res1)
	assert.Equal(t, 1, testTool.executeCount)

	// 2. Second call with "hi" (should match "hello")
	req2 := &tool.ExecutionRequest{
		ToolName:   testServiceToolName,
		ToolInputs: []byte("hi"),
	}
	res2, err := cacheMiddleware.Execute(ctx, req2, nextFunc)
	require.NoError(t, err)
	assert.Equal(t, successResult, res2)
	assert.Equal(t, 1, testTool.executeCount, "Should be semantic cache hit")
}

func TestCachingMiddleware_ProviderFactory(t *testing.T) {
	// Setup
	tm := &mockToolManager{}
	cacheMiddleware := middleware.NewCachingMiddleware(tm)

	// Helper to trigger factory
	triggerFactory := func(conf *configv1.SemanticCacheConfig) error {
		testTool := &mockTool{
			tool: v1.Tool_builder{
				Name:      proto.String(testToolName),
				ServiceId: proto.String("test-service-factory"),
			}.Build(),
			cacheConfig: configv1.CacheConfig_builder{
				IsEnabled:      proto.Bool(true),
				Strategy:       proto.String("semantic"),
				SemanticConfig: conf,
				Ttl:            durationpb.New(1 * time.Hour),
			}.Build(),
		}
		req := &tool.ExecutionRequest{
			ToolName:   "test-service-factory.test-tool",
			ToolInputs: []byte("hello"),
		}
		ctx := tool.NewContextWithTool(context.Background(), testTool)
		nextFunc := func(ctx context.Context, req *tool.ExecutionRequest) (any, error) {
			return successResult, nil
		}

		// Execute triggers executeSemantic -> providerFactory
		_, err := cacheMiddleware.Execute(ctx, req, nextFunc)
		return err
	}

	// Test 1: OpenAI Config
	err := triggerFactory(configv1.SemanticCacheConfig_builder{
		Openai: configv1.OpenAIEmbeddingProviderConfig_builder{
			ApiKey: configv1.SecretValue_builder{PlainText: proto.String("sk-test")}.Build(),
			Model:  proto.String("text-embedding-3-small"),
		}.Build(),
	}.Build())
	assert.NoError(t, err)

	// Test 2: Ollama Config
	err = triggerFactory(configv1.SemanticCacheConfig_builder{
		Ollama: configv1.OllamaEmbeddingProviderConfig_builder{
			BaseUrl: proto.String("http://localhost:11434"),
			Model:   proto.String("nomic-embed-text"),
		}.Build(),
	}.Build())
	assert.NoError(t, err)

	// Test 3: HTTP Config
	err = triggerFactory(configv1.SemanticCacheConfig_builder{
		Http: configv1.HttpEmbeddingProviderConfig_builder{
			Url:              proto.String("http://localhost:8080"),
			ResponseJsonPath: proto.String("$.embedding"),
		}.Build(),
	}.Build())
	assert.NoError(t, err)

	// Test 4: Legacy OpenAI
	err = triggerFactory(configv1.SemanticCacheConfig_builder{
		Provider: proto.String("openai"),
		ApiKey:   configv1.SecretValue_builder{PlainText: proto.String("sk-test")}.Build(),
		Model:    proto.String("text-embedding-ada-002"),
	}.Build())
	assert.NoError(t, err)

	// Test 5: Unknown Provider
	err = triggerFactory(configv1.SemanticCacheConfig_builder{
		Provider: proto.String("unknown"),
	}.Build())
	assert.NoError(t, err)
}
