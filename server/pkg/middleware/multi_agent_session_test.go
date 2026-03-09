// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestMultiAgentSessionManager_APIHandler(t *testing.T) {
	manager := NewRecursiveContextManager()
	multiAgentManager := NewMultiAgentSessionManager(manager)
	handler := multiAgentManager.APIHandler()

	// 1. Test init session
	initReqBody, _ := json.Marshal(map[string]interface{}{
		"initial_state": map[string]interface{}{"task": "write code"},
	})
	req := httptest.NewRequest(http.MethodPost, "/session/init", bytes.NewReader(initReqBody))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Errorf("Expected status 201 Created, got %v", rr.Code)
	}

	var initResp map[string]interface{}
	if err := json.Unmarshal(rr.Body.Bytes(), &initResp); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}
	sessionID, ok := initResp["session_id"].(string)
	if !ok || sessionID == "" {
		t.Fatalf("Failed to retrieve session ID from response")
	}

	// 2. Test get state
	req = httptest.NewRequest(http.MethodGet, "/session/"+sessionID+"/state", nil)
	rr = httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200 OK, got %v", rr.Code)
	}
	var stateResp map[string]interface{}
	if err := json.Unmarshal(rr.Body.Bytes(), &stateResp); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}
	stateMap := stateResp["state"].(map[string]interface{})
	if stateMap["task"] != "write code" {
		t.Errorf("Expected task 'write code', got %v", stateMap["task"])
	}

	// 3. Test handoff
	handoffBody, _ := json.Marshal(map[string]interface{}{
		"target_agent": "reviewer",
		"added_state": map[string]interface{}{"code": "print('hello')"},
	})
	req = httptest.NewRequest(http.MethodPost, "/session/"+sessionID+"/handoff", bytes.NewReader(handoffBody))
	rr = httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200 OK, got %v", rr.Code)
	}
	var handoffResp map[string]interface{}
	if err := json.Unmarshal(rr.Body.Bytes(), &handoffResp); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}
	newState := handoffResp["state"].(map[string]interface{})
	if newState["current_agent"] != "reviewer" || newState["code"] != "print('hello')" {
		t.Errorf("Handoff state incorrect: %v", newState)
	}

	// 4. Test invalid methods and paths
	req = httptest.NewRequest(http.MethodPost, "/session/init", bytes.NewReader([]byte("invalid json")))
	rr = httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400 Bad Request, got %v", rr.Code)
	}

	req = httptest.NewRequest(http.MethodPost, "/session/"+sessionID+"/handoff", bytes.NewReader([]byte("invalid json")))
	rr = httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400 Bad Request, got %v", rr.Code)
	}

	req = httptest.NewRequest(http.MethodGet, "/session/invalid_id/state", nil)
	rr = httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusNotFound {
		t.Errorf("Expected status 404 Not Found, got %v", rr.Code)
	}

	req = httptest.NewRequest(http.MethodPost, "/session/invalid_id/handoff", bytes.NewReader(handoffBody))
	rr = httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusNotFound {
		t.Errorf("Expected status 404 Not Found, got %v", rr.Code)
	}

	req = httptest.NewRequest(http.MethodGet, "/invalid_path", nil)
	rr = httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusNotFound {
		t.Errorf("Expected status 404 Not Found, got %v", rr.Code)
	}

	req = httptest.NewRequest(http.MethodPost, "/session/123", nil)
	rr = httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusNotFound {
		t.Errorf("Expected status 404 Not Found, got %v", rr.Code)
	}

	req = httptest.NewRequest(http.MethodPost, "/session/123/state", nil)
	rr = httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusNotFound {
		t.Errorf("Expected status 404 Not Found, got %v", rr.Code)
	}
}
