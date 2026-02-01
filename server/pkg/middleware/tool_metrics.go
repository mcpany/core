// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"time"

	"github.com/mcpany/core/server/pkg/tokenizer"
	"github.com/mcpany/core/server/pkg/tool"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	registerMetricsOnce sync.Once

	// Define Prometheus metrics.
	toolExecutionDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "mcpany_tools_call_latency_seconds",
			Help:    "Histogram of tool execution duration in seconds.",
			Buckets: prometheus.DefBuckets, // Use default buckets or customize
		},
		[]string{"tool", "service_id", "status", "error_type"},
	)

	toolExecutionTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "mcpany_tools_call_total",
			// Help string must match the existing registration to avoid conflicts.
			// The conflicting registration seems to use the name as the help string.
			Help: "mcpany_tools_call_total",
		},
		[]string{"tool", "service_id", "status", "error_type"},
	)

	toolExecutionInputBytes = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name: "mcpany_tools_call_input_bytes",
			Help: "Histogram of tool input size in bytes.",
			// Buckets from 100B to 10MB
			Buckets: prometheus.ExponentialBuckets(100, 10, 6),
		},
		[]string{"tool", "service_id"},
	)

	toolExecutionOutputBytes = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name: "mcpany_tools_call_output_bytes",
			Help: "Histogram of tool output size in bytes.",
			// Buckets from 100B to 10MB
			Buckets: prometheus.ExponentialBuckets(100, 10, 6),
		},
		[]string{"tool", "service_id"},
	)

	toolExecutionTokensTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "mcpany_tools_call_tokens_total",
			Help: "Total number of tokens in tool executions.",
		},
		[]string{"tool", "service_id", "direction"}, // direction: input, output
	)

	toolExecutionsInFlight = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "mcpany_tools_call_in_flight",
			Help: "Current number of tool executions in flight.",
		},
		[]string{"tool", "service_id"},
	)
)

// ToolMetricsMiddleware provides detailed metrics for tool executions.
type ToolMetricsMiddleware struct {
	tokenizer tokenizer.Tokenizer
}

// NewToolMetricsMiddleware creates a new ToolMetricsMiddleware.
//
// Parameters:
//   - t: tokenizer.Tokenizer. The tokenizer used to count tokens in tool inputs and outputs.
//     If nil, a simple default tokenizer is used.
//
// Returns:
//   - *ToolMetricsMiddleware: A new instance of ToolMetricsMiddleware with metrics registered.
func NewToolMetricsMiddleware(t tokenizer.Tokenizer) *ToolMetricsMiddleware {
	registerMetricsOnce.Do(func() {
		// Register metrics with the default registry (which server/pkg/metrics also uses/exposes)
		prometheus.MustRegister(toolExecutionDuration)
		prometheus.MustRegister(toolExecutionTotal)
		prometheus.MustRegister(toolExecutionInputBytes)
		prometheus.MustRegister(toolExecutionOutputBytes)
		prometheus.MustRegister(toolExecutionTokensTotal)
		prometheus.MustRegister(toolExecutionsInFlight)
	})
	if t == nil {
		t = tokenizer.NewSimpleTokenizer()
	}
	return &ToolMetricsMiddleware{
		tokenizer: t,
	}
}

// Execute executes the tool metrics middleware.
//
// ctx is the context for the request.
// req is the request object.
// next is the next.
//
// Returns the result.
// Returns an error if the operation fails.
func (m *ToolMetricsMiddleware) Execute(ctx context.Context, req *tool.ExecutionRequest, next tool.ExecutionFunc) (any, error) {
	// Get Service ID if possible (from context or tool)
	var serviceID string
	if t, ok := tool.GetFromContext(ctx); ok && t.Tool() != nil {
		serviceID = t.Tool().GetServiceId()
	}

	labels := prometheus.Labels{
		"tool":       req.ToolName,
		"service_id": serviceID,
	}

	toolExecutionsInFlight.With(labels).Inc()
	defer toolExecutionsInFlight.With(labels).Dec()

	start := time.Now()

	// Record input size and tokens
	inputSize := len(req.ToolInputs)
	inputTokens := m.countInputTokens(req)

	toolExecutionInputBytes.With(labels).Observe(float64(inputSize))
	toolExecutionTokensTotal.With(prometheus.Labels{
		"tool":       req.ToolName,
		"service_id": serviceID,
		"direction":  "input",
	}).Add(float64(inputTokens))

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

	resultLabels := prometheus.Labels{
		"tool":       req.ToolName,
		"service_id": serviceID,
		"status":     status,
		"error_type": errorType,
	}

	toolExecutionTotal.With(resultLabels).Inc()
	toolExecutionDuration.With(resultLabels).Observe(duration)

	// Calculate output size and tokens
	outputSize := calculateOutputSize(result)
	outputTokens := m.countOutputTokens(result)

	toolExecutionOutputBytes.With(labels).Observe(float64(outputSize))
	toolExecutionTokensTotal.With(prometheus.Labels{
		"tool":       req.ToolName,
		"service_id": serviceID,
		"direction":  "output",
	}).Add(float64(outputTokens))

	return result, err
}

func (m *ToolMetricsMiddleware) countInputTokens(req *tool.ExecutionRequest) int {
	// Try to use Arguments map first as it's already parsed
	if req.Arguments != nil {
		tokens, _ := tokenizer.CountTokensInValue(m.tokenizer, req.Arguments)
		return tokens
	}
	// Fallback to ToolInputs byte slice (less accurate if it's JSON, but simple tokenizer might handle strings)
	// Or unmarshal first? The RateLimit middleware logic does this too.
	// For metrics, we can try to unmarshal if not present, but better to avoid double unmarshal if possible.
	// Tool request usually populates Arguments.
	// If ToolInputs is present, we can treat it as string.
	if len(req.ToolInputs) > 0 {
		tokens, _ := m.tokenizer.CountTokens(string(req.ToolInputs))
		return tokens
	}
	return 0
}

func (m *ToolMetricsMiddleware) countOutputTokens(result any) int {
	if result == nil {
		return 0
	}
	return CalculateToolResultTokens(m.tokenizer, result)
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
