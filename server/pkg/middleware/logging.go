// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	"github.com/mcpany/core/server/pkg/logging"
	"github.com/mcpany/core/server/pkg/metrics"
	"github.com/mcpany/core/server/pkg/util"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// LazyLogPayload lazily marshals and redacts a payload for logging.
type LazyLogPayload struct {
	Value any
}

func (l LazyLogPayload) LogValue() slog.Value {
	if l.Value == nil {
		return slog.StringValue("<nil>")
	}
	b, err := json.Marshal(l.Value)
	if err != nil {
		return slog.StringValue("failed to marshal payload")
	}
	// Redact sensitive info
	redacted := util.RedactJSON(b)
	return slog.StringValue(string(redacted))
}

// LoggingMiddleware creates an MCP middleware that logs information about each
// incoming request. It records the start and completion of each request,
// including the duration of the handling.
//
// This is useful for debugging and monitoring the flow of requests through the
// server.
//
// Parameters:
//   - log: The logger to be used. If `nil`, the default global logger will be
//     used.
//
// Returns an `mcp.Middleware` function.
func LoggingMiddleware(log *slog.Logger) mcp.Middleware {
	if log == nil {
		log = logging.GetLogger()
	}

	// Pre-allocate metric names to avoid allocation on every request
	metricRequestTotal := []string{"middleware", "request", "total"}
	metricRequestLatency := []string{"middleware", "request", "latency"}
	metricRequestError := []string{"middleware", "request", "error"}
	metricRequestSuccess := []string{"middleware", "request", "success"}

	return func(next mcp.MethodHandler) mcp.MethodHandler {
		return func(ctx context.Context, method string, req mcp.Request) (mcp.Result, error) {
			start := time.Now()
			metrics.IncrCounter(metricRequestTotal, 1)
			defer metrics.MeasureSince(metricRequestLatency, start)

			// Optimization: Removed redundant "Request received" log to reduce I/O and noise.
			// We log completion/failure below which is sufficient.
			result, err := next(ctx, method, req)
			if err != nil {
				metrics.IncrCounter(metricRequestError, 1)
				log.Error("Request failed",
					"method", method,
					"duration", time.Since(start),
					"error", err,
					"request_payload", LazyLogPayload{Value: req},
				)
			} else {
				metrics.IncrCounter(metricRequestSuccess, 1)
				log.Info("Request completed",
					"method", method,
					"duration", time.Since(start),
					"request_payload", LazyLogPayload{Value: req},
					"response_payload", LazyLogPayload{Value: result},
				)
			}
			return result, err
		}
	}
}
