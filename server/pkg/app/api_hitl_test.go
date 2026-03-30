// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/mcpany/core/proto/bus"
	corebus "github.com/mcpany/core/server/pkg/bus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestBus(t *testing.T) *corebus.Provider {
	busConfig := &bus.MessageBus{}
	busConfig.SetInMemory(&bus.InMemoryBus{})
	p, err := corebus.NewProvider(busConfig)
	require.NoError(t, err)
	return p
}

func TestMountHITL_SubscribeRequest(t *testing.T) {
	bp := setupTestBus(t)
	app := &Application{
		busProvider: bp,
	}
	mux := http.NewServeMux()

	// Reset global state
	globalHITLState.mu.Lock()
	globalHITLState.approvals = make(map[string]hitlApprovalRequest)
	globalHITLState.mu.Unlock()

	app.mountHITL(mux)

	// Publish a request
	reqBus, err := corebus.GetBus[hitlApprovalRequest](bp, "hitl.requests")
	require.NoError(t, err)

	err = reqBus.Publish(context.Background(), "hitl.requests", hitlApprovalRequest{
		ExecutionID: "test-id",
		ToolName:    "test-tool",
		RequireMFA:  true,
	})
	require.NoError(t, err)

	// Wait for processing
	assert.Eventually(t, func() bool {
		globalHITLState.mu.RLock()
		defer globalHITLState.mu.RUnlock()
		_, ok := globalHITLState.approvals["test-id"]
		return ok
	}, time.Second, 10*time.Millisecond)
}

func TestMountHITL_GetApprovals(t *testing.T) {
	bp := setupTestBus(t)
	app := &Application{
		busProvider: bp,
	}
	mux := http.NewServeMux()
	app.mountHITL(mux)

	globalHITLState.mu.Lock()
	globalHITLState.approvals = map[string]hitlApprovalRequest{
		"test-id-1": {ExecutionID: "test-id-1", ToolName: "test-tool-1", RequireMFA: true},
	}
	globalHITLState.mu.Unlock()

	req := httptest.NewRequest(http.MethodGet, "/hitl/approvals", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	var list []map[string]interface{}
	err := json.Unmarshal(rr.Body.Bytes(), &list)
	require.NoError(t, err)

	assert.Len(t, list, 1)
	assert.Equal(t, "test-id-1", list[0]["id"])
	assert.Equal(t, "test-tool-1", list[0]["tool"])
	assert.Equal(t, "pending", list[0]["status"])
	assert.Equal(t, true, list[0]["requireMfa"])
}

func TestMountHITL_PostApproval(t *testing.T) {
	bp := setupTestBus(t)
	app := &Application{
		busProvider: bp,
	}
	mux := http.NewServeMux()
	app.mountHITL(mux)

	globalHITLState.mu.Lock()
	globalHITLState.approvals = map[string]hitlApprovalRequest{
		"test-id-2": {ExecutionID: "test-id-2", ToolName: "test-tool-2", RequireMFA: true},
	}
	globalHITLState.mu.Unlock()

	// Capture response
	resBus, err := corebus.GetBus[hitlApprovalResponse](bp, "hitl.responses.test-id-2")
	require.NoError(t, err)

	resChan := make(chan hitlApprovalResponse, 1)
	resBus.Subscribe(context.Background(), "hitl.responses.test-id-2", func(res hitlApprovalResponse) {
		resChan <- res
	})

	body := []byte(`{"action": "approved", "mfaCode": "123456"}`)
	req := httptest.NewRequest(http.MethodPost, "/hitl/approvals/test-id-2", bytes.NewReader(body))
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	select {
	case res := <-resChan:
		assert.True(t, res.Approved)
		assert.Equal(t, "test-id-2", res.ExecutionID)
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for response")
	}

	globalHITLState.mu.RLock()
	_, ok := globalHITLState.approvals["test-id-2"]
	globalHITLState.mu.RUnlock()
	assert.False(t, ok, "should be removed from pending")
}

func TestMountHITL_PostDeny(t *testing.T) {
	bp := setupTestBus(t)
	app := &Application{
		busProvider: bp,
	}
	mux := http.NewServeMux()
	app.mountHITL(mux)

	globalHITLState.mu.Lock()
	globalHITLState.approvals = map[string]hitlApprovalRequest{
		"test-id-3": {ExecutionID: "test-id-3", ToolName: "test-tool-3", RequireMFA: false},
	}
	globalHITLState.mu.Unlock()

	// Capture response
	resBus, err := corebus.GetBus[hitlApprovalResponse](bp, "hitl.responses.test-id-3")
	require.NoError(t, err)

	resChan := make(chan hitlApprovalResponse, 1)
	resBus.Subscribe(context.Background(), "hitl.responses.test-id-3", func(res hitlApprovalResponse) {
		resChan <- res
	})

	body := []byte(`{"action": "denied", "mfaCode": ""}`)
	req := httptest.NewRequest(http.MethodPost, "/hitl/approvals/test-id-3", bytes.NewReader(body))
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	select {
	case res := <-resChan:
		assert.False(t, res.Approved)
		assert.Equal(t, "test-id-3", res.ExecutionID)
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for response")
	}

	globalHITLState.mu.RLock()
	_, ok := globalHITLState.approvals["test-id-3"]
	globalHITLState.mu.RUnlock()
	assert.False(t, ok, "should be removed from pending")
}

func TestMountHITL_MethodNotAllowed(t *testing.T) {
	bp := setupTestBus(t)
	app := &Application{
		busProvider: bp,
	}
	mux := http.NewServeMux()
	app.mountHITL(mux)

	// Test GET /hitl/approvals/
	req := httptest.NewRequest(http.MethodGet, "/hitl/approvals/test-id", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusMethodNotAllowed, rr.Code)

	// Test POST /hitl/approvals
	req2 := httptest.NewRequest(http.MethodPost, "/hitl/approvals", bytes.NewReader([]byte{}))
	rr2 := httptest.NewRecorder()
	mux.ServeHTTP(rr2, req2)
	assert.Equal(t, http.StatusMethodNotAllowed, rr2.Code)
}

func TestMountHITL_BadRequest(t *testing.T) {
	bp := setupTestBus(t)
	app := &Application{
		busProvider: bp,
	}
	mux := http.NewServeMux()
	app.mountHITL(mux)

	// Test POST with invalid JSON
	req := httptest.NewRequest(http.MethodPost, "/hitl/approvals/test", bytes.NewReader([]byte("{invalid}")))
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestMountHITL_NilBusProvider(t *testing.T) {
	app := &Application{
		busProvider: nil,
	}
	mux := http.NewServeMux()
	app.mountHITL(mux)

	// Should not panic
}
