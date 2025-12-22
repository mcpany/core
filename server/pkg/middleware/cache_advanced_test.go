// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware_test

import (
	"context"
	"encoding/json"
	"sync"
	"sync/atomic"
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

func TestCachingMiddleware_CanonicalJSON(t *testing.T) {
	// Setup
	tm := &mockToolManager{}
	cacheMiddleware := middleware.NewCachingMiddleware(tm)

	testTool := &mockTool{
		tool: v1.Tool_builder{
			Name:      proto.String("test-tool-canonical"),
			ServiceId: proto.String("test-service"),
		}.Build(),
		cacheConfig: configv1.CacheConfig_builder{
			IsEnabled: proto.Bool(true),
			Ttl:       durationpb.New(1 * time.Minute),
		}.Build(),
	}

	// Two inputs: semantically identical but different byte representation
	input1 := json.RawMessage(`{"a": 1, "b": 2}`)
	input2 := json.RawMessage(`{"b": 2, "a": 1}`)

	req1 := &tool.ExecutionRequest{
		ToolName:   "test-service.test-tool-canonical",
		ToolInputs: input1,
	}
	req2 := &tool.ExecutionRequest{
		ToolName:   "test-service.test-tool-canonical",
		ToolInputs: input2,
	}

	ctx := tool.NewContextWithTool(context.Background(), testTool)

	nextFunc := func(ctx context.Context, req *tool.ExecutionRequest) (any, error) {
		t, _ := tool.GetFromContext(ctx)
		return t.Execute(ctx, req)
	}

	// 1. First call with input1
	res1, err := cacheMiddleware.Execute(ctx, req1, nextFunc)
	require.NoError(t, err)
	assert.Equal(t, successResult, res1)
	assert.Equal(t, 1, testTool.executeCount)

	// 2. Second call with input2 (should hit cache if canonicalization works)
	res2, err := cacheMiddleware.Execute(ctx, req2, nextFunc)
	require.NoError(t, err)
	assert.Equal(t, successResult, res2)
	assert.Equal(t, 1, testTool.executeCount, "Tool should NOT be executed again due to canonical JSON match")
}

func TestCachingMiddleware_Singleflight(t *testing.T) {
	// Setup
	tm := &mockToolManager{}
	cacheMiddleware := middleware.NewCachingMiddleware(tm)

	// Use a tool that sleeps to simulate latency
	testTool := &mockTool{
		tool: v1.Tool_builder{
			Name:      proto.String("test-tool-singleflight"),
			ServiceId: proto.String("test-service"),
		}.Build(),
		cacheConfig: configv1.CacheConfig_builder{
			IsEnabled: proto.Bool(true),
			Ttl:       durationpb.New(1 * time.Minute),
		}.Build(),
	}

	// Override Execute to sleep
	var execCount int32
	blockingNextFunc := func(ctx context.Context, req *tool.ExecutionRequest) (any, error) {
		atomic.AddInt32(&execCount, 1)
		time.Sleep(100 * time.Millisecond)
		return successResult, nil
	}

	req := &tool.ExecutionRequest{
		ToolName:   "test-service.test-tool-singleflight",
		ToolInputs: json.RawMessage(`{}`),
	}
	ctx := tool.NewContextWithTool(context.Background(), testTool)

	// Launch concurrent requests
	var wg sync.WaitGroup
	concurrency := 5
	wg.Add(concurrency)

	for i := 0; i < concurrency; i++ {
		go func() {
			defer wg.Done()
			_, err := cacheMiddleware.Execute(ctx, req, blockingNextFunc)
			assert.NoError(t, err)
		}()
	}

	wg.Wait()

	assert.Equal(t, int32(1), atomic.LoadInt32(&execCount), "Tool should execute only once due to singleflight")
}
