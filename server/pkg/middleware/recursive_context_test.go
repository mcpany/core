// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestRecursiveContextManager_APIHandler(t *testing.T) {
	manager := NewRecursiveContextManager()
	handler := manager.APIHandler()

	// Test POST /context/session
	data := map[string]interface{}{"key": "value"}
	reqBody := map[string]interface{}{
		"data":        data,
		"ttl_seconds": 60,
	}
	bodyBytes, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/context/session", bytes.NewReader(bodyBytes))
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected status 201 Created, got %d", rec.Code)
	}

	var session SessionState
	if err := json.Unmarshal(rec.Body.Bytes(), &session); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if session.ID == "" {
		t.Fatal("expected session ID, got empty string")
	}
	if session.Data["key"] != "value" {
		t.Fatalf("expected data 'value', got %v", session.Data["key"])
	}

	// Test GET /context/session/:id
	req = httptest.NewRequest(http.MethodGet, "/context/session/"+session.ID, nil)
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200 OK, got %d", rec.Code)
	}

	var retrievedSession SessionState
	if err := json.Unmarshal(rec.Body.Bytes(), &retrievedSession); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if retrievedSession.ID != session.ID {
		t.Fatalf("expected session ID %s, got %s", session.ID, retrievedSession.ID)
	}

	// Test GET with missing ID
	req = httptest.NewRequest(http.MethodGet, "/context/session", nil)
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400 Bad Request, got %d", rec.Code)
	}

	// Test invalid method
	req = httptest.NewRequest(http.MethodPut, "/context/session", nil)
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected status 405 Method Not Allowed, got %d", rec.Code)
	}
}

func TestRecursiveContextManager_HandleContext(t *testing.T) {
	manager := NewRecursiveContextManager()
	data := map[string]interface{}{"role": "admin"}
	session := manager.CreateSession(data, 1*time.Hour)

	var capturedData map[string]interface{}
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		val := r.Context().Value(RecursiveContextDataKey)
		if val != nil {
			capturedData = val.(map[string]interface{})
		}
		w.WriteHeader(http.StatusOK)
	})

	handler := manager.HandleContext(nextHandler)

	// Test with valid X-MCP-Parent-Context-ID header
	req := httptest.NewRequest(http.MethodGet, "/mcp/v1/tools/call", nil)
	req.Header.Set("X-MCP-Parent-Context-ID", session.ID)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200 OK, got %d", rec.Code)
	}
	if capturedData == nil {
		t.Fatal("expected context data to be injected, got nil")
	}
	if capturedData["role"] != "admin" {
		t.Fatalf("expected role 'admin', got %v", capturedData["role"])
	}

	// Test with invalid session ID
	capturedData = nil
	req = httptest.NewRequest(http.MethodGet, "/mcp/v1/tools/call", nil)
	req.Header.Set("X-MCP-Parent-Context-ID", "invalid-id")
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200 OK, got %d", rec.Code)
	}
	if capturedData != nil {
		t.Fatal("expected no context data to be injected, got non-nil")
	}

	// Test without header
	capturedData = nil
	req = httptest.NewRequest(http.MethodGet, "/mcp/v1/tools/call", nil)
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200 OK, got %d", rec.Code)
	}
	if capturedData != nil {
		t.Fatal("expected no context data to be injected, got non-nil")
	}
}
