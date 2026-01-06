// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	registerProtocolMetricsOnce sync.Once

	// Define Prometheus metrics for general MCP protocol operations.
	mcpOperationDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "mcp_operation_duration_seconds",
			Help:    "Histogram of MCP operation duration in seconds.",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "status", "error_type"},
	)

	mcpOperationTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "mcp_operations_total",
			Help: "Total number of MCP operations.",
		},
		[]string{"method", "status", "error_type"},
	)

	mcpPayloadSizeBytes = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "mcp_payload_size_bytes",
			Help:    "Histogram of MCP payload size in bytes.",
			Buckets: prometheus.ExponentialBuckets(100, 10, 6),
		},
		[]string{"method", "direction"}, // direction: request, response
	)
)

// PrometheusMetricsMiddleware provides protocol-level metrics for all MCP requests.
// It intercepts requests to track duration, success/failure counts, and payload sizes.
func PrometheusMetricsMiddleware() mcp.Middleware {
	registerProtocolMetricsOnce.Do(func() {
		prometheus.MustRegister(mcpOperationDuration)
		prometheus.MustRegister(mcpOperationTotal)
		prometheus.MustRegister(mcpPayloadSizeBytes)
	})

	return func(next mcp.MethodHandler) mcp.MethodHandler {
		return func(ctx context.Context, method string, req mcp.Request) (mcp.Result, error) {
			start := time.Now()

			result, err := next(ctx, method, req)

			duration := time.Since(start).Seconds()

			status := "success"
			errorType := "none"

			if err != nil {
				status = "error"
				errorType = "unknown"
				if errors.Is(err, context.Canceled) {
					errorType = "canceled"
				} else if errors.Is(err, context.DeadlineExceeded) {
					errorType = "deadline_exceeded"
				} else {
					errMsg := err.Error()
					if len(errMsg) > 0 {
						errorType = "execution_failed"
					}
				}
			}

			labels := prometheus.Labels{
				"method":     method,
				"status":     status,
				"error_type": errorType,
			}

			mcpOperationTotal.With(labels).Inc()
			mcpOperationDuration.With(labels).Observe(duration)

			if result != nil {
				if method == "resources/read" || method == "tools/call" || method == "prompts/get" {
					size := estimateResultSize(result)
					mcpPayloadSizeBytes.With(prometheus.Labels{
						"method":    method,
						"direction": "response",
					}).Observe(float64(size))
				}
			}

			return result, err
		}
	}
}

func estimateResultSize(res mcp.Result) int {
	if res == nil {
		return 0
	}
	switch r := res.(type) {
	case *mcp.ReadResourceResult:
		size := 0
		for _, content := range r.Contents {
			if content != nil {
				if content.Blob != nil {
					size += len(content.Blob)
				}
				if content.Text != "" {
					size += len(content.Text)
				}
			}
		}
		return size
	case *mcp.CallToolResult:
		return calculateToolResultSize(r)
	case *mcp.GetPromptResult:
		size := 0
		for _, msg := range r.Messages {
			if msg.Content != nil {
				switch c := msg.Content.(type) {
				case *mcp.TextContent:
					size += len(c.Text)
				case *mcp.ImageContent:
					size += len(c.Data)
				case *mcp.EmbeddedResource:
					if c.Resource != nil {
						size += len(c.Resource.Blob)
					}
				}
			}
		}
		return size
	}

	b, err := json.Marshal(res)
	if err == nil {
		return len(b)
	}
	return 0
}

func calculateToolResultSize(result any) int {
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
		b, err := json.Marshal(v)
		if err == nil {
			return len(b)
		}
	}
	return 0
}
