// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"context"

	"github.com/mcpany/core/server/pkg/tool"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// TracingMiddleware provides OpenTelemetry tracing for tool executions.
type TracingMiddleware struct {
	tracer trace.Tracer
}

// NewTracingMiddleware creates a new TracingMiddleware.
func NewTracingMiddleware() *TracingMiddleware {
	return &TracingMiddleware{
		tracer: otel.Tracer("mcpany/server/middleware/tracing"),
	}
}

// Execute executes the tracing middleware.
func (m *TracingMiddleware) Execute(ctx context.Context, req *tool.ExecutionRequest, next tool.ExecutionFunc) (any, error) {
	ctx, span := m.tracer.Start(ctx, "tool.execute", trace.WithAttributes(
		attribute.String("tool.name", req.ToolName),
	))
	defer span.End()

	// Add service ID if available
	if t, ok := tool.GetFromContext(ctx); ok && t.Tool() != nil {
		span.SetAttributes(attribute.String("service.id", t.Tool().GetServiceId()))
	}

	// Add input size
	if req.ToolInputs != nil {
		span.SetAttributes(attribute.Int("tool.input_size", len(req.ToolInputs)))
	}

	result, err := next(ctx, req)

	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	} else {
		span.SetStatus(codes.Ok, "OK")
	}

	return result, err
}
