// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/mcpany/core/pkg/tool"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestToolMetricsMiddleware_Execute(t *testing.T) {
	middleware := NewToolMetricsMiddleware()

	// Helper to check histogram sum for a specific metric and tool
	checkHistogramSum := func(t *testing.T, toolName string, metricName string, expectedValue float64) {
		mfs, err := prometheus.DefaultGatherer.Gather()
		require.NoError(t, err)

		found := false
		for _, mf := range mfs {
			if mf.GetName() == metricName {
				for _, m := range mf.GetMetric() {
					currentTool := ""
					for _, l := range m.GetLabel() {
						if l.GetName() == "tool" {
							currentTool = l.GetValue()
						}
					}
					if currentTool == toolName {
						found = true
						assert.Equal(t, expectedValue, m.GetHistogram().GetSampleSum(), "Metric %s sum mismatch", metricName)
						assert.GreaterOrEqual(t, m.GetHistogram().GetSampleCount(), uint64(1), "Metric %s count mismatch", metricName)
					}
				}
			}
		}
		assert.True(t, found, "Metric %s for tool %s not found", metricName, toolName)
	}

	t.Run("Success with TextContent", func(t *testing.T) {
		toolName := "test_tool_text"
		req := &tool.ExecutionRequest{
			ToolName:   toolName,
			ToolInputs: json.RawMessage(`{"arg": "value"}`), // 16 bytes
		}

		next := func(ctx context.Context, req *tool.ExecutionRequest) (any, error) {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: "result"}, // 6 bytes
				},
			}, nil
		}

		ctx := tool.NewContextWithTool(context.Background(), mockTool("service_text"))
		_, err := middleware.Execute(ctx, req, next)
		require.NoError(t, err)

		checkHistogramSum(t, toolName, "mcp_tool_execution_output_bytes", 6)
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

		ctx := tool.NewContextWithTool(context.Background(), mockTool("service_image"))
		_, err := middleware.Execute(ctx, req, next)
		require.NoError(t, err)

		checkHistogramSum(t, toolName, "mcp_tool_execution_output_bytes", 10)
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

		ctx := tool.NewContextWithTool(context.Background(), mockTool("service_resource"))
		_, err := middleware.Execute(ctx, req, next)
		require.NoError(t, err)

		checkHistogramSum(t, toolName, "mcp_tool_execution_output_bytes", 13)
	})

	t.Run("String Result", func(t *testing.T) {
		toolName := "str_tool"
		req := &tool.ExecutionRequest{ToolName: toolName, ToolInputs: json.RawMessage(`{}`)}
		next := func(ctx context.Context, req *tool.ExecutionRequest) (any, error) {
			return "some string", nil // 11 bytes
		}
		ctx := tool.NewContextWithTool(context.Background(), mockTool("service_str"))
		_, err := middleware.Execute(ctx, req, next)
		require.NoError(t, err)

		checkHistogramSum(t, toolName, "mcp_tool_execution_output_bytes", 11)
	})

	t.Run("Byte Slice Result", func(t *testing.T) {
		toolName := "bytes_tool"
		req := &tool.ExecutionRequest{ToolName: toolName, ToolInputs: json.RawMessage(`{}`)}
		next := func(ctx context.Context, req *tool.ExecutionRequest) (any, error) {
			return []byte("some bytes"), nil // 10 bytes
		}
		ctx := tool.NewContextWithTool(context.Background(), mockTool("service_bytes"))
		_, err := middleware.Execute(ctx, req, next)
		require.NoError(t, err)

		checkHistogramSum(t, toolName, "mcp_tool_execution_output_bytes", 10)
	})

	t.Run("JSON Fallback", func(t *testing.T) {
		toolName := "obj_tool"
		req := &tool.ExecutionRequest{ToolName: toolName, ToolInputs: json.RawMessage(`{}`)}
		next := func(ctx context.Context, req *tool.ExecutionRequest) (any, error) {
			return map[string]string{"key": "value"}, nil
		}
		ctx := tool.NewContextWithTool(context.Background(), mockTool("service_obj"))
		_, err := middleware.Execute(ctx, req, next)
		require.NoError(t, err)

		// json: {"key":"value"} = 15 bytes
		checkHistogramSum(t, toolName, "mcp_tool_execution_output_bytes", 15)
	})

	t.Run("Nil Result", func(t *testing.T) {
		toolName := "nil_tool"
		req := &tool.ExecutionRequest{ToolName: toolName, ToolInputs: json.RawMessage(`{}`)}
		next := func(ctx context.Context, req *tool.ExecutionRequest) (any, error) {
			return nil, nil
		}
		ctx := tool.NewContextWithTool(context.Background(), mockTool("service_nil"))
		_, err := middleware.Execute(ctx, req, next)
		require.NoError(t, err)

		checkHistogramSum(t, toolName, "mcp_tool_execution_output_bytes", 0)
	})

	t.Run("ContextCanceled Error", func(t *testing.T) {
		toolName := "test_tool_canceled"
		req := &tool.ExecutionRequest{
			ToolName:   toolName,
			ToolInputs: json.RawMessage(`{}`),
		}

		next := func(ctx context.Context, req *tool.ExecutionRequest) (any, error) {
			return nil, context.Canceled
		}

		ctx := tool.NewContextWithTool(context.Background(), mockTool("service_canceled"))
		_, err := middleware.Execute(ctx, req, next)
		require.Error(t, err)
		assert.Equal(t, context.Canceled, err)

		// Verify error_type label
		mfs, err := prometheus.DefaultGatherer.Gather()
		require.NoError(t, err)

		found := false
		for _, mf := range mfs {
			if mf.GetName() == "mcp_tool_executions_total" {
				for _, m := range mf.GetMetric() {
					var toolLabel, errorTypeLabel string
					for _, l := range m.GetLabel() {
						if l.GetName() == "tool" {
							toolLabel = l.GetValue()
						}
						if l.GetName() == "error_type" {
							errorTypeLabel = l.GetValue()
						}
					}
					if toolLabel == toolName && errorTypeLabel == "context_canceled" {
						found = true
						assert.Equal(t, float64(1), m.GetCounter().GetValue())
					}
				}
			}
		}
		assert.True(t, found, "mcp_tool_executions_total with error_type=context_canceled not found")
	})
}

func mockTool(serviceID string) tool.Tool {
	mockProtoTool := &v1.Tool{
		ServiceId: &serviceID,
	}
	return &tool.MockTool{
		ToolFunc: func() *v1.Tool {
			return mockProtoTool
		},
	}
}

func BenchmarkToolMetricsMiddleware_Execute(b *testing.B) {
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
