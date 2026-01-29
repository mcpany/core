// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/middleware"
	"github.com/mcpany/core/server/pkg/tool"
	"github.com/mcpany/core/server/pkg/validation"
	"google.golang.org/protobuf/proto"
)

func setupTracesTestApp(t *testing.T) (*Application, *middleware.AuditMiddleware) {
	tempDir := t.TempDir()
	// Allow temp dir for audit logs
	validation.SetAllowedPaths([]string{tempDir})
	t.Cleanup(func() {
		validation.SetAllowedPaths(nil)
	})

	// Initialize Audit Middleware
	storageType := configv1.AuditConfig_STORAGE_TYPE_FILE
	auditConfig := configv1.AuditConfig_builder{
		Enabled:      proto.Bool(true),
		StorageType:  &storageType,
		OutputPath:   proto.String(filepath.Join(tempDir, "audit.log")),
		LogArguments: proto.Bool(true),
		LogResults:   proto.Bool(true),
	}.Build()

	auditMiddleware, err := middleware.NewAuditMiddleware(auditConfig)
	if err != nil {
		t.Fatalf("Failed to create audit middleware: %v", err)
	}

	// Initialize Standard Middlewares
	standardMiddlewares := &middleware.StandardMiddlewares{
		Audit: auditMiddleware,
	}

	// Initialize Application
	app := &Application{
		standardMiddlewares: standardMiddlewares,
	}

	return app, auditMiddleware
}

func TestHandleTraces_Empty(t *testing.T) {
	app, _ := setupTracesTestApp(t)

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

	if len(traces) != 0 {
		t.Errorf("Expected 0 traces, got %d", len(traces))
	}
}

func TestHandleTraces_WithData(t *testing.T) {
	app, am := setupTracesTestApp(t)

	// Inject an audit entry
	ctx := context.Background()
	inputs, _ := json.Marshal(map[string]interface{}{
		"arg1": "value1",
	})
	req := &tool.ExecutionRequest{
		ToolName:   "test-tool",
		ToolInputs: json.RawMessage(inputs),
	}
	next := func(ctx context.Context, req *tool.ExecutionRequest) (any, error) {
		return map[string]interface{}{"result": "success"}, nil
	}

	// Execute to generate log
	_, err := am.Execute(ctx, req, next)
	if err != nil {
		t.Fatalf("Failed to execute middleware: %v", err)
	}

	httpReq := httptest.NewRequest("GET", "/traces", nil)
	w := httptest.NewRecorder()

	app.handleTraces().ServeHTTP(w, httpReq)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status OK, got %v", resp.Status)
	}

	var traces []Trace
	if err := json.NewDecoder(resp.Body).Decode(&traces); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if len(traces) != 1 {
		t.Fatalf("Expected 1 trace, got %d", len(traces))
	}

	trace := traces[0]
	if trace.RootSpan.Name != "test-tool" {
		t.Errorf("Expected tool name 'test-tool', got '%s'", trace.RootSpan.Name)
	}
	if trace.Status != "success" {
		t.Errorf("Expected status 'success', got '%s'", trace.Status)
	}
}

func TestHandleTracesWS(t *testing.T) {
	app, am := setupTracesTestApp(t)

	// Inject initial data
	ctx := context.Background()
	req1 := &tool.ExecutionRequest{ToolName: "initial-tool"}
	next := func(ctx context.Context, req *tool.ExecutionRequest) (any, error) {
		return "ok", nil
	}
	_, _ = am.Execute(ctx, req1, next)

	// Start Test Server
	server := httptest.NewServer(app.handleTracesWS())
	defer server.Close()

	// Convert http URL to ws URL
	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")

	// Connect
	ws, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("Failed to connect to websocket: %v", err)
	}
	defer ws.Close()

	// 1. Verify History (Should receive the initial tool)
	var trace1 Trace
	if err := ws.ReadJSON(&trace1); err != nil {
		t.Fatalf("Failed to read history trace: %v", err)
	}
	if trace1.RootSpan.Name != "initial-tool" {
		t.Errorf("Expected initial tool 'initial-tool', got '%s'", trace1.RootSpan.Name)
	}

	// 2. Verify Real-time Update
	// Inject new data
	req2 := &tool.ExecutionRequest{ToolName: "realtime-tool"}
	_, _ = am.Execute(ctx, req2, next)

	// Read next message
	ws.SetReadDeadline(time.Now().Add(1 * time.Second))
	var trace2 Trace
	if err := ws.ReadJSON(&trace2); err != nil {
		t.Fatalf("Failed to read realtime trace: %v", err)
	}
	if trace2.RootSpan.Name != "realtime-tool" {
		t.Errorf("Expected realtime tool 'realtime-tool', got '%s'", trace2.RootSpan.Name)
	}
}

func TestHandleTraces_DisabledAudit(t *testing.T) {
	// Setup app with nil audit middleware
	app := &Application{
		standardMiddlewares: &middleware.StandardMiddlewares{
			Audit: nil,
		},
	}

	req := httptest.NewRequest("GET", "/traces", nil)
	w := httptest.NewRecorder()

	app.handleTraces().ServeHTTP(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status OK, got %v", resp.Status)
	}

	body := w.Body.String()
	if body != "[]" {
		t.Errorf("Expected '[]', got '%s'", body)
	}
}

func TestHandleTracesWS_DisabledAudit(t *testing.T) {
	// Setup app with nil audit middleware
	app := &Application{
		standardMiddlewares: &middleware.StandardMiddlewares{
			Audit: nil,
		},
	}

	server := httptest.NewServer(app.handleTracesWS())
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")

	ws, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("Failed to connect to websocket: %v", err)
	}
	defer ws.Close()

	// Should receive a close message or closure
	ws.SetReadDeadline(time.Now().Add(1 * time.Second))
	_, _, err = ws.ReadMessage()
	if err == nil {
		t.Errorf("Expected connection closure, got a message")
	}
}
