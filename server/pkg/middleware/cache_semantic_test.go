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

func TestCachingMiddleware_Semantic(t *testing.T) {
	// Setup
	tm := &mockToolManager{}
	cacheMiddleware := middleware.NewCachingMiddleware(tm)

	strategy := configv1.CacheConfig_STRATEGY_SEMANTIC

	// Configure tool with Semantic Caching
	testTool := &mockTool{
		tool: &v1.Tool{
			Name:      proto.String("test-tool"),
			ServiceId: proto.String("test-service"),
		},
		cacheConfig: &configv1.CacheConfig{
			IsEnabled: proto.Bool(true),
			Ttl:       durationpb.New(1 * time.Hour),
			Type:      &strategy,
			SemanticConfig: &configv1.SemanticCacheConfig{
				SimilarityThreshold: proto.Float32(0.7),
			},
		},
	}

	ctx := tool.NewContextWithTool(context.Background(), testTool)
	nextFunc := func(ctx context.Context, req *tool.ExecutionRequest) (any, error) {
		t, _ := tool.GetFromContext(ctx)
		return t.Execute(ctx, req)
	}

	// 1. First call: "How is the weather?" -> Miss
	req1 := &tool.ExecutionRequest{
		ToolName:   "test-service.test-tool",
		ToolInputs: []byte(`{"query": "How is the weather?"}`),
	}
	res1, err := cacheMiddleware.Execute(ctx, req1, nextFunc)
	require.NoError(t, err)
	assert.Equal(t, successResult, res1)
	assert.Equal(t, 1, testTool.executeCount)

	// 2. Second call: "What is the weather?" -> Should Hit (Semantic Match)
	// "How is the weather?" vs "What is the weather?"
	// BagOfWords (hashed):
	// "how", "is", "the", "weather"
	// "what", "is", "the", "weather"
	// 3 matches out of 4 unique words each. Sim ~0.75.
	req2 := &tool.ExecutionRequest{
		ToolName:   "test-service.test-tool",
		ToolInputs: []byte(`{"query": "What is the weather?"}`),
	}
	res2, err := cacheMiddleware.Execute(ctx, req2, nextFunc)
	require.NoError(t, err)
	assert.Equal(t, successResult, res2)
	assert.Equal(t, 1, testTool.executeCount, "Should be a semantic cache hit")

	// 3. Third call: "Something completely different" -> Miss
	req3 := &tool.ExecutionRequest{
		ToolName:   "test-service.test-tool",
		ToolInputs: []byte(`{"query": "Buy some milk"}`),
	}
	res3, err := cacheMiddleware.Execute(ctx, req3, nextFunc)
	require.NoError(t, err)
	assert.Equal(t, successResult, res3)
	assert.Equal(t, 2, testTool.executeCount, "Should be a miss")
}
