// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	armonmetrics "github.com/armon/go-metrics"
	"github.com/mcpany/core/pkg/tool"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestToolMetricsMiddleware_Execute(t *testing.T) {
	// Setup InmemSink
	sink := armonmetrics.NewInmemSink(10*time.Second, 10*time.Minute)
	// Replace global metrics instance
	_, err := armonmetrics.NewGlobal(armonmetrics.DefaultConfig("test"), sink)
	require.NoError(t, err)

	middleware := NewToolMetricsMiddleware()

	t.Run("Success", func(t *testing.T) {
		req := &tool.ExecutionRequest{
			ToolName:   "test_tool",
			ToolInputs: json.RawMessage(`{"arg": "value"}`), // 16 bytes
		}

		next := func(ctx context.Context, req *tool.ExecutionRequest) (any, error) {
			time.Sleep(10 * time.Millisecond)
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: "result"}, // 6 bytes
				},
			}, nil
		}

		// Mock Tool for Service ID
		serviceID := "test_service"
		mockProtoTool := &v1.Tool{
			ServiceId: &serviceID,
		}
		mockTool := &tool.MockTool{
			ToolFunc: func() *v1.Tool {
				return mockProtoTool
			},
		}

		ctx := tool.NewContextWithTool(context.Background(), mockTool)

		_, err := middleware.Execute(ctx, req, next)
		require.NoError(t, err)

		// Verify metrics
		intervals := sink.Data()
		require.NotEmpty(t, intervals)

		metrics := intervals[0]

		// Helper to find metric by partial name
		findMetric := func(namePart string) bool {
			for k := range metrics.Counters {
				if strings.Contains(k, namePart) {
					return true
				}
			}
			for k := range metrics.Samples {
				if strings.Contains(k, namePart) {
					return true
				}
			}
			return false
		}

		assert.True(t, findMetric("tool.execution.total"), "Should record execution total")
		assert.True(t, findMetric("tool.execution.input_bytes"), "Should record input bytes")
		assert.True(t, findMetric("tool.execution.output_bytes"), "Should record output bytes")
		assert.True(t, findMetric("tool.execution.duration"), "Should record duration")
	})
}

func BenchmarkToolMetricsMiddleware_Execute(b *testing.B) {
	// Setup (Global metrics mock might be slow, so do it once)
	sink := armonmetrics.NewInmemSink(10*time.Second, 10*time.Minute)
	_, _ = armonmetrics.NewGlobal(armonmetrics.DefaultConfig("bench"), sink)

	middleware := NewToolMetricsMiddleware()
	req := &tool.ExecutionRequest{
		ToolName:   "bench_tool",
		ToolInputs: json.RawMessage(`{"arg": "value"}`),
	}
	ctx := context.Background()

	// Mock Next
	next := func(ctx context.Context, req *tool.ExecutionRequest) (any, error) {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: "result"},
			},
		}, nil
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = middleware.Execute(ctx, req, next)
	}
}
