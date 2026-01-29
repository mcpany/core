// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/audit"
	"github.com/mcpany/core/server/pkg/middleware"
	"github.com/mcpany/core/server/pkg/tool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestTraceMockAuditStore implements audit.Store for testing
type TestTraceMockAuditStore struct {
	mu      sync.Mutex
	Entries []audit.Entry
}

func (m *TestTraceMockAuditStore) Write(ctx context.Context, entry audit.Entry) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Entries = append(m.Entries, entry)
	return nil
}

func (m *TestTraceMockAuditStore) Read(ctx context.Context, filter audit.Filter) ([]audit.Entry, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	// Return a copy to avoid race conditions if caller modifies it (though slice is by ref, elements are structs)
	// But Read returns slice, so append to m.Entries later might realloc.
	// Copy is safer.
	result := make([]audit.Entry, len(m.Entries))
	copy(result, m.Entries)
	return result, nil
}

func (m *TestTraceMockAuditStore) Close() error {
	return nil
}

func TestHandleTraces(t *testing.T) {
	// Setup Application
	app := &Application{}

	// Setup Audit Middleware
	auditConfig := &configv1.AuditConfig{}
	auditConfig.SetEnabled(true)
	auditConfig.SetLogArguments(true)
	auditConfig.SetLogResults(true)

	auditMiddleware, err := middleware.NewAuditMiddleware(auditConfig)
	require.NoError(t, err)
	mockStore := &TestTraceMockAuditStore{}
	auditMiddleware.SetStore(mockStore)

	// Inject middleware into app
	app.standardMiddlewares = &middleware.StandardMiddlewares{
		Audit: auditMiddleware,
	}

	// Create some audit entries via middleware
	ctx := context.Background()
	inputs := map[string]interface{}{"key": "value"}
	inputsBytes, _ := json.Marshal(inputs)
	req := &tool.ExecutionRequest{
		ToolName:   "test-tool",
		ToolInputs: json.RawMessage(inputsBytes),
	}

	// Invoke middleware to generate log
	_, err = auditMiddleware.Execute(ctx, req, func(ctx context.Context, req *tool.ExecutionRequest) (any, error) {
		return map[string]interface{}{"result": "success"}, nil
	})
	require.NoError(t, err)

	// Test GET /api/v1/traces
	handler := app.handleTraces()

	reqHTTP := httptest.NewRequest("GET", "/api/v1/traces", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, reqHTTP)

	assert.Equal(t, http.StatusOK, w.Code)

	var traces []*Trace
	err = json.Unmarshal(w.Body.Bytes(), &traces)
	require.NoError(t, err)

	require.Len(t, traces, 1)
	assert.Equal(t, "test-tool", traces[0].RootSpan.Name)
	assert.Equal(t, "success", traces[0].RootSpan.Status)
}

func TestHandleTracesWS(t *testing.T) {
	app := &Application{}
	auditConfig := &configv1.AuditConfig{}
	auditConfig.SetEnabled(true)

	auditMiddleware, err := middleware.NewAuditMiddleware(auditConfig)
	require.NoError(t, err)
	auditMiddleware.SetStore(&TestTraceMockAuditStore{})

	app.standardMiddlewares = &middleware.StandardMiddlewares{
		Audit: auditMiddleware,
	}

	// Generate an initial trace
	ctx := context.Background()
	inputs := map[string]interface{}{}
	inputsBytes, _ := json.Marshal(inputs)
	req := &tool.ExecutionRequest{
		ToolName:   "initial-tool",
		ToolInputs: json.RawMessage(inputsBytes),
	}
	_, err = auditMiddleware.Execute(ctx, req, func(ctx context.Context, req *tool.ExecutionRequest) (any, error) {
		return nil, nil
	})
	require.NoError(t, err)

	// Start Test Server
	server := httptest.NewServer(app.handleTracesWS())
	defer server.Close()

	// Convert http URL to ws URL
	url := "ws" + strings.TrimPrefix(server.URL, "http")

	// Connect WS
	ws, _, err := websocket.DefaultDialer.Dial(url, nil)
	require.NoError(t, err)
	defer ws.Close()

	// Read initial history
	var trace Trace
	err = ws.ReadJSON(&trace)
	require.NoError(t, err)
	assert.Equal(t, "initial-tool", trace.RootSpan.Name)

	// Trigger new event
	go func() {
		time.Sleep(100 * time.Millisecond)
		req2 := &tool.ExecutionRequest{
			ToolName:   "new-tool",
			ToolInputs: json.RawMessage([]byte("{}")),
		}
		_, err := auditMiddleware.Execute(ctx, req2, func(ctx context.Context, req *tool.ExecutionRequest) (any, error) {
			return nil, nil
		})
		// We are in a goroutine, using t.Error/Log might be racy if t ends, but sleep ensures it runs during test.
		// However, t.Log is thread safe.
		if err != nil {
			t.Logf("Failed to execute middleware in background: %v", err)
		}
	}()

	// Read new event
	err = ws.ReadJSON(&trace)
	require.NoError(t, err)
	assert.Equal(t, "new-tool", trace.RootSpan.Name)
}
