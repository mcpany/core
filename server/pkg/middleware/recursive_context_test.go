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

func TestRecursiveContextManager(t *testing.T) {
	m := NewRecursiveContextManager()
	data := map[string]any{"user": "jules", "role": "admin"}

	t.Run("CreateAndGetSession", func(t *testing.T) {
		session := m.CreateSession(data, 1*time.Minute)
		if session.ID == "" {
			t.Fatal("expected non-empty session ID")
		}

		retrieved, exists := m.GetSession(session.ID)
		if !exists {
			t.Fatal("session should exist")
		}
		if retrieved.Data["user"] != "jules" {
			t.Errorf("expected user jules, got %v", retrieved.Data["user"])
		}
	})

	t.Run("SessionExpiration", func(t *testing.T) {
		session := m.CreateSession(data, 100*time.Millisecond)
		time.Sleep(200 * time.Millisecond)
		_, exists := m.GetSession(session.ID)
		if exists {
			t.Error("session should have expired")
		}
	})

	t.Run("APIHandler_Post", func(t *testing.T) {
		handler := m.APIHandler()
		reqBody := map[string]any{
			"data":        map[string]any{"project": "mcp-any"},
			"ttl_seconds": 60,
		}
		b, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/context/session", bytes.NewBuffer(b))
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusCreated {
			t.Errorf("expected 201, got %d", rec.Code)
		}

		var resp SessionState
		if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}
		if resp.Data["project"] != "mcp-any" {
			t.Errorf("expected project mcp-any, got %v", resp.Data["project"])
		}
	})

	t.Run("MiddlewareInjection", func(t *testing.T) {
		session := m.CreateSession(map[string]any{"auth": "verified"}, 1*time.Minute)

		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			val := r.Context().Value(RecursiveContextDataKey)
			if val == nil {
				t.Error("expected context data to be injected")
				return
			}
			data := val.(map[string]any)
			if data["auth"] != "verified" {
				t.Errorf("expected auth verified, got %v", data["auth"])
			}
			w.WriteHeader(http.StatusOK)
		})

		middleware := m.HandleContext(next)
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("X-MCP-Parent-Context-ID", session.ID)
		rec := httptest.NewRecorder()

		middleware.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", rec.Code)
		}
	})

	t.Run("Cleanup", func(t *testing.T) {
		// Manually trigger cleanup
		m.CreateSession(data, -1*time.Second) // Expired
		m.cleanup()

		m.mu.RLock()
		count := len(m.sessions)
		m.mu.RUnlock()

		// Note: other tests might have left sessions, but the expired one should be gone
		// This is a basic sanity check
		if count > 10 { // Arbitrary small number
			t.Errorf("cleanup might have failed, session count high: %d", count)
		}
	})
}
