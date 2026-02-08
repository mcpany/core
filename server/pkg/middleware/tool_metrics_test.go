package middleware

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/mcpany/core/server/pkg/tool"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestToolMetricsMiddleware_Execute(t *testing.T) {
	// Ensure metrics are registered
	middleware := NewToolMetricsMiddleware(nil)

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
		mockProtoTool := v1.Tool_builder{
			ServiceId: &serviceID,
		}.Build()
		mockTool := &tool.MockTool{
			ToolFunc: func() *v1.Tool {
				return mockProtoTool
			},
		}

		ctx := tool.NewContextWithTool(context.Background(), mockTool)

		_, err := middleware.Execute(ctx, req, next)
		require.NoError(t, err)

		checkMetric(t, toolName, "mcpany_tools_call_output_bytes", 6, "histogram_sum")
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

		checkMetric(t, toolName, "mcpany_tools_call_output_bytes", 10, "histogram_sum")
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

		checkMetric(t, toolName, "mcpany_tools_call_output_bytes", 13, "histogram_sum")
	})

	t.Run("String Result", func(t *testing.T) {
		toolName := "str_tool"
		req := &tool.ExecutionRequest{ToolName: toolName, ToolInputs: json.RawMessage(`{}`)}
		next := func(ctx context.Context, req *tool.ExecutionRequest) (any, error) {
			return "some string", nil // 11 bytes
		}
		_, err := middleware.Execute(context.Background(), req, next)
		require.NoError(t, err)

		checkMetric(t, toolName, "mcpany_tools_call_output_bytes", 11, "histogram_sum")
	})

	t.Run("Byte Slice Result", func(t *testing.T) {
		toolName := "bytes_tool"
		req := &tool.ExecutionRequest{ToolName: toolName, ToolInputs: json.RawMessage(`{}`)}
		next := func(ctx context.Context, req *tool.ExecutionRequest) (any, error) {
			return []byte("some bytes"), nil // 10 bytes
		}
		_, err := middleware.Execute(context.Background(), req, next)
		require.NoError(t, err)

		checkMetric(t, toolName, "mcpany_tools_call_output_bytes", 10, "histogram_sum")
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
		checkMetric(t, toolName, "mcpany_tools_call_output_bytes", 15, "histogram_sum")
	})

	t.Run("TokenUsage", func(t *testing.T) {
		toolName := "token_tool"
		// Simple tokenizer counts words. "hello world" -> 2 tokens.
		req := &tool.ExecutionRequest{
			ToolName:   toolName,
			ToolInputs: json.RawMessage(`"hello world"`),
		}

		next := func(ctx context.Context, req *tool.ExecutionRequest) (any, error) {
			// "response value" -> 2 tokens
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: "response value"},
				},
			}, nil
		}

		serviceID := "test_service_token"
		mockProtoTool := v1.Tool_builder{ServiceId: &serviceID}.Build()
		mockTool := &tool.MockTool{ToolFunc: func() *v1.Tool { return mockProtoTool }}
		ctx := tool.NewContextWithTool(context.Background(), mockTool)

		// Before execution, gauge should be 0 (or undefined for this tool)
		// We can't easily check "before" because it might not exist yet.

		// During execution? Hard to test with synchronous call unless we use a goroutine.
		// Let's test the final counters first.

		_, err := middleware.Execute(ctx, req, next)
		require.NoError(t, err)

		// Check Input Tokens
		// Input: `"hello world"`. `len` = 13. `13 / 4 = 3`. Correct.
		// Output: `"response value"`. `len` = 14. `14 / 4 = 3`. Correct.

		checkMetric(t, toolName, "mcpany_tools_call_tokens_total", 3, "counter")
		checkMetric(t, toolName, "mcpany_tools_call_tokens_total", 3, "counter")
	})

	t.Run("Concurrency", func(t *testing.T) {
		toolName := "concurrency_tool"
		req := &tool.ExecutionRequest{
			ToolName:   toolName,
			ToolInputs: json.RawMessage(`{}`),
		}

		startCh := make(chan struct{})
		doneCh := make(chan struct{})

		next := func(ctx context.Context, req *tool.ExecutionRequest) (any, error) {
			close(startCh)
			<-doneCh
			return nil, nil
		}

		go func() {
			_, _ = middleware.Execute(context.Background(), req, next)
		}()

		<-startCh
		// Now it is in flight.
		checkMetric(t, toolName, "mcpany_tools_call_in_flight", 1, "gauge")

		close(doneCh)
		// Wait for goroutine to finish?
		time.Sleep(10 * time.Millisecond) // Flaky but simple

		checkMetric(t, toolName, "mcpany_tools_call_in_flight", 0, "gauge")
	})

	t.Run("Nil Result", func(t *testing.T) {
		toolName := "nil_tool"
		req := &tool.ExecutionRequest{ToolName: toolName, ToolInputs: json.RawMessage(`{}`)}
		next := func(ctx context.Context, req *tool.ExecutionRequest) (any, error) {
			return nil, nil
		}
		_, err := middleware.Execute(context.Background(), req, next)
		require.NoError(t, err)

		checkMetric(t, toolName, "mcpany_tools_call_output_bytes", 0, "histogram_sum")
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
		mockProtoTool := v1.Tool_builder{ServiceId: &serviceID}.Build()
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
			if mf.GetName() == "mcpany_tools_call_total" {
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
		assert.True(t, foundError, "mcpany_tools_call_total with error_type=context_canceled not found")
	})
}

func BenchmarkToolMetricsMiddleware_Execute(b *testing.B) {
	middleware := NewToolMetricsMiddleware(nil)
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

// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

func TestCalculateOutputSize(t *testing.T) {
	tests := []struct {
		name     string
		input    any
		expected int
	}{
		{
			name:     "Nil",
			input:    nil,
			expected: 0,
		},
		{
			name: "CallToolResult with TextContent",
			input: &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: "hello"},
				},
			},
			expected: 5,
		},
		{
			name: "CallToolResult with ImageContent",
			input: &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.ImageContent{Data: []byte("image_data")},
				},
			},
			expected: 10,
		},
		{
			name: "CallToolResult with EmbeddedResource",
			input: &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.EmbeddedResource{
						Resource: &mcp.ResourceContents{
							Blob: []byte("resource_blob"),
						},
					},
				},
			},
			expected: 13,
		},
		{
			name: "CallToolResult with Mixed Content",
			input: &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: "ab"},
					&mcp.ImageContent{Data: []byte("cd")},
				},
			},
			expected: 4,
		},
		{
			name:     "String",
			input:    "hello world",
			expected: 11,
		},
		{
			name:     "Bytes",
			input:    []byte("byte slice"),
			expected: 10,
		},
		{
			name:     "Map (JSON fallback)",
			input:    map[string]string{"k": "v"},
			expected: 9, // {"k":"v"}
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, calculateOutputSize(tt.input))
		})
	}
}

func TestCachingMiddleware_Clear(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Mock not needed for NewCachingMiddleware call but needed for args type
	mockToolManager := tool.NewMockManagerInterface(ctrl)

	m := NewCachingMiddleware(mockToolManager)

	// Set something in cache
	ctx := context.Background()
	req := &tool.ExecutionRequest{
		ToolName:   "test_tool",
		ToolInputs: []byte("{}"),
	}
	key := m.getCacheKey(req)
	err := m.cache.Set(ctx, key, "value")
	require.NoError(t, err)

	// Verify it's there
	val, err := m.cache.Get(ctx, key)
	require.NoError(t, err)
	assert.Equal(t, "value", val)

	// Clear
	err = m.Clear(ctx)
	require.NoError(t, err)

	// Verify it's gone
	_, err = m.cache.Get(ctx, key)
	assert.Error(t, err)
}
