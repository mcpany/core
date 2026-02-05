// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

// Package telemetry provides functions for initializing and managing telemetry.
package telemetry

import (
	"context"
	"fmt"
	"io"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/exporters/stdout/stdoutmetric"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"

	config_v1 "github.com/mcpany/core/proto/config/v1"
)

const (
	exporterOTLP   = "otlp"
	exporterStdout = "stdout"
	exporterNone   = "none"
)

// InitTelemetry initializes OpenTelemetry tracing and metrics.
// It writes traces/metrics to the provided writer (e.g., os.Stderr) if stdout exporter is selected.
// It returns a shutdown function that should be called when the application exits.
//
// Summary: Initializes OpenTelemetry tracing and metrics subsystems.
//
// Parameters:
//   - ctx: context.Context. The context for the initialization.
//   - serviceName: string. The name of the service for telemetry identification.
//   - version: string. The version of the service.
//   - cfg: *config_v1.TelemetryConfig. The telemetry configuration object.
//   - writer: io.Writer. The writer to use for stdout export (if configured).
//
// Returns:
//   - func(context.Context) error: A cleanup function to shut down telemetry providers.
//   - error: An error if initialization fails.
func InitTelemetry(ctx context.Context, serviceName string, version string, cfg *config_v1.TelemetryConfig, writer io.Writer) (func(context.Context) error, error) {
	// If writer is nil, discard output
	if writer == nil {
		writer = io.Discard
	}

	if cfg == nil {
		cfg = &config_v1.TelemetryConfig{}
	}
	// Allow service name override from config
	if cfg.GetServiceName() != "" {
		serviceName = cfg.GetServiceName()
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

	// Initialize Tracer
	shutdownTracer, err := initTracer(ctx, res, cfg, writer)
	if err != nil {
		return nil, fmt.Errorf("failed to init tracer: %w", err)
	}

	// Initialize Meter
	shutdownMeter, err := initMeter(ctx, res, cfg, writer)
	if err != nil {
		_ = shutdownTracer(ctx) // Attempt to shutdown tracer on failure
		return nil, fmt.Errorf("failed to init meter: %w", err)
	}

	return func(ctx context.Context) error {
		err1 := shutdownTracer(ctx)
		err2 := shutdownMeter(ctx)
		if err1 != nil {
			return err1
		}
		return err2
	}, nil
}

func initTracer(ctx context.Context, res *resource.Resource, cfg *config_v1.TelemetryConfig, writer io.Writer) (func(context.Context) error, error) {
	exporterType := cfg.GetTracesExporter()
	// If OTLP endpoint is set, default to otlp if type not specified
	if cfg.GetOtlpEndpoint() != "" && (exporterType == "" || exporterType == exporterOTLP) {
		exporterType = exporterOTLP
	}

	var exporter trace.SpanExporter
	var err error

	switch exporterType {
	case exporterOTLP:
		opts := []otlptracehttp.Option{
			otlptracehttp.WithInsecure(),
		}
		if cfg.GetOtlpEndpoint() != "" {
			opts = append(opts, otlptracehttp.WithEndpoint(cfg.GetOtlpEndpoint()))
		}
		exporter, err = otlptracehttp.New(ctx, opts...)
		if err != nil {
			return nil, fmt.Errorf("failed to create otlp trace exporter: %w", err)
		}
	case exporterStdout:
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

	tp := trace.NewTracerProvider(
		trace.WithBatcher(exporter),
		trace.WithResource(res),
	)

	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))

	return tp.Shutdown, nil
}

func initMeter(ctx context.Context, res *resource.Resource, cfg *config_v1.TelemetryConfig, _ io.Writer) (func(context.Context) error, error) {
	exporterType := cfg.GetMetricsExporter()
	// If OTLP endpoint is set, default to otlp if type not specified
	if cfg.GetOtlpEndpoint() != "" && (exporterType == "" || exporterType == exporterOTLP) {
		exporterType = exporterOTLP
	}

	var exporter metric.Reader
	var err error

	switch exporterType {
	case exporterOTLP:
		var exp metric.Exporter
		opts := []otlpmetrichttp.Option{
			otlpmetrichttp.WithInsecure(),
		}
		if cfg.GetOtlpEndpoint() != "" {
			opts = append(opts, otlpmetrichttp.WithEndpoint(cfg.GetOtlpEndpoint()))
		}
		exp, err = otlpmetrichttp.New(ctx, opts...)
		if err != nil {
			return nil, fmt.Errorf("failed to create otlp metric exporter: %w", err)
		}
		exporter = metric.NewPeriodicReader(exp)
	case exporterStdout:
		var exp metric.Exporter
		exp, err = stdoutmetric.New(stdoutmetric.WithPrettyPrint())
		if err != nil {
			return nil, fmt.Errorf("failed to create stdout metric exporter: %w", err)
		}
		// Register stdout exporter
		exporter = metric.NewPeriodicReader(exp)
	default:
		return func(context.Context) error { return nil }, nil
	}

	mp := metric.NewMeterProvider(
		metric.WithReader(exporter),
		metric.WithResource(res),
	)
	otel.SetMeterProvider(mp)

	return mp.Shutdown, nil
}
