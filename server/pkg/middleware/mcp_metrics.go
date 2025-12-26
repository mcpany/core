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
	registerMCPMetricsOnce sync.Once

	// mcpRequestDuration tracks the duration of MCP requests.
	mcpRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "mcp_request_duration_seconds",
			Help:    "Histogram of MCP request duration in seconds.",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "status", "error_type"},
	)

	// mcpRequestsTotal tracks the total number of MCP requests.
	mcpRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "mcp_requests_total",
			Help: "Total number of MCP requests.",
		},
		[]string{"method", "status", "error_type"},
	)

	// mcpRequestInputBytes tracks the size of request payloads.
	mcpRequestInputBytes = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "mcp_request_input_bytes",
			Help:    "Histogram of MCP request input size in bytes.",
			Buckets: prometheus.ExponentialBuckets(100, 10, 6),
		},
		[]string{"method"},
	)

	// mcpResponseOutputBytes tracks the size of response payloads.
	mcpResponseOutputBytes = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "mcp_response_output_bytes",
			Help:    "Histogram of MCP response output size in bytes.",
			Buckets: prometheus.ExponentialBuckets(100, 10, 6),
		},
		[]string{"method"},
	)
)

// MCPMetricsMiddleware provides detailed metrics for all MCP protocol requests.
type MCPMetricsMiddleware struct{}

// NewMCPMetricsMiddleware creates a new MCPMetricsMiddleware.
func NewMCPMetricsMiddleware() *MCPMetricsMiddleware {
	registerMCPMetricsOnce.Do(func() {
		prometheus.MustRegister(mcpRequestDuration)
		prometheus.MustRegister(mcpRequestsTotal)
		prometheus.MustRegister(mcpRequestInputBytes)
		prometheus.MustRegister(mcpResponseOutputBytes)
	})
	return &MCPMetricsMiddleware{}
}

// Middleware returns the MCP middleware function.
func (m *MCPMetricsMiddleware) Middleware(next mcp.MethodHandler) mcp.MethodHandler {
	return func(
		ctx context.Context,
		method string,
		req mcp.Request,
	) (mcp.Result, error) {
		start := time.Now()

		// Calculate Input Size
		// Since req is an interface, we marshal it to estimate size.
		// This adds some overhead, but is necessary for accurate byte metrics.
		var inputSize int
		if reqBytes, err := json.Marshal(req); err == nil {
			inputSize = len(reqBytes)
		}
		mcpRequestInputBytes.WithLabelValues(method).Observe(float64(inputSize))

		// Execute Next Handler
		result, err := next(ctx, method, req)

		duration := time.Since(start).Seconds()

		// Determine Status and Error Type
		status := "success"
		errorType := "none"
		if err != nil {
			status = "error"
			// Generic error handling since SDK v1.1.0 doesn't expose JSON-RPC error types publicly
			errorType = "internal_error"
			if errors.Is(err, context.Canceled) {
				errorType = "context_canceled"
			} else if errors.Is(err, context.DeadlineExceeded) {
				errorType = "deadline_exceeded"
			}
		}

		// Record Metrics
		labels := prometheus.Labels{
			"method":     method,
			"status":     status,
			"error_type": errorType,
		}

		mcpRequestsTotal.With(labels).Inc()
		mcpRequestDuration.With(labels).Observe(duration)

		// Calculate Output Size
		var outputSize int
		if result != nil {
			if resBytes, err := json.Marshal(result); err == nil {
				outputSize = len(resBytes)
			}
		}
		mcpResponseOutputBytes.WithLabelValues(method).Observe(float64(outputSize))

		return result, err
	}
}
