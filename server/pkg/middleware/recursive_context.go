// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"context"
	"encoding/json"
	"net/http"
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

// NewRecursiveContextManager initializes and returns a new RecursiveContextManager.
//
// Parameters:
//   - None.
//
// Returns:
//   - *RecursiveContextManager: A pointer to the newly created manager instance.
//
// Errors:
//   - None.
//
// Side Effects:
//   - Allocates memory for the manager and its internal session map.
//
// Summary: Initializes a new RecursiveContextManager.
func NewRecursiveContextManager() *RecursiveContextManager {
	return &RecursiveContextManager{
		sessions: make(map[string]*SessionState),
	}
}

// CreateSession generates a new recursive context session with the provided data.
//
// Parameters:
//   - data (map[string]any): The initial state data to be stored.
//   - ttl (time.Duration): The duration before the session expires.
//
// Returns:
//   - *SessionState: A pointer to the newly created session state.
//
// Errors:
//   - None.
//
// Side Effects:
//   - Modifies the internal sessions map.
//
// Summary: Creates a new context session.
func (m *RecursiveContextManager) CreateSession(data map[string]any, ttl time.Duration) *SessionState {
	m.mu.Lock()
	defer m.mu.Unlock()

	id := uuid.New().String()
	now := time.Now()
	session := &SessionState{
		ID:        id,
		Data:      data,
		CreatedAt: now,
		ExpiresAt: now.Add(ttl),
	}
	m.sessions[id] = session

	// Simple cleanup of expired sessions
	for k, v := range m.sessions {
		if now.After(v.ExpiresAt) {
			delete(m.sessions, k)
		}
	}

	return session
}

// GetSession retrieves an active context session by its unique identifier.
//
// Parameters:
//   - id (string): The unique UUID string of the session to retrieve.
//
// Returns:
//   - *SessionState: A pointer to the requested session state.
//   - bool: True if the session was found and is active.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
//
// Summary: Retrieves an active context session by ID.
func (m *RecursiveContextManager) GetSession(id string) (*SessionState, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	session, exists := m.sessions[id]
	if !exists {
		return nil, false
	}
	if time.Now().After(session.ExpiresAt) {
		return nil, false
	}
	return session, true
}

// APIHandler constructs an HTTP handler for managing context sessions.
//
// Parameters:
//   - None.
//
// Returns:
//   - http.HandlerFunc: A handler function that processes context requests.
//
// Errors:
//   - Returns HTTP 400 for invalid requests.
//   - Returns HTTP 404 for missing sessions.
//
// Side Effects:
//   - Modifies the HTTP response writer.
//
// Summary: Returns an HTTP handler for the context session API.
func (m *RecursiveContextManager) APIHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			var req struct {
				Data map[string]any `json:"data"`
				TTL  int            `json:"ttl_seconds"`
			}
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				http.Error(w, "Invalid JSON body", http.StatusBadRequest)
				return
			}
			ttl := time.Duration(req.TTL) * time.Second
			if ttl == 0 {
				ttl = 1 * time.Hour
			}

			session := m.CreateSession(req.Data, ttl)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(session)
			return
		}

		if r.Method == http.MethodGet {
			id := r.URL.Query().Get("id")
			if id == "" {
				path := r.URL.Path
				if len(path) > 17 && path[:17] == "/context/session/" {
					id = path[17:]
				}
			}

			if id == "" {
				http.Error(w, "Session ID required", http.StatusBadRequest)
				return
			}

			session, exists := m.GetSession(id)
			if !exists {
				http.Error(w, "Session not found", http.StatusNotFound)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(session)
			return
		}
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// RecursiveContextKeyType is a custom type for context keys.
type RecursiveContextKeyType string

const (
	// RecursiveContextDataKey is the key for context data storage.
	RecursiveContextDataKey RecursiveContextKeyType = "recursive_context_data"
)

// HandleContext intercepts HTTP requests to inject recursive context state.
//
// Parameters:
//   - next (http.Handler): The next HTTP handler in the chain.
//
// Returns:
//   - http.Handler: A new HTTP handler with context injection logic.
//
// Errors:
//   - None.
//
// Side Effects:
//   - Modifies the request context.
//
// Summary: Middleware for recursive context injection.
func (m *RecursiveContextManager) HandleContext(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Header.Get("X-MCP-Parent-Context-ID")
		if id != "" {
			if session, exists := m.GetSession(id); exists {
				ctx := context.WithValue(r.Context(),
					RecursiveContextDataKey, session.Data)
				r = r.WithContext(ctx)
			}
		}
		next.ServeHTTP(w, r)
	})
}
