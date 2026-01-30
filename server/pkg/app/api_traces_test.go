// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/mcpany/core/server/pkg/tracing"
	"github.com/mcpany/core/server/pkg/validation"
)

func setupTracesTestApp(t *testing.T) *Application {
	tempDir := t.TempDir()
	// Allow temp dir for audit logs
	validation.SetAllowedPaths([]string{tempDir})
	t.Cleanup(func() {
		validation.SetAllowedPaths(nil)
	})

	// Initialize Application with TraceExporter
	app := &Application{
		TraceExporter: tracing.NewInMemoryExporter(100),
	}

	return app
}

func TestHandleTraces_Empty(t *testing.T) {
	app := setupTracesTestApp(t)

	req := httptest.NewRequest("GET", "/traces", nil)
	w := httptest.NewRecorder()

	app.handleTraces().ServeHTTP(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status OK, got %v", resp.Status)
	}

	var traces []tracing.Trace
	if err := json.NewDecoder(resp.Body).Decode(&traces); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if len(traces) != 0 {
		t.Errorf("Expected 0 traces, got %d", len(traces))
	}
}

func TestHandleTraces_WithData(t *testing.T) {
	app := setupTracesTestApp(t)

	// Inject a trace
	now := time.Now()
	trace := &tracing.Trace{
		ID: "test-trace",
		RootSpan: &tracing.Span{
			Name: "test-tool",
			Status: "success",
			StartTime: now.UnixMilli(),
			EndTime: now.Add(100 * time.Millisecond).UnixMilli(),
		},
		Status: "success",
		Timestamp: now.Format(time.RFC3339),
	}
	app.TraceExporter.Seed([]*tracing.Trace{trace})

	httpReq := httptest.NewRequest("GET", "/traces", nil)
	w := httptest.NewRecorder()

	app.handleTraces().ServeHTTP(w, httpReq)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status OK, got %v", resp.Status)
	}

	var traces []tracing.Trace
	if err := json.NewDecoder(resp.Body).Decode(&traces); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if len(traces) != 1 {
		t.Fatalf("Expected 1 trace, got %d", len(traces))
	}

	tTrace := traces[0]
	if tTrace.RootSpan.Name != "test-tool" {
		t.Errorf("Expected tool name 'test-tool', got '%s'", tTrace.RootSpan.Name)
	}
	if tTrace.Status != "success" {
		t.Errorf("Expected status 'success', got '%s'", tTrace.Status)
	}
}
