// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/middleware"
	"github.com/mcpany/core/server/pkg/tool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// ensure MockAuditStore is available (defined in api_audit_test.go)

func setupTraceTestApp(t *testing.T) (*Application, *middleware.AuditMiddleware, *MockAuditStore) {
	app := NewApplication()
	mockStore := new(MockAuditStore)

	// Initialize middleware
	auditConfig := &configv1.AuditConfig{}
	auditConfig.SetEnabled(true)
	auditConfig.SetLogArguments(true)
	auditConfig.SetLogResults(true)

	am, err := middleware.NewAuditMiddleware(auditConfig)
	require.NoError(t, err)
	am.SetStore(mockStore)

	app.standardMiddlewares = &middleware.StandardMiddlewares{
		Audit: am,
	}

	return app, am, mockStore
}

func TestHandleTraces_Get(t *testing.T) {
	app, am, mockStore := setupTraceTestApp(t)

	// Mock store write to succeed
	mockStore.On("Write", mock.Anything, mock.Anything).Return(nil)

	// Trigger an audit event to populate history
	ctx := context.Background()
	req := &tool.ExecutionRequest{
		ToolName:   "test-tool",
		ToolInputs: []byte(`{"arg":"val"}`),
	}

	// Execute middleware to generate audit log
	_, err := am.Execute(ctx, req, func(ctx context.Context, r *tool.ExecutionRequest) (any, error) {
		return map[string]string{"result": "ok"}, nil
	})
	require.NoError(t, err)

	// Wait a bit for broadcast to happen (it's synchronous in writeLog but good to be safe)

	reqHTTP := httptest.NewRequest(http.MethodGet, "/traces", nil)
	w := httptest.NewRecorder()

	app.handleTraces()(w, reqHTTP)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

	var traces []*Trace
	err = json.Unmarshal(w.Body.Bytes(), &traces)
	require.NoError(t, err)

	require.Len(t, traces, 1)
	trace := traces[0]
	assert.Equal(t, "test-tool", trace.RootSpan.Name)
	assert.Equal(t, "success", trace.Status)

	// Check Input
	assert.Equal(t, "val", trace.RootSpan.Input["arg"])

	// Check Output
	assert.Equal(t, "ok", trace.RootSpan.Output["result"])
}

func TestHandleTraces_Get_Error(t *testing.T) {
	app, am, mockStore := setupTraceTestApp(t)

	mockStore.On("Write", mock.Anything, mock.Anything).Return(nil)

	ctx := context.Background()
	req := &tool.ExecutionRequest{
		ToolName: "fail-tool",
	}

	// Execute with error
	_, err := am.Execute(ctx, req, func(ctx context.Context, r *tool.ExecutionRequest) (any, error) {
		return nil, assert.AnError
	})
	require.Error(t, err)

	reqHTTP := httptest.NewRequest(http.MethodGet, "/traces", nil)
	w := httptest.NewRecorder()

	app.handleTraces()(w, reqHTTP)

	var traces []*Trace
	err = json.Unmarshal(w.Body.Bytes(), &traces)
	require.NoError(t, err)

	require.Len(t, traces, 1)
	trace := traces[0]
	assert.Equal(t, "fail-tool", trace.RootSpan.Name)
	assert.Equal(t, "error", trace.Status)
	assert.Equal(t, assert.AnError.Error(), trace.RootSpan.ErrorMessage)
}

func TestHandleTraces_WS(t *testing.T) {
	app, am, mockStore := setupTraceTestApp(t)
	mockStore.On("Write", mock.Anything, mock.Anything).Return(nil)

	// 1. Populate some history
	ctx := context.Background()
	req1 := &tool.ExecutionRequest{ToolName: "tool-1", ToolInputs: []byte("{}")}
	_, _ = am.Execute(ctx, req1, func(ctx context.Context, r *tool.ExecutionRequest) (any, error) {
		return nil, nil
	})

	// 2. Start Test Server
	s := httptest.NewServer(app.handleTracesWS())
	defer s.Close()

	// Convert http URL to ws URL
	u := "ws" + strings.TrimPrefix(s.URL, "http")

	// 3. Connect WebSocket
	ws, _, err := websocket.DefaultDialer.Dial(u, nil)
	require.NoError(t, err)
	defer ws.Close()

	// 4. Read History (should get tool-1)
	var trace1 Trace
	err = ws.ReadJSON(&trace1)
	require.NoError(t, err)
	assert.Equal(t, "tool-1", trace1.RootSpan.Name)

	// 5. Trigger new event
	req2 := &tool.ExecutionRequest{ToolName: "tool-2", ToolInputs: []byte("{}")}
	go func() {
		// Needs to run in goroutine? Execute is synchronous, but Broadcast pushes to channel.
		// If channel buffer is full, it might block?
		// Broadcaster uses non-blocking send or large buffer?
		// logging/broadcaster.go: Broadcast is usually non-blocking or managed.
		// But let's run in goroutine to be safe, though not strictly necessary if designed well.
		time.Sleep(100 * time.Millisecond)
		_, _ = am.Execute(ctx, req2, func(ctx context.Context, r *tool.ExecutionRequest) (any, error) {
			return nil, nil
		})
	}()

	// 6. Read new event
	var trace2 Trace
	err = ws.ReadJSON(&trace2)
	require.NoError(t, err)
	assert.Equal(t, "tool-2", trace2.RootSpan.Name)
}

func TestHandleTraces_AuditDisabled(t *testing.T) {
	app := NewApplication()
	// No middleware set

	// GET
	reqHTTP := httptest.NewRequest(http.MethodGet, "/traces", nil)
	w := httptest.NewRecorder()
	app.handleTraces()(w, reqHTTP)
	assert.Equal(t, "[]", w.Body.String())

	// WS
	s := httptest.NewServer(app.handleTracesWS())
	defer s.Close()
	u := "ws" + strings.TrimPrefix(s.URL, "http")

	_, _, err := websocket.DefaultDialer.Dial(u, nil)
	// Dial might succeed but connection closed immediately with CloseMessage
	// OR Dial fails if server closes immediately?
	// Usually Dial succeeds, then Read returns error.

	if err == nil {
		ws, _, err := websocket.DefaultDialer.Dial(u, nil)
		if err == nil {
			defer ws.Close()
			_, _, err = ws.ReadMessage()
			assert.Error(t, err, "Should get error reading from closed connection")
			// Check close error?
			if closeErr, ok := err.(*websocket.CloseError); ok {
				assert.Equal(t, websocket.CloseNormalClosure, closeErr.Code)
				assert.Equal(t, "Audit disabled", closeErr.Text)
			}
		}
	}
}
