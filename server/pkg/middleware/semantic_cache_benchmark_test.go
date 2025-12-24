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
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/durationpb"
)

func BenchmarkSemanticCache(b *testing.B) {
	// Setup persistent store
	tmpFile, err := os.CreateTemp("", "bench-vector-store-*.db")
	require.NoError(b, err)
	dbPath := tmpFile.Name()
	defer os.Remove(dbPath)
	_ = tmpFile.Close()

	tm := &mockToolManager{}
	cacheMiddleware := middleware.NewCachingMiddleware(tm, dbPath)

	// Mock Embedding Provider
	mockFactory := &MockProviderFactory{
		embeddings: map[string][]float32{
			"hello": {1.0, 0.0, 0.0},
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

	// Pre-fill cache
	_, _ = cacheMiddleware.Execute(ctx, req, nextFunc)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = cacheMiddleware.Execute(ctx, req, nextFunc)
	}
}
