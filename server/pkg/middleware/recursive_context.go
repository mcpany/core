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
)

// SessionState represents the shared state for a recursive context session.
//
// Summary: Stores data and expiration metadata for a context session.
type SessionState struct {
	ID        string         `json:"id"`
	Data      map[string]any `json:"data"`
	CreatedAt time.Time      `json:"created_at"`
	ExpiresAt time.Time      `json:"expires_at"`
}

// RecursiveContextManager manages the shared context sessions.
//
// Summary: Provides thread-safe storage and retrieval of context sessions.
type RecursiveContextManager struct {
	mu       sync.RWMutex
	sessions map[string]*SessionState
}

// NewRecursiveContextManager initializes a new manager.
//
// Returns:
//   - *RecursiveContextManager: New instance.
func NewRecursiveContextManager() *RecursiveContextManager {
	return &RecursiveContextManager{
		sessions: make(map[string]*SessionState),
	}
}

// CreateSession generates a new session.
//
// Parameters:
//   - data (map[string]any): Initial state.
//   - ttl (time.Duration): Session expiration.
//
// Returns:
//   - *SessionState: New session.
func (m *RecursiveContextManager) CreateSession(data map[string]any, ttl time.Duration) *SessionState {
	m.mu.Lock()
	defer m.mu.Unlock()
	id := uuid.New().String()
	now := time.Now()
	s := &SessionState{ID: id, Data: data, CreatedAt: now, ExpiresAt: now.Add(ttl)}
	m.sessions[id] = s
	return s
}

// GetSession retrieves an active session.
//
// Parameters:
//   - id (string): Session UUID.
//
// Returns:
//   - *SessionState: Session state if found and active.
//   - bool: Success status.
func (m *RecursiveContextManager) GetSession(id string) (*SessionState, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	s, e := m.sessions[id]
	if !e || time.Now().After(s.ExpiresAt) {
		return nil, false
	}
	return s, true
}

// APIHandler constructs an HTTP handler for context management.
//
// Returns:
//   - http.HandlerFunc: Handler for POST/GET session requests.
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

// RecursiveContextKeyType is a custom type for context keys.
type RecursiveContextKeyType string

// RecursiveContextDataKey is the key for context storage.
const RecursiveContextDataKey RecursiveContextKeyType = "recursive_context_data"

// HandleContext intercepts requests to inject context.
//
// Parameters:
//   - next (http.Handler): Chain successor.
//
// Returns:
//   - http.Handler: Wrapped handler.
func (m *RecursiveContextManager) HandleContext(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Header.Get("X-MCP-Parent-Context-ID")
		if id != "" {
			if s, exists := m.GetSession(id); exists {
				r = r.WithContext(context.WithValue(r.Context(), RecursiveContextDataKey, s.Data))
			}
		}
		next.ServeHTTP(w, r)
	})
}
