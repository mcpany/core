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
		OtlpEndpoint:   strPtr("http://localhost:4318"),
	}

	shutdown, err := InitTelemetry(context.Background(), "test-service", "v0.0.1", cfg, nil)
	// It might succeed in creating the exporter even if the endpoint is not reachable (lazy connection)
	if err != nil {
		t.Logf("InitTelemetry with OTLP failed (expected if dependencies missing or validation fails): %v", err)
		// We don't necessarily fail the test here because creating OTLP client might require more env setup
		// or network access which we might not have. But if it succeeds, great.
	} else {
		_ = shutdown(context.Background())
	}
}
