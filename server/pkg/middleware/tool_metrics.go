// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"context"
	"encoding/json"
	"time"

	"github.com/mcpany/core/pkg/metrics"
	"github.com/mcpany/core/pkg/tool"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// ToolMetricsMiddleware provides detailed metrics for tool executions.
type ToolMetricsMiddleware struct{}

// NewToolMetricsMiddleware creates a new ToolMetricsMiddleware.
// Returns the result.
func NewToolMetricsMiddleware() *ToolMetricsMiddleware {
	return &ToolMetricsMiddleware{}
}

// Execute executes the tool metrics middleware.
// ctx is the context.
// req is the req.
// next is the next.
// Returns the result, an error.
func (m *ToolMetricsMiddleware) Execute(ctx context.Context, req *tool.ExecutionRequest, next tool.ExecutionFunc) (any, error) {
	start := time.Now()

	// Record input size
	inputSize := len(req.ToolInputs)

	// Get Service ID if possible (from context or tool)
	var serviceID string
	if t, ok := tool.GetFromContext(ctx); ok && t.Tool() != nil {
		serviceID = t.Tool().GetServiceId()
	}

	result, err := next(ctx, req)

	status := "success"
	errorType := "none"
	if err != nil {
		status = "error"
		errorType = "execution_failed"
	}

	labels := []metrics.Label{
		{Name: "tool", Value: req.ToolName},
		{Name: "service_id", Value: serviceID},
		{Name: "status", Value: status},
		{Name: "error_type", Value: errorType},
	}

	metrics.IncrCounterWithLabels([]string{"tool", "execution", "total"}, 1, labels)
	metrics.MeasureSinceWithLabels([]string{"tool", "execution", "duration"}, start, labels)
	metrics.AddSampleWithLabels([]string{"tool", "execution", "input_bytes"}, float32(inputSize), labels)

	// Calculate output size
	outputSize := calculateOutputSize(result)
	metrics.AddSampleWithLabels([]string{"tool", "execution", "output_bytes"}, float32(outputSize), labels)

	return result, err
}

func calculateOutputSize(result any) int {
	if result == nil {
		return 0
	}

	switch v := result.(type) {
	case *mcp.CallToolResult:
		size := 0
		for _, c := range v.Content {
			if tc, ok := c.(*mcp.TextContent); ok {
				size += len(tc.Text)
			} else if ic, ok := c.(*mcp.ImageContent); ok {
				size += len(ic.Data)
			} else if er, ok := c.(*mcp.EmbeddedResource); ok && er.Resource != nil {
				size += len(er.Resource.Blob)
			}
		}
		return size
	case string:
		return len(v)
	case []byte:
		return len(v)
	default:
		// Fallback to json marshal size estimate
		b, err := json.Marshal(v)
		if err == nil {
			return len(b)
		}
	}
	return 0
}
