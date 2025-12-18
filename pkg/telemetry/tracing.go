// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package telemetry

import (
	"context"
	"fmt"
	"io"
	"os"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
)

// InitTracer initializes the OpenTelemetry tracer provider.
// It writes traces to the provided writer (e.g., os.Stderr).
// It returns a shutdown function that should be called when the application exits.
func InitTracer(ctx context.Context, serviceName string, version string, writer io.Writer) (func(context.Context) error, error) {
	// If writer is nil, discard output
	if writer == nil {
		writer = io.Discard
	}

	// We use stdout trace exporter for demonstration/MVP.
	// In a real environment, you would check for OTEL_EXPORTER_OTLP_ENDPOINT and use otlptrace.
	// We check if OTEL_TRACES_EXPORTER is explicitly set to "stdout" to enable it.
	// Default is disabled to avoid log pollution.
	if os.Getenv("OTEL_TRACES_EXPORTER") != "stdout" {
		return func(context.Context) error { return nil }, nil
	}

	exporter, err := stdouttrace.New(
		stdouttrace.WithWriter(writer),
		stdouttrace.WithPrettyPrint(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create stdout trace exporter: %w", err)
	}

	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceName(serviceName),
			semconv.ServiceVersion(version),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	tp := trace.NewTracerProvider(
		trace.WithBatcher(exporter),
		trace.WithResource(res),
	)

	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))

	return tp.Shutdown, nil
}
