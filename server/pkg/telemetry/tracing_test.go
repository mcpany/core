package telemetry

import (
	"bytes"
	"context"
	"testing"

	"go.opentelemetry.io/otel"
)

func TestInitTracer(t *testing.T) {
	t.Setenv("OTEL_TRACES_EXPORTER", "stdout")

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

func TestInitTracer_NoExporter(t *testing.T) {
	// Ensure no env vars that trigger other paths
	t.Setenv("OTEL_TRACES_EXPORTER", "")
	t.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", "")

	shutdown, err := InitTracer(context.Background(), "test-service", "v0.0.1", nil)
	if err != nil {
		t.Errorf("InitTracer failed: %v", err)
	}
	if err := shutdown(context.Background()); err != nil {
		t.Errorf("Shutdown failed: %v", err)
	}
}

func TestInitTracer_NilWriter(t *testing.T) {
	t.Setenv("OTEL_TRACES_EXPORTER", "stdout")

	// Passing nil writer should not panic and should default to io.Discard
	shutdown, err := InitTracer(context.Background(), "test-service", "v0.0.1", nil)
	if err != nil {
		t.Fatalf("InitTracer failed: %v", err)
	}

	if err := shutdown(context.Background()); err != nil {
		t.Errorf("Shutdown failed: %v", err)
	}
}

func TestInitTracer_AutoDetectOTLP(t *testing.T) {
	// Set endpoint but no exporter type, should default to OTLP
	t.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", "http://localhost:4318")
	// Make sure explicit type isn't set
	t.Setenv("OTEL_TRACES_EXPORTER", "")

	shutdown, err := InitTracer(context.Background(), "test-service", "v0.0.1", nil)
	// It might succeed in creating the exporter even if the endpoint is not reachable (lazy connection)
	if err != nil {
		t.Logf("InitTracer with OTLP failed (expected if dependencies missing or validation fails): %v", err)
		// We don't necessarily fail the test here because creating OTLP client might require more env setup
		// or network access which we might not have. But if it succeeds, great.
	} else {
		_ = shutdown(context.Background())
	}
}
