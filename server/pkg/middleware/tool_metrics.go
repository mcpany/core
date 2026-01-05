// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"time"

	"github.com/mcpany/core/server/pkg/tool"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	registerMetricsOnce sync.Once

	// Define Prometheus metrics.
	toolExecutionDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "tool_execution_duration_seconds",
			Help:    "Histogram of tool execution duration in seconds.",
			Buckets: prometheus.DefBuckets, // Use default buckets or customize
		},
		[]string{"tool", "service_id", "status", "error_type"},
	)

	toolExecutionTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "tool_executions_total",
			Help: "Total number of tool executions.",
		},
		[]string{"tool", "service_id", "status", "error_type"},
	)

	toolExecutionInputBytes = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name: "tool_execution_input_bytes",
			Help: "Histogram of tool input size in bytes.",
			// Buckets from 100B to 10MB
			Buckets: prometheus.ExponentialBuckets(100, 10, 6),
		},
		[]string{"tool", "service_id"},
	)

	toolExecutionOutputBytes = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name: "tool_execution_output_bytes",
			Help: "Histogram of tool output size in bytes.",
			// Buckets from 100B to 10MB
			Buckets: prometheus.ExponentialBuckets(100, 10, 6),
		},
		[]string{"tool", "service_id"},
	)
)

// ToolMetricsMiddleware provides detailed metrics for tool executions.
type ToolMetricsMiddleware struct{}

// NewToolMetricsMiddleware creates a new ToolMetricsMiddleware.
func NewToolMetricsMiddleware() *ToolMetricsMiddleware {
	registerMetricsOnce.Do(func() {
		// Register metrics with the default registry (which server/pkg/metrics also uses/exposes)
		prometheus.MustRegister(toolExecutionDuration)
		prometheus.MustRegister(toolExecutionTotal)
		prometheus.MustRegister(toolExecutionInputBytes)
		prometheus.MustRegister(toolExecutionOutputBytes)
	})
	return &ToolMetricsMiddleware{}
}

// Execute executes the tool metrics middleware.
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

	duration := time.Since(start).Seconds()

	status := "success"
	errorType := "none"
	if err != nil {
		status = "error"
		errorType = "execution_failed"

		switch {
		case errors.Is(err, context.Canceled):
			errorType = "context_canceled"
		case errors.Is(err, context.DeadlineExceeded):
			errorType = "deadline_exceeded"
		}
	}

	labels := prometheus.Labels{
		"tool":       req.ToolName,
		"service_id": serviceID,
		"status":     status,
		"error_type": errorType,
	}

	toolExecutionTotal.With(labels).Inc()
	toolExecutionDuration.With(labels).Observe(duration)

	// Byte metrics usually don't need status/error_type, just tool/service
	byteLabels := prometheus.Labels{
		"tool":       req.ToolName,
		"service_id": serviceID,
	}
	toolExecutionInputBytes.With(byteLabels).Observe(float64(inputSize))

	// Calculate output size
	outputSize := calculateOutputSize(result)
	toolExecutionOutputBytes.With(byteLabels).Observe(float64(outputSize))

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
