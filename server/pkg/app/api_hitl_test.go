package app

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/mcpany/core/server/pkg/bus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMountHITL_GetApprovals(t *testing.T) {
	// Reset global state
	globalHITLState.mu.Lock()
	globalHITLState.approvals = make(map[string]hitlApprovalRequest)
	globalHITLState.approvals["test-id-1"] = hitlApprovalRequest{
		ExecutionID: "test-id-1",
		ToolName:    "test.tool",
		RequireMFA:  true,
	}
	globalHITLState.mu.Unlock()

	bp, err := bus.NewProvider(nil)
	require.NoError(t, err)

	app := &Application{
		busProvider: bp,
	}

	mux := http.NewServeMux()
	app.mountHITL(mux)

	t.Run("Valid GET request", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/hitl/approvals", nil)
		rr := httptest.NewRecorder()

		mux.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))

		var response []map[string]interface{}
		err := json.NewDecoder(rr.Body).Decode(&response)
		require.NoError(t, err)
		assert.Len(t, response, 1)
		assert.Equal(t, "test-id-1", response[0]["id"])
		assert.Equal(t, "test.tool", response[0]["tool"])
		assert.Equal(t, "Pending verification for sensitive tool", response[0]["intent"])
		assert.Equal(t, "pending", response[0]["status"])
		assert.Equal(t, true, response[0]["requireMfa"])
	})

	t.Run("Invalid POST request", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/hitl/approvals", nil)
		rr := httptest.NewRecorder()

		mux.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusMethodNotAllowed, rr.Code)
	})
}

func TestMountHITL_PostApproval(t *testing.T) {
	bp, err := bus.NewProvider(nil)
	require.NoError(t, err)

	app := &Application{
		busProvider: bp,
	}

	mux := http.NewServeMux()
	app.mountHITL(mux)

	t.Run("Approve", func(t *testing.T) {
		// Set up global state
		id := "test-approve-id"
		globalHITLState.mu.Lock()
		globalHITLState.approvals[id] = hitlApprovalRequest{
			ExecutionID: id,
			ToolName:    "test.tool",
			RequireMFA:  false,
		}
		globalHITLState.mu.Unlock()

		// Subscribe to responses
		resBus, err := bus.GetBus[hitlApprovalResponse](bp, "hitl.responses."+id)
		require.NoError(t, err)

		responseChan := make(chan hitlApprovalResponse, 1)
		resBus.Subscribe(context.Background(), "test-subscriber", func(msg hitlApprovalResponse) {
			responseChan <- msg
		})

		body := map[string]string{"action": "approved"}
		bodyBytes, _ := json.Marshal(body)
		req := httptest.NewRequest(http.MethodPost, "/hitl/approvals/"+id, bytes.NewReader(bodyBytes))
		rr := httptest.NewRecorder()

		mux.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		// Verify state is cleared
		globalHITLState.mu.RLock()
		_, exists := globalHITLState.approvals[id]
		globalHITLState.mu.RUnlock()
		assert.False(t, exists)

		// Verify bus message
		select {
		case msg := <-responseChan:
			assert.Equal(t, id, msg.ExecutionID)
			assert.True(t, msg.Approved)
		case <-time.After(time.Second):
			t.Fatal("timeout waiting for bus message")
		}
	})

	t.Run("Reject", func(t *testing.T) {
		// Set up global state
		id := "test-reject-id"
		globalHITLState.mu.Lock()
		globalHITLState.approvals[id] = hitlApprovalRequest{
			ExecutionID: id,
			ToolName:    "test.tool",
			RequireMFA:  false,
		}
		globalHITLState.mu.Unlock()

		// Subscribe to responses
		resBus, err := bus.GetBus[hitlApprovalResponse](bp, "hitl.responses."+id)
		require.NoError(t, err)

		responseChan := make(chan hitlApprovalResponse, 1)
		resBus.Subscribe(context.Background(), "test-subscriber", func(msg hitlApprovalResponse) {
			responseChan <- msg
		})

		body := map[string]string{"action": "rejected"}
		bodyBytes, _ := json.Marshal(body)
		req := httptest.NewRequest(http.MethodPost, "/hitl/approvals/"+id, bytes.NewReader(bodyBytes))
		rr := httptest.NewRecorder()

		mux.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		// Verify state is cleared
		globalHITLState.mu.RLock()
		_, exists := globalHITLState.approvals[id]
		globalHITLState.mu.RUnlock()
		assert.False(t, exists)

		// Verify bus message
		select {
		case msg := <-responseChan:
			assert.Equal(t, id, msg.ExecutionID)
			assert.False(t, msg.Approved)
		case <-time.After(time.Second):
			t.Fatal("timeout waiting for bus message")
		}
	})

	t.Run("Invalid Method", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/hitl/approvals/123", nil)
		rr := httptest.NewRecorder()
		mux.ServeHTTP(rr, req)
		assert.Equal(t, http.StatusMethodNotAllowed, rr.Code)
	})

	t.Run("Invalid Body", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/hitl/approvals/123", bytes.NewReader([]byte("invalid json")))
		rr := httptest.NewRecorder()
		mux.ServeHTTP(rr, req)
		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})
}

func TestMountHITL_BusSubscription(t *testing.T) {
	// Reset global state
	globalHITLState.mu.Lock()
	globalHITLState.approvals = make(map[string]hitlApprovalRequest)
	globalHITLState.mu.Unlock()

	bp, err := bus.NewProvider(nil)
	require.NoError(t, err)

	app := &Application{
		busProvider: bp,
	}

	mux := http.NewServeMux()
	app.mountHITL(mux)

	// Publish a request
	reqBus, err := bus.GetBus[hitlApprovalRequest](bp, "hitl.requests")
	require.NoError(t, err)

	id := "bus-test-id"
	err = reqBus.Publish(context.Background(), "test-publisher", hitlApprovalRequest{
		ExecutionID: id,
		ToolName:    "bus.tool",
		RequireMFA:  true,
	})
	require.NoError(t, err)

	// Give it a moment to process the message asynchronously
	require.Eventually(t, func() bool {
		globalHITLState.mu.RLock()
		defer globalHITLState.mu.RUnlock()
		_, exists := globalHITLState.approvals[id]
		return exists
	}, time.Second, 10*time.Millisecond)

	globalHITLState.mu.RLock()
	req := globalHITLState.approvals[id]
	globalHITLState.mu.RUnlock()

	assert.Equal(t, id, req.ExecutionID)
	assert.Equal(t, "bus.tool", req.ToolName)
	assert.True(t, req.RequireMFA)
}

func TestMountHITL_NilBusProvider(t *testing.T) {
	app := &Application{
		busProvider: nil,
	}
	mux := http.NewServeMux()
	// Should not panic
	app.mountHITL(mux)
}
