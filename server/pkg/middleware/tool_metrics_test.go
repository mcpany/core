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

	// Helper to reset sink - InmemSink accumulates, so we just rely on checking the latest interval or specific values if unique labels used.
	// But since labels are same for same tool/service, values aggregate in the same interval.
	// We should probably force a new interval or just check diff?
	// The sink rotates intervals based on time.
	// We can't easily force rotation.
	// However, for unit tests, we can just create a NEW sink for each test if we want isolation.
	// But NewGlobal is global.
	// We can rely on checking the Sum increase? Or just that it IS recorded.
	// If we use different tool names, we get different keys!
	// So ensure each test uses unique ToolName.

	checkMetric := func(t *testing.T, toolName string, metricSuffix string, expectedValue float64) {
		intervals := sink.Data()
		require.NotEmpty(t, intervals)
		// We might have multiple intervals. Check the last one or all?
		// Since we use unique tool names, we just need to find the key in any interval (likely the last/current one).

		found := false
		for _, interval := range intervals {
			for k, v := range interval.Samples {
				// Key format: name;label=val;label=val
				// We look for name part and tool label
				if strings.Contains(k, "tool.execution."+metricSuffix) && strings.Contains(k, "tool="+toolName) {
					found = true
					// Check Sum (since we added 1 sample, Sum should be Value)
					// Or if multiple calls, Sum accumulates.
					// We expect 1 call per test case.
					assert.Equal(t, expectedValue, v.Sum, "Metric %s value mismatch", k)
					return
				}
			}
		}
		assert.True(t, found, "Metric %s for tool %s not found", metricSuffix, toolName)
	}

	t.Run("Success with TextContent", func(t *testing.T) {
		toolName := "test_tool_text"
		req := &tool.ExecutionRequest{
			ToolName:   toolName,
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

		checkMetric(t, toolName, "output_bytes", 6)
	})

	t.Run("ImageContent", func(t *testing.T) {
		toolName := "image_tool"
		req := &tool.ExecutionRequest{
			ToolName:   toolName,
			ToolInputs: json.RawMessage(`{}`),
		}

		data := []byte("image_data") // 10 bytes
		next := func(ctx context.Context, req *tool.ExecutionRequest) (any, error) {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.ImageContent{Data: data, MIMEType: "image/png"},
				},
			}, nil
		}

		_, err := middleware.Execute(context.Background(), req, next)
		require.NoError(t, err)

		checkMetric(t, toolName, "output_bytes", 10)
	})

	t.Run("EmbeddedResource", func(t *testing.T) {
		toolName := "resource_tool"
		req := &tool.ExecutionRequest{
			ToolName:   toolName,
			ToolInputs: json.RawMessage(`{}`),
		}

		data := []byte("resource_blob") // 13 bytes
		next := func(ctx context.Context, req *tool.ExecutionRequest) (any, error) {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.EmbeddedResource{
						Resource: &mcp.ResourceContents{Blob: data},
					},
				},
			}, nil
		}

		_, err := middleware.Execute(context.Background(), req, next)
		require.NoError(t, err)

		checkMetric(t, toolName, "output_bytes", 13)
	})

	t.Run("String Result", func(t *testing.T) {
		toolName := "str_tool"
		req := &tool.ExecutionRequest{ToolName: toolName, ToolInputs: json.RawMessage(`{}`)}
		next := func(ctx context.Context, req *tool.ExecutionRequest) (any, error) {
			return "some string", nil // 11 bytes
		}
		_, err := middleware.Execute(context.Background(), req, next)
		require.NoError(t, err)

		checkMetric(t, toolName, "output_bytes", 11)
	})

	t.Run("Byte Slice Result", func(t *testing.T) {
		toolName := "bytes_tool"
		req := &tool.ExecutionRequest{ToolName: toolName, ToolInputs: json.RawMessage(`{}`)}
		next := func(ctx context.Context, req *tool.ExecutionRequest) (any, error) {
			return []byte("some bytes"), nil // 10 bytes
		}
		_, err := middleware.Execute(context.Background(), req, next)
		require.NoError(t, err)

		checkMetric(t, toolName, "output_bytes", 10)
	})

	t.Run("JSON Fallback", func(t *testing.T) {
		toolName := "obj_tool"
		req := &tool.ExecutionRequest{ToolName: toolName, ToolInputs: json.RawMessage(`{}`)}
		next := func(ctx context.Context, req *tool.ExecutionRequest) (any, error) {
			return map[string]string{"key": "value"}, nil
		}
		_, err := middleware.Execute(context.Background(), req, next)
		require.NoError(t, err)

		// json: {"key":"value"} = 15 bytes
		checkMetric(t, toolName, "output_bytes", 15)
	})

	t.Run("Nil Result", func(t *testing.T) {
		toolName := "nil_tool"
		req := &tool.ExecutionRequest{ToolName: toolName, ToolInputs: json.RawMessage(`{}`)}
		next := func(ctx context.Context, req *tool.ExecutionRequest) (any, error) {
			return nil, nil
		}
		_, err := middleware.Execute(context.Background(), req, next)
		require.NoError(t, err)

		checkMetric(t, toolName, "output_bytes", 0)
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
