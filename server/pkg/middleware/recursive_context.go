// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/mcpany/core/server/pkg/logging"
)

// SessionState represents the shared state for a recursive context session.
type SessionState struct {
	ID        string         `json:"id"`
	Data      map[string]any `json:"data"`
	CreatedAt time.Time      `json:"created_at"`
	ExpiresAt time.Time      `json:"expires_at"`
}

// RecursiveContextManager manages the shared context sessions.
type RecursiveContextManager struct {
	mu       sync.RWMutex
	sessions map[string]*SessionState
}

// NewRecursiveContextManager initializes a new manager with background cleanup.
func NewRecursiveContextManager() *RecursiveContextManager {
	m := &RecursiveContextManager{
		sessions: make(map[string]*SessionState),
	}
	go m.cleanupLoop()
	return m
}

func (m *RecursiveContextManager) cleanupLoop() {
	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		m.cleanup()
	}
}

func (m *RecursiveContextManager) cleanup() {
	m.mu.Lock()
	defer m.mu.Unlock()
	now := time.Now()
	for id, s := range m.sessions {
		if now.After(s.ExpiresAt) {
			delete(m.sessions, id)
		}
	}
}

// CreateSession generates a new session and performs a quick opportunistic cleanup.
func (m *RecursiveContextManager) CreateSession(data map[string]any, ttl time.Duration) *SessionState {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Opportunistic cleanup of expired sessions to prevent runaway growth
	now := time.Now()
	if len(m.sessions) > 1000 {
		for id, s := range m.sessions {
			if now.After(s.ExpiresAt) {
				delete(m.sessions, id)
			}
		}
	}

	id := uuid.New().String()
	s := &SessionState{ID: id, Data: data, CreatedAt: now, ExpiresAt: now.Add(ttl)}
	m.sessions[id] = s
	return s
}

// GetSession retrieves an active session.
func (m *RecursiveContextManager) GetSession(id string) (*SessionState, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	s, e := m.sessions[id]
	if !e || time.Now().After(s.ExpiresAt) {
		return nil, false
	}
	return s, true
}

// APIHandler returns an HTTP handler for context management.
func (m *RecursiveContextManager) APIHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			var req struct {
				Data map[string]any `json:"data"`
				TTL  int            `json:"ttl_seconds"`
			}
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				http.Error(w, "Invalid JSON", 400)
				return
			}
			ttl := time.Duration(req.TTL) * time.Second
			if ttl == 0 {
				ttl = time.Hour
			}
			s := m.CreateSession(req.Data, ttl)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(s)
			return
		}
		if r.Method == http.MethodGet {
			id := r.URL.Query().Get("id")
			if id == "" {
				path := r.URL.Path
				if idx := strings.LastIndex(path, "/"); idx != -1 {
					id = path[idx+1:]
				}
			}
			if id == "" || id == "session" {
				http.Error(w, "Session ID required", 400)
				return
			}
			s, exists := m.GetSession(id)
			if !exists {
				http.Error(w, "Not found", 404)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(s)
			return
		}
		http.Error(w, "Method not allowed", 405)
	}
}

type RecursiveContextKeyType string

const RecursiveContextDataKey RecursiveContextKeyType = "recursive_context_data"

// HandleContext intercepts requests to inject context with observability.
func (m *RecursiveContextManager) HandleContext(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Header.Get("X-MCP-Parent-Context-ID")
		if id != "" {
			if s, exists := m.GetSession(id); exists {
				logging.GetLogger().Debug("Injecting recursive context", "session_id", id)
				ctx := context.WithValue(r.Context(), RecursiveContextDataKey, s.Data)
				r = r.WithContext(ctx)
			} else {
				logging.GetLogger().Warn("Recursive context session not found or expired", "session_id", id)
			}
		}
		next.ServeHTTP(w, r)
	})
}
