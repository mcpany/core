// Package telemetry provides functions for initializing and managing telemetry.

package telemetry

import (
	"context"
	"fmt"
	"io"
	"os"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
)

const exporterOTLP = "otlp"

// InitTracer initializes the OpenTelemetry tracer provider.
// It writes traces to the provided writer (e.g., os.Stderr).
// It returns a shutdown function that should be called when the application exits.
func InitTracer(ctx context.Context, serviceName string, version string, writer io.Writer) (func(context.Context) error, error) {
	// If writer is nil, discard output
	if writer == nil {
		writer = io.Discard
	}

	exporterType := os.Getenv("OTEL_TRACES_EXPORTER")
	// If OTEL_EXPORTER_OTLP_ENDPOINT is set, default to otlp if type not specified or is otlp
	if os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT") != "" && (exporterType == "" || exporterType == exporterOTLP) {
		exporterType = exporterOTLP
	}

	var exporter trace.SpanExporter
	var err error

	switch exporterType {
	case exporterOTLP:
		// Use OTLP HTTP Exporter
		// Options are automatically read from env vars:
		// OTEL_EXPORTER_OTLP_ENDPOINT, OTEL_EXPORTER_OTLP_HEADERS, etc.
		exporter, err = otlptracehttp.New(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to create otlp trace exporter: %w", err)
		}
	case "stdout":
		exporter, err = stdouttrace.New(
			stdouttrace.WithWriter(writer),
			stdouttrace.WithPrettyPrint(),
		)
		if err != nil {
			return nil, fmt.Errorf("failed to create stdout trace exporter: %w", err)
		}
	default:
		return func(context.Context) error { return nil }, nil
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
