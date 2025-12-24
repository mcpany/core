// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware_test

import (
	"context"
	"os"
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

func TestCachingMiddleware_SemanticCache_Persistence(t *testing.T) {
	// 1. Setup SQLite DB
	f, err := os.CreateTemp("", "semantic_cache_test_persistence_*.db")
	require.NoError(t, err)
	dbPath := f.Name()
	f.Close()
	defer os.Remove(dbPath)

	// Mock Embedding Provider
	mockFactory := &MockProviderFactory{
		embeddings: map[string][]float32{
			"hello": {1.0, 0.0, 0.0},
		},
	}

	// Define Tool
	toolName := "test-tool"
	serviceName := "test-service"
	fullToolName := serviceName + "." + toolName

	testTool := &mockTool{
		tool: v1.Tool_builder{
			Name:      proto.String(toolName),
			ServiceId: proto.String(serviceName),
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

	ctx := tool.NewContextWithTool(context.Background(), testTool)
	req := &tool.ExecutionRequest{
		ToolName:   fullToolName,
		ToolInputs: []byte("hello"),
	}

	// ---------------------------------------------------------
	// PHASE 1: Run Server Instance 1, populate cache
	// ---------------------------------------------------------
	{
		tm := &mockToolManager{}
		mw := middleware.NewCachingMiddleware(tm, dbPath)
		mw.SetProviderFactory(mockFactory.Create)

		nextFunc := func(ctx context.Context, req *tool.ExecutionRequest) (any, error) {
			t, _ := tool.GetFromContext(ctx)
			return t.Execute(ctx, req)
		}

		// Execute - should be a miss, so tool executes
		res, err := mw.Execute(ctx, req, nextFunc)
		require.NoError(t, err)
		assert.Equal(t, successResult, res)
		assert.Equal(t, 1, testTool.executeCount)

		// Wait a bit for async writes if any (SQLiteVectorStore writes synchronously though)
	}

	// ---------------------------------------------------------
	// PHASE 2: Run Server Instance 2 (New Middleware, SAME DB)
	// ---------------------------------------------------------
	{
		tm := &mockToolManager{}
		mw := middleware.NewCachingMiddleware(tm, dbPath)
		mw.SetProviderFactory(mockFactory.Create)

		// Reset tool execution count to verify cache hit
		testTool.executeCount = 0

		nextFunc := func(ctx context.Context, req *tool.ExecutionRequest) (any, error) {
			t, _ := tool.GetFromContext(ctx)
			return t.Execute(ctx, req)
		}

		// Execute - should be a HIT from SQLite
		res, err := mw.Execute(ctx, req, nextFunc)
		require.NoError(t, err)
		assert.Equal(t, successResult, res)
		assert.Equal(t, 0, testTool.executeCount, "Tool should NOT execute; result should be from persistent cache")
	}
}

func TestCachingMiddleware_SemanticCache_Persistence_StructResult(t *testing.T) {
	// 1. Setup SQLite DB
	f, err := os.CreateTemp("", "semantic_cache_test_persistence_struct_*.db")
	require.NoError(t, err)
	dbPath := f.Name()
	f.Close()
	defer os.Remove(dbPath)

	// Mock Embedding Provider
	mockFactory := &MockProviderFactory{
		embeddings: map[string][]float32{
			"complex": {0.5, 0.5, 0.0},
		},
	}

	// Tool that returns a struct
	testTool := &mockTool{
		tool: v1.Tool_builder{
			Name:      proto.String("complex-tool"),
			ServiceId: proto.String("test-service"),
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

	// We can't easily mock Execute to return a complex struct because mockTool.Execute returns "success" string.
	// But we can update mockTool to support custom result.
	// We'll define a new mock for this test.
	complexTool := &complexMockTool{
		mockTool: testTool,
		result: map[string]interface{}{
			"content": []interface{}{
				map[string]interface{}{
					"type": "text",
					"text": "complex result",
				},
			},
			"isError": false,
		},
	}

	ctx := tool.NewContextWithTool(context.Background(), complexTool)
	req := &tool.ExecutionRequest{
		ToolName:   "test-service.complex-tool",
		ToolInputs: []byte("complex"),
	}

	// PHASE 1: Populate
	{
		tm := &mockToolManager{}
		mw := middleware.NewCachingMiddleware(tm, dbPath)
		mw.SetProviderFactory(mockFactory.Create)

		nextFunc := func(ctx context.Context, req *tool.ExecutionRequest) (any, error) {
			t, _ := tool.GetFromContext(ctx)
			return t.Execute(ctx, req)
		}

		res, err := mw.Execute(ctx, req, nextFunc)
		require.NoError(t, err)
		// Result should be the map
		assert.Equal(t, complexTool.result, res)
	}

	// PHASE 2: Retrieve and Verify Type
	{
		tm := &mockToolManager{}
		mw := middleware.NewCachingMiddleware(tm, dbPath)
		mw.SetProviderFactory(mockFactory.Create)

		nextFunc := func(ctx context.Context, req *tool.ExecutionRequest) (any, error) {
			t, _ := tool.GetFromContext(ctx)
			return t.Execute(ctx, req)
		}

		res, err := mw.Execute(ctx, req, nextFunc)
		require.NoError(t, err)

		// The result from cache (SQLite -> JSON -> Map) should match the original result (Map)
		// If we stored a *struct*, it would come back as map.
		// Since we stored a map, it comes back as map.
		// Key verification: Does it exactly match?
		assert.Equal(t, complexTool.result, res)
	}
}

type complexMockTool struct {
	*mockTool
	result any
}

func (m *complexMockTool) Execute(_ context.Context, _ *tool.ExecutionRequest) (any, error) {
	m.mockTool.executeCount++
	return m.result, nil
}
