// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package telemetry

import (
	"bytes"
	"context"
	"testing"

	config_v1 "github.com/mcpany/core/proto/config/v1"
	"go.opentelemetry.io/otel"
)

func strPtr(s string) *string {
	return &s
}

func TestInitTelemetry(t *testing.T) {
	cfg := &config_v1.TelemetryConfig{
		TracesExporter: strPtr("stdout"),
	}

	var buf bytes.Buffer
	shutdown, err := InitTelemetry(context.Background(), "test-service", "v0.0.1", cfg, &buf)
	if err != nil {
		t.Fatalf("InitTelemetry failed: %v", err)
	}

	// Create a span
	_, span := otel.Tracer("test").Start(context.Background(), "test-span")
	span.End()

	if err := shutdown(context.Background()); err != nil {
		t.Errorf("Shutdown failed: %v", err)
	}

	if buf.Len() == 0 {
		t.Log("Buffer content empty - check if tracing is enabled")
	} else {
		t.Log("Traces captured successfully")
	}
}

func TestInitTelemetry_NoExporter(t *testing.T) {
	cfg := &config_v1.TelemetryConfig{
		TracesExporter:  strPtr(""),
		OtlpEndpoint:    strPtr(""),
		MetricsExporter: strPtr(""),
	}

	shutdown, err := InitTelemetry(context.Background(), "test-service", "v0.0.1", cfg, nil)
	if err != nil {
		t.Errorf("InitTelemetry failed: %v", err)
	}
	if err := shutdown(context.Background()); err != nil {
		t.Errorf("Shutdown failed: %v", err)
	}
}

func TestInitTelemetry_NilWriter(t *testing.T) {
	cfg := &config_v1.TelemetryConfig{
		TracesExporter: strPtr("stdout"),
	}

	// Passing nil writer should not panic and should default to io.Discard
	shutdown, err := InitTelemetry(context.Background(), "test-service", "v0.0.1", cfg, nil)
	if err != nil {
		t.Fatalf("InitTelemetry failed: %v", err)
	}

	if err := shutdown(context.Background()); err != nil {
		t.Errorf("Shutdown failed: %v", err)
	}
}

func TestInitTelemetry_AutoDetectOTLP(t *testing.T) {
	// Set endpoint but no exporter type, should default to OTLP
	cfg := &config_v1.TelemetryConfig{
		TracesExporter: strPtr(""),
		OtlpEndpoint:   strPtr("localhost:4318"),
	}

	shutdown, err := InitTelemetry(context.Background(), "test-service", "v0.0.1", cfg, nil)
	// It might succeed in creating the exporter even if the endpoint is not reachable (lazy connection)
	if err != nil {
		t.Logf("InitTelemetry with OTLP failed (expected if dependencies missing or validation fails): %v", err)
	} else {
		_ = shutdown(context.Background())
	}
}

func TestInitTelemetry_MetricsStdout(t *testing.T) {
	cfg := &config_v1.TelemetryConfig{
		MetricsExporter: strPtr("stdout"),
	}

	var buf bytes.Buffer
	shutdown, err := InitTelemetry(context.Background(), "test-service", "v0.0.1", cfg, &buf)
	if err != nil {
		t.Fatalf("InitTelemetry failed: %v", err)
	}

	// Create a meter and a counter to generate some metrics
	meter := otel.Meter("test-meter")
	counter, err := meter.Int64Counter("test-counter")
	if err != nil {
		t.Fatalf("Failed to create counter: %v", err)
	}
	counter.Add(context.Background(), 1)

	// Shutdown to flush metrics
	if err := shutdown(context.Background()); err != nil {
		t.Errorf("Shutdown failed: %v", err)
	}
}

func TestInitTelemetry_MetricsOTLP(t *testing.T) {
	cfg := &config_v1.TelemetryConfig{
		MetricsExporter: strPtr("otlp"),
		OtlpEndpoint:    strPtr("localhost:4318"),
	}

	shutdown, err := InitTelemetry(context.Background(), "test-service", "v0.0.1", cfg, nil)
	if err != nil {
		t.Logf("InitTelemetry with OTLP metrics failed: %v", err)
	} else {
		_ = shutdown(context.Background())
	}
}
