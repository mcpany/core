// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"context"
	"testing"

	"github.com/mcpany/core/server/pkg/tool"
	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

func TestTracingMiddleware_Execute(t *testing.T) {
	// Setup tracer
	exporter := tracetest.NewInMemoryExporter()
	tp := trace.NewTracerProvider(
		trace.WithSyncer(exporter),
	)
	otel.SetTracerProvider(tp)

	middleware := NewTracingMiddleware()

	req := &tool.ExecutionRequest{
		ToolName:   "test-tool",
		ToolInputs: []byte(`{"arg": "val"}`),
	}

	next := func(ctx context.Context, req *tool.ExecutionRequest) (any, error) {
		return "result", nil
	}

	result, err := middleware.Execute(context.Background(), req, next)

	assert.NoError(t, err)
	assert.Equal(t, "result", result)

	spans := exporter.GetSpans()
	assert.Len(t, spans, 1)
	assert.Equal(t, "tool.execute", spans[0].Name)

	attrs := spans[0].Attributes
	var toolName string
	for _, attr := range attrs {
		if attr.Key == "tool.name" {
			toolName = attr.Value.AsString()
		}
	}
	assert.Equal(t, "test-tool", toolName)
}

func BenchmarkTracingMiddleware_Execute(b *testing.B) {
	// Setup no-op tracer for benchmark to measure middleware overhead only
	tp := trace.NewTracerProvider() // No exporter = no-op processing mostly
	otel.SetTracerProvider(tp)

	middleware := NewTracingMiddleware()
	req := &tool.ExecutionRequest{
		ToolName:   "bench-tool",
		ToolInputs: []byte(`{}`),
	}
	next := func(ctx context.Context, req *tool.ExecutionRequest) (any, error) {
		return nil, nil
	}
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = middleware.Execute(ctx, req, next)
	}
}
