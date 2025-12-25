// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/mcpany/core/pkg/tool"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestToolMetricsMiddleware_Execute(t *testing.T) {
	// Ensure metrics are registered
	middleware := NewToolMetricsMiddleware()

	checkMetric := func(t *testing.T, toolName string, metricName string, expectedValue float64, checkType string) {
		metrics, err := prometheus.DefaultGatherer.Gather()
		require.NoError(t, err)

		found := false
		for _, mf := range metrics {
			if mf.GetName() == metricName {
				for _, m := range mf.GetMetric() {
					// Check label "tool"
					currentTool := ""
					for _, pair := range m.GetLabel() {
						if pair.GetName() == "tool" {
							currentTool = pair.GetValue()
							break
						}
					}
					if currentTool == toolName {
						found = true
						if checkType == "counter" {
							assert.Equal(t, expectedValue, m.GetCounter().GetValue(), "Counter %s mismatch", metricName)
						} else if checkType == "histogram_sum" {
							assert.Equal(t, expectedValue, m.GetHistogram().GetSampleSum(), "Histogram sum %s mismatch", metricName)
						}
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

		checkMetric(t, toolName, "tool_execution_output_bytes", 6, "histogram_sum")
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

		checkMetric(t, toolName, "tool_execution_output_bytes", 10, "histogram_sum")
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

		checkMetric(t, toolName, "tool_execution_output_bytes", 13, "histogram_sum")
	})

	t.Run("String Result", func(t *testing.T) {
		toolName := "str_tool"
		req := &tool.ExecutionRequest{ToolName: toolName, ToolInputs: json.RawMessage(`{}`)}
		next := func(ctx context.Context, req *tool.ExecutionRequest) (any, error) {
			return "some string", nil // 11 bytes
		}
		_, err := middleware.Execute(context.Background(), req, next)
		require.NoError(t, err)

		checkMetric(t, toolName, "tool_execution_output_bytes", 11, "histogram_sum")
	})

	t.Run("Byte Slice Result", func(t *testing.T) {
		toolName := "bytes_tool"
		req := &tool.ExecutionRequest{ToolName: toolName, ToolInputs: json.RawMessage(`{}`)}
		next := func(ctx context.Context, req *tool.ExecutionRequest) (any, error) {
			return []byte("some bytes"), nil // 10 bytes
		}
		_, err := middleware.Execute(context.Background(), req, next)
		require.NoError(t, err)

		checkMetric(t, toolName, "tool_execution_output_bytes", 10, "histogram_sum")
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
		checkMetric(t, toolName, "tool_execution_output_bytes", 15, "histogram_sum")
	})

	t.Run("Nil Result", func(t *testing.T) {
		toolName := "nil_tool"
		req := &tool.ExecutionRequest{ToolName: toolName, ToolInputs: json.RawMessage(`{}`)}
		next := func(ctx context.Context, req *tool.ExecutionRequest) (any, error) {
			return nil, nil
		}
		_, err := middleware.Execute(context.Background(), req, next)
		require.NoError(t, err)

		checkMetric(t, toolName, "tool_execution_output_bytes", 0, "histogram_sum")
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

		serviceID := "test_service"
		mockProtoTool := &v1.Tool{ServiceId: &serviceID}
		mockTool := &tool.MockTool{ToolFunc: func() *v1.Tool { return mockProtoTool }}
		ctx := tool.NewContextWithTool(context.Background(), mockTool)

		_, err := middleware.Execute(ctx, req, next)
		require.Error(t, err)
		assert.Equal(t, context.Canceled, err)

		// Verify error_type label
		metrics, err := prometheus.DefaultGatherer.Gather()
		require.NoError(t, err)

		foundError := false
		for _, mf := range metrics {
			if mf.GetName() == "tool_executions_total" {
				for _, m := range mf.GetMetric() {
					toolLabel := ""
					errorTypeLabel := ""
					for _, l := range m.GetLabel() {
						if l.GetName() == "tool" {
							toolLabel = l.GetValue()
						}
						if l.GetName() == "error_type" {
							errorTypeLabel = l.GetValue()
						}
					}
					if toolLabel == toolName && errorTypeLabel == "context_canceled" {
						foundError = true
						assert.Equal(t, float64(1), m.GetCounter().GetValue())
					}
				}
			}
		}
		assert.True(t, foundError, "tool_executions_total with error_type=context_canceled not found")
	})
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
