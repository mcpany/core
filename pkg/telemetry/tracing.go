// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package telemetry

import (
	"context"
	"fmt"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
)

// InitTracer initializes the global OpenTelemetry tracer provider.
// It returns a shutdown function that should be called when the application exits.
func InitTracer(ctx context.Context, cfg *configv1.TracingConfig) (func(context.Context) error, error) {
	if cfg == nil || !cfg.GetEnabled() {
		return func(context.Context) error { return nil }, nil
	}

	endpoint := cfg.GetEndpoint()
	if endpoint == "" {
		endpoint = "localhost:4317"
	}

	opts := []otlptracegrpc.Option{
		otlptracegrpc.WithEndpoint(endpoint),
	}

	if cfg.GetInsecure() {
		opts = append(opts, otlptracegrpc.WithInsecure())
	}

	exporter, err := otlptracegrpc.New(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create OTLP trace exporter: %w", err)
	}

	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceName("mcpany"),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
	)

	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	return tp.Shutdown, nil
}
