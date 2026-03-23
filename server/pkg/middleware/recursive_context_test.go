// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestRecursiveContextManager(t *testing.T) {
	m := NewRecursiveContextManager()
	data := map[string]any{"user": "jules"}
	session := m.CreateSession(data, 1*time.Minute)

	if session.ID == "" {
		t.Fatal("expected session ID")
	}

	retrieved, exists := m.GetSession(session.ID)
	if !exists || retrieved.Data["user"] != "jules" {
		t.Fatal("failed to retrieve session")
	}

	handler := m.APIHandler()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/context/session/"+session.ID, nil)
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}
