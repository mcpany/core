package middleware

import (
	"context"
	"sync"
	"time"

	"github.com/mcpany/core/pkg/logging"
	"github.com/mcpany/core/pkg/observability"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

// Observability is an MCP middleware that records observability metrics.
func Observability(next mcp.MethodHandler) mcp.MethodHandler {
	var (
		meter    metric.Meter
		requests metric.Int64Counter
		latency  metric.Int64Histogram
		once     sync.Once
	)

	initMetrics := func() {
		meter = observability.MeterProvider.Meter("github.com/mcpany/core/pkg/observability")

		var err error
		requests, err = meter.Int64Counter("mcp.server.requests", metric.WithDescription("Number of MCP requests received."))
		if err != nil {
			logging.GetLogger().Error("failed to create requests counter", "error", err)
		}

		latency, err = meter.Int64Histogram("mcp.server.latency", metric.WithDescription("Latency of MCP requests."), metric.WithUnit("ms"))
		if err != nil {
			logging.GetLogger().Error("failed to create latency histogram", "error", err)
		}
	}

	return func(ctx context.Context, method string, req mcp.Request) (mcp.Result, error) {
		once.Do(initMetrics)

		start := time.Now()
		result, err := next(ctx, method, req)
		duration := time.Since(start)

		attrs := attribute.NewSet(
			attribute.String("method", method),
			attribute.Bool("error", err != nil),
		)

		if requests != nil {
			requests.Add(ctx, 1, metric.WithAttributeSet(attrs))
		}
		if latency != nil {
			latency.Record(ctx, duration.Milliseconds(), metric.WithAttributeSet(attrs))
		}

		return result, err
	}
}
