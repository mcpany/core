// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mcpany/core/server/pkg/tool"
)

func TestHandleTraces_Limit(t *testing.T) {
	app, am := setupTracesTestApp(t)

	// Inject 10 audit entries
	ctx := context.Background()
	next := func(ctx context.Context, req *tool.ExecutionRequest) (any, error) {
		return "ok", nil
	}

	for i := 0; i < 10; i++ {
		req := &tool.ExecutionRequest{
			ToolName: fmt.Sprintf("tool-%d", i),
		}
		_, err := am.Execute(ctx, req, next)
		if err != nil {
			t.Fatalf("Failed to execute middleware: %v", err)
		}
	}

	// Case 1: No limit (should return all 10, reversed?)
	req := httptest.NewRequest("GET", "/traces", nil)
	w := httptest.NewRecorder()
	app.handleTraces().ServeHTTP(w, req)

    resp := w.Result()
    if resp.StatusCode != http.StatusOK {
        t.Errorf("Expected status OK, got %v", resp.Status)
    }

	var traces []Trace
	if err := json.NewDecoder(resp.Body).Decode(&traces); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if len(traces) != 10 {
		t.Errorf("Expected 10 traces, got %d", len(traces))
	}
	if traces[0].RootSpan.Name != "tool-9" {
		t.Errorf("Expected newest trace first (tool-9), got %s", traces[0].RootSpan.Name)
	}

	// Case 2: Limit 5
	req = httptest.NewRequest("GET", "/traces?limit=5", nil)
	w = httptest.NewRecorder()
	app.handleTraces().ServeHTTP(w, req)

    resp = w.Result()
    if resp.StatusCode != http.StatusOK {
        t.Errorf("Expected status OK, got %v", resp.Status)
    }

	var tracesLimit []Trace
	if err := json.NewDecoder(resp.Body).Decode(&tracesLimit); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if len(tracesLimit) != 5 {
		t.Errorf("Expected 5 traces, got %d", len(tracesLimit))
	}
	if tracesLimit[0].RootSpan.Name != "tool-9" {
		t.Errorf("Expected newest trace first (tool-9), got %s", tracesLimit[0].RootSpan.Name)
	}
	if tracesLimit[4].RootSpan.Name != "tool-5" {
		t.Errorf("Expected 5th trace to be tool-5, got %s", tracesLimit[4].RootSpan.Name)
	}
}
