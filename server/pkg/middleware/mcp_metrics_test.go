// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/assert"
)

func TestMCPMetricsMiddleware_Middleware(t *testing.T) {
	// Initialize metrics (safe to call multiple times due to sync.Once)
	middleware := NewMCPMetricsMiddleware()

	// Mock Handler
	mockHandler := func(ctx context.Context, method string, req mcp.Request) (mcp.Result, error) {
		if method == "error_method" {
			return nil, errors.New("mock error")
		}
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: "success"},
			},
		}, nil
	}

	// Wrap Handler
	handler := middleware.Middleware(mockHandler)

	// Test Success Case
	// Fix: Construct CallToolParamsRaw properly
	args := map[string]interface{}{"foo": "bar"}
	argsBytes, _ := json.Marshal(args)
	req := &mcp.CallToolRequest{
		Params: &mcp.CallToolParamsRaw{
			Name:      "test_tool",
			Arguments: json.RawMessage(argsBytes),
		},
	}
	ctx := context.Background()

	_, err := handler(ctx, "tools/call", req)
	assert.NoError(t, err)

	// Verify Metrics
	// We can use testutil.ToFloat64 to check counter values

	// Check mcp_requests_total{method="tools/call", status="success", error_type="none"}
	counterCount := testutil.ToFloat64(mcpRequestsTotal.With(prometheus.Labels{
		"method":     "tools/call",
		"status":     "success",
		"error_type": "none",
	}))
	assert.GreaterOrEqual(t, counterCount, 1.0)

	// Test Error Case
	_, err = handler(ctx, "error_method", req)
	assert.Error(t, err)

	// Check mcp_requests_total{method="error_method", status="error", error_type="internal_error"}
	errorCounterCount := testutil.ToFloat64(mcpRequestsTotal.With(prometheus.Labels{
		"method":     "error_method",
		"status":     "error",
		"error_type": "internal_error",
	}))
	assert.GreaterOrEqual(t, errorCounterCount, 1.0)
}

func BenchmarkMCPMetricsMiddleware(b *testing.B) {
	middleware := NewMCPMetricsMiddleware()
	mockHandler := func(ctx context.Context, method string, req mcp.Request) (mcp.Result, error) {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: "success"},
			},
		}, nil
	}
	handler := middleware.Middleware(mockHandler)

	args := map[string]interface{}{"foo": "bar", "baz": 123, "list": []int{1, 2, 3}}
	argsBytes, _ := json.Marshal(args)
	req := &mcp.CallToolRequest{
		Params: &mcp.CallToolParamsRaw{
			Name:      "test_tool",
			Arguments: json.RawMessage(argsBytes),
		},
	}
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = handler(ctx, "tools/call", req)
	}
}
