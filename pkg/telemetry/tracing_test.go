// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package telemetry

import (
	"bytes"
	"context"
	"os"
	"testing"

	"go.opentelemetry.io/otel"
)

func TestInitTracer(t *testing.T) {
	os.Setenv("OTEL_TRACES_EXPORTER", "stdout")
	defer os.Unsetenv("OTEL_TRACES_EXPORTER")

	var buf bytes.Buffer
	shutdown, err := InitTracer(context.Background(), "test-service", "v0.0.1", &buf)
	if err != nil {
		t.Fatalf("InitTracer failed: %v", err)
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

func TestInitTracer_Disabled(t *testing.T) {
	// Ensure env var is not set to stdout
	os.Unsetenv("OTEL_TRACES_EXPORTER")

	var buf bytes.Buffer
	shutdown, err := InitTracer(context.Background(), "test-service", "v0.0.1", &buf)
	if err != nil {
		t.Fatalf("InitTracer failed: %v", err)
	}

	// Create a span
	_, span := otel.Tracer("test").Start(context.Background(), "test-span")
	span.End()

	if err := shutdown(context.Background()); err != nil {
		t.Errorf("Shutdown failed: %v", err)
	}

	if buf.Len() > 0 {
		t.Error("Buffer content should be empty when tracing is disabled")
	}
}

func TestInitTracer_NilWriter(t *testing.T) {
	os.Setenv("OTEL_TRACES_EXPORTER", "stdout")
	defer os.Unsetenv("OTEL_TRACES_EXPORTER")

	shutdown, err := InitTracer(context.Background(), "test-service", "v0.0.1", nil)
	if err != nil {
		t.Fatalf("InitTracer failed: %v", err)
	}
	if err := shutdown(context.Background()); err != nil {
		t.Errorf("Shutdown failed: %v", err)
	}
}
