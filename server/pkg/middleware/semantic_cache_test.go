// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

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
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/durationpb"
)

func TestCachingMiddleware_SemanticHit(t *testing.T) {
	// Setup
	tm := &mockToolManager{}
	cacheMiddleware := middleware.NewCachingMiddleware(tm)

	testTool := &mockTool{
		tool: v1.Tool_builder{
			Name:      proto.String("test-tool-semantic"),
			ServiceId: proto.String("test-service"),
		}.Build(),
		cacheConfig: configv1.CacheConfig_builder{
			IsEnabled: proto.Bool(true),
			Ttl:       durationpb.New(1 * time.Hour),
			Semantic: configv1.SemanticCacheConfig_builder{
				IsEnabled:           proto.Bool(true),
				SimilarityThreshold: proto.Float32(0.8),
				EmbeddingProvider:   proto.String("local"),
			}.Build(),
		}.Build(),
	}

	// Request 1: "Hello World"
	req1 := &tool.ExecutionRequest{
		ToolName:   "test-service.test-tool-semantic",
		ToolInputs: json.RawMessage(`{"query": "Hello World"}`),
	}

	// Context
	ctx := tool.NewContextWithTool(context.Background(), testTool)

	// Next Func
	nextFunc := func(ctx context.Context, req *tool.ExecutionRequest) (any, error) {
		t, _ := tool.GetFromContext(ctx)
		return t.Execute(ctx, req)
	}

	// 1. First call - should execute (Miss)
	res, err := cacheMiddleware.Execute(ctx, req1, nextFunc)
	require.NoError(t, err)
	assert.Equal(t, successResult, res)
	assert.Equal(t, 1, testTool.executeCount)

	// Request 2: "Hello World." (Should be semantically similar enough)
	req2 := &tool.ExecutionRequest{
		ToolName:   "test-service.test-tool-semantic",
		ToolInputs: json.RawMessage(`{"query": "Hello World."}`),
	}

	// 2. Second call - should hit semantic cache
	res, err = cacheMiddleware.Execute(ctx, req2, nextFunc)
	require.NoError(t, err)
	assert.Equal(t, successResult, res)
	// Execute count should remain 1 if semantic cache worked
	assert.Equal(t, 1, testTool.executeCount, "Should have hit semantic cache")
}
