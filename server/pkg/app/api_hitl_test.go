package app

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHITLApprovalsList(t *testing.T) {
	app, _ := setupApiTestApp()

	// Reinitialize globalHITLState to have a clean slate
	globalHITLState = newHITLState()
	globalHITLState.approvals["test-id-1"] = hitlApprovalRequest{
		ExecutionID: "test-id-1",
		ToolName:    "test.tool",
		RequireMFA:  true,
	}

	mux := http.NewServeMux()
	app.mountHITL(mux)

	req := httptest.NewRequest(http.MethodGet, "/hitl/approvals", nil)
	w := httptest.NewRecorder()

	mux.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

	type uiApproval struct {
		ID         string `json:"id"`
		Tool       string `json:"tool"`
		Intent     string `json:"intent"`
		Status     string `json:"status"`
		RequireMfa bool   `json:"requireMfa"`
	}

	var res []uiApproval
	err := json.NewDecoder(w.Body).Decode(&res)
	require.NoError(t, err)

	assert.Len(t, res, 1)
	assert.Equal(t, "test-id-1", res[0].ID)
	assert.Equal(t, "test.tool", res[0].Tool)
	assert.Equal(t, "pending", res[0].Status)
	assert.True(t, res[0].RequireMfa)
}

func TestHITLApprovalsList_MethodNotAllowed(t *testing.T) {
	app, _ := setupApiTestApp()
	mux := http.NewServeMux()
	app.mountHITL(mux)

	req := httptest.NewRequest(http.MethodPost, "/hitl/approvals", nil)
	w := httptest.NewRecorder()

	mux.ServeHTTP(w, req)

	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
}

func TestHITLApprovalProcess_Approve(t *testing.T) {
	app, _ := setupApiTestApp()

	// Reinitialize globalHITLState to have a clean slate
	globalHITLState = newHITLState()
	globalHITLState.approvals["test-id-2"] = hitlApprovalRequest{
		ExecutionID: "test-id-2",
		ToolName:    "test.tool2",
		RequireMFA:  false,
	}

	mux := http.NewServeMux()
	app.mountHITL(mux)

	body := map[string]interface{}{
		"action":  "approved",
		"mfaCode": "123456",
	}
	bodyBytes, err := json.Marshal(body)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/hitl/approvals/test-id-2", bytes.NewReader(bodyBytes))
	w := httptest.NewRecorder()

	mux.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Check if removed from pending
	globalHITLState.mu.RLock()
	_, exists := globalHITLState.approvals["test-id-2"]
	globalHITLState.mu.RUnlock()

	assert.False(t, exists, "Approval should be removed from pending state after processing")
}

func TestHITLApprovalProcess_Reject(t *testing.T) {
	app, _ := setupApiTestApp()

	globalHITLState = newHITLState()
	globalHITLState.approvals["test-id-3"] = hitlApprovalRequest{
		ExecutionID: "test-id-3",
		ToolName:    "test.tool3",
		RequireMFA:  false,
	}

	mux := http.NewServeMux()
	app.mountHITL(mux)

	body := map[string]interface{}{
		"action":  "rejected",
	}
	bodyBytes, err := json.Marshal(body)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/hitl/approvals/test-id-3", bytes.NewReader(bodyBytes))
	w := httptest.NewRecorder()

	mux.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Check if removed from pending
	globalHITLState.mu.RLock()
	_, exists := globalHITLState.approvals["test-id-3"]
	globalHITLState.mu.RUnlock()

	assert.False(t, exists, "Approval should be removed from pending state after processing")
}

func TestHITLApprovalProcess_MethodNotAllowed(t *testing.T) {
	app, _ := setupApiTestApp()
	mux := http.NewServeMux()
	app.mountHITL(mux)

	req := httptest.NewRequest(http.MethodGet, "/hitl/approvals/test-id-4", nil)
	w := httptest.NewRecorder()

	mux.ServeHTTP(w, req)

	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
}

func TestHITLApprovalProcess_BadRequest(t *testing.T) {
	app, _ := setupApiTestApp()
	mux := http.NewServeMux()
	app.mountHITL(mux)

	// Invalid JSON body
	req := httptest.NewRequest(http.MethodPost, "/hitl/approvals/test-id-5", bytes.NewReader([]byte("{invalid json}")))
	w := httptest.NewRecorder()

	mux.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}
