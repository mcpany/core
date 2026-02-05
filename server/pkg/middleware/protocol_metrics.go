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

	mcpOperationTokensTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "mcp_operation_tokens_total",
			Help: "Total number of tokens in MCP operations.",
		},
		[]string{"method", "direction", "status"}, // direction: request, response
	)
)

// PrometheusMetricsMiddleware provides protocol-level metrics for all MCP requests.
//
// Summary: Middleware that tracks request duration, success/failure counts, payload sizes, and token counts.
//
// Parameters:
//   - t: tokenizer.Tokenizer. The tokenizer used to estimate token counts.
//
// Returns:
//   - mcp.Middleware: The metrics middleware function.
func PrometheusMetricsMiddleware(t tokenizer.Tokenizer) mcp.Middleware {
	registerProtocolMetricsOnce.Do(func() {
		prometheus.MustRegister(mcpOperationDuration)
		prometheus.MustRegister(mcpOperationTotal)
		prometheus.MustRegister(mcpPayloadSizeBytes)
		prometheus.MustRegister(mcpOperationTokensTotal)
	})

	return func(next mcp.MethodHandler) mcp.MethodHandler {
		return func(ctx context.Context, method string, req mcp.Request) (mcp.Result, error) {
			start := time.Now()

			// Count request tokens
			reqTokens := estimateRequestTokens(t, req)

			result, err := next(ctx, method, req)

			duration := time.Since(start).Seconds()

			status := "success"
			errorType := "none"

			if err != nil {
				status = "error"
				errorType = "unknown"
				switch {
				case errors.Is(err, context.Canceled):
					errorType = "canceled"
				case errors.Is(err, context.DeadlineExceeded):
					errorType = "deadline_exceeded"
				default:
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

			// Record request tokens
			mcpOperationTokensTotal.With(prometheus.Labels{
				"method":    method,
				"direction": "request",
				"status":    status,
			}).Add(float64(reqTokens))

			if result != nil {
				if method == "resources/read" || method == "tools/call" || method == "prompts/get" {
					size := estimateResultSize(result)
					mcpPayloadSizeBytes.With(prometheus.Labels{
						"method":    method,
						"direction": "response",
					}).Observe(float64(size))

					// Count response tokens
					resTokens := estimateResultTokens(t, result)
					mcpOperationTokensTotal.With(prometheus.Labels{
						"method":    method,
						"direction": "response",
						"status":    status,
					}).Add(float64(resTokens))
				}
			}

			return result, err
		}
	}
}

func estimateRequestTokens(t tokenizer.Tokenizer, req mcp.Request) int {
	if req == nil {
		return 0
	}

	// Type switch to access Params
	switch r := req.(type) {
	case *mcp.CallToolRequest:
		if r.Params != nil && r.Params.Arguments != nil {
			c, _ := tokenizer.CountTokensInValue(t, r.Params.Arguments)
			return c
		}
	case *mcp.GetPromptRequest:
		if r.Params != nil && r.Params.Arguments != nil {
			c, _ := tokenizer.CountTokensInValue(t, r.Params.Arguments)
			return c
		}
	}

	return 0
}

func estimateResultTokens(t tokenizer.Tokenizer, res mcp.Result) int {
	if res == nil {
		return 0
	}
	switch r := res.(type) {
	case *mcp.ReadResourceResult:
		count := 0
		for _, content := range r.Contents {
			if content != nil {
				if content.Text != "" {
					c, _ := t.CountTokens(content.Text)
					count += c
				}
				// Blobs are not tokens usually, unless we base64 them?
				// For now ignore blobs for token count, or count as 0.
			}
		}
		return count
	case *mcp.CallToolResult:
		return CalculateToolResultTokens(t, r)
	case *mcp.GetPromptResult:
		count := 0
		for _, msg := range r.Messages {
			if msg.Content != nil {
				switch c := msg.Content.(type) {
				case *mcp.TextContent:
					tc, _ := t.CountTokens(c.Text)
					count += tc
				case *mcp.ImageContent:
					// Images cost tokens too, but hard to estimate without model info.
					// Use a fixed cost or 0?
					// Let's use 0 for now or simple heuristic if we had one.
				case *mcp.EmbeddedResource:
					// Ignore
				}
			}
		}
		return count
	}

	// Fallback to JSON marshaling + token counting string
	// This is expensive but safe fallback
	b, err := json.Marshal(res)
	if err == nil {
		c, _ := t.CountTokens(string(b))
		return c
	}
	return 0
}

// CalculateToolResultTokens calculates the number of tokens in a tool result.
//
// Parameters:
//   - t: tokenizer.Tokenizer. The tokenizer to use for counting.
//   - result: any. The result object to analyze (can be *mcp.CallToolResult, string, []byte, or others).
//
// Returns:
//   - int: The estimated token count.
func CalculateToolResultTokens(t tokenizer.Tokenizer, result any) int {
	if result == nil {
		return 0
	}

	switch v := result.(type) {
	case *mcp.CallToolResult:
		count := 0
		for _, c := range v.Content {
			if tc, ok := c.(*mcp.TextContent); ok {
				c, _ := t.CountTokens(tc.Text)
				count += c
			}
			// Ignore images/resources for now
		}
		return count
	case string:
		c, _ := t.CountTokens(v)
		return c
	case []byte:
		c, _ := t.CountTokens(string(v))
		return c
	default:
		// Try to count in value directly if supported by tokenizer extended utils,
		// otherwise marshal.
		// `CountTokensInValue` supports arbitrary structs.
		c, _ := tokenizer.CountTokensInValue(t, v)
		return c
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
