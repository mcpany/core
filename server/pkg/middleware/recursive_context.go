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
	"github.com/mcpany/core/server/pkg/logging"
)

// SessionState represents the shared state for a recursive context session.
type SessionState struct {
	ID        string                 `json:"id"`
	Data      map[string]interface{} `json:"data"`
	CreatedAt time.Time              `json:"created_at"`
	ExpiresAt time.Time              `json:"expires_at"`
}

// RecursiveContextManager manages the shared context sessions (Blackboard).
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
func NewRecursiveContextManager() *RecursiveContextManager {
	return &RecursiveContextManager{
		sessions: make(map[string]*SessionState),
	}
}

// CreateSession generates a new recursive context session with the provided data and expiration time.
//
// Parameters:
//   - data (map[string]interface{}): The initial state data to be stored in the session.
//   - ttl (time.Duration): The time-to-live duration for the session before it expires.
//
// Returns:
//   - *SessionState: A pointer to the newly created session state.
//
// Errors:
//   - None.
//
// Side Effects:
//   - Modifies the internal sessions map by adding a new session.
//   - Performs a cleanup of expired sessions during insertion, removing them from the map.
func (m *RecursiveContextManager) CreateSession(data map[string]interface{}, ttl time.Duration) *SessionState {
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
//   - *SessionState: A pointer to the requested session state, or nil if not found or expired.
//   - bool: True if the session was successfully found and is active, false otherwise.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
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

// APIHandler constructs an HTTP handler function for managing Recursive Context Protocol endpoints.
//
// Parameters:
//   - None.
//
// Returns:
//   - http.HandlerFunc: A handler function that processes POST (create session) and GET (retrieve session) requests.
//
// Errors:
//   - None.
//
// Side Effects:
//   - Modifies the HTTP response writer based on the request logic, including sending JSON responses and error codes.
//   - When processing a POST request, creates a new session in the manager.
func (m *RecursiveContextManager) APIHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			var req struct {
				Data map[string]interface{} `json:"data"`
				TTL  int                    `json:"ttl_seconds"`
			}
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				http.Error(w, "Invalid JSON body", http.StatusBadRequest)
				return
			}
			ttl := time.Duration(req.TTL) * time.Second
			if ttl == 0 {
				ttl = 1 * time.Hour // Default TTL
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
				// Try to extract from path if not in query
				pathParts := r.URL.Path
				if len(pathParts) > 17 && pathParts[:17] == "/context/session/" {
					id = pathParts[17:]
				}
			}

			if id == "" {
				http.Error(w, "Session ID required", http.StatusBadRequest)
				return
			}

			session, exists := m.GetSession(id)
			if !exists {
				http.Error(w, "Session not found or expired", http.StatusNotFound)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(session)
			return
		}

		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// contextKey is a custom type for context keys to avoid collisions.


// RecursiveContextKeyType is a custom type for context keys to avoid collisions.
type RecursiveContextKeyType string

const (
	// RecursiveContextDataKey is the key used to store the recursive context data in the request context.
	RecursiveContextDataKey RecursiveContextKeyType = "recursive_context_data"
)

// HandleContext intercepts HTTP requests to inject recursive context state based on the X-MCP-Parent-Context-ID header.
//
// Parameters:
//   - next (http.Handler): The next HTTP handler in the middleware chain.
//
// Returns:
//   - http.Handler: A new HTTP handler that wraps the provided handler with context injection logic.
//
// Errors:
//   - None.
//
// Side Effects:
//   - Reads from the incoming HTTP request headers.
//   - Modifies the request context by injecting session data if a valid context ID is found.
//   - Logs debug or warning messages depending on the presence and validity of the context session.
func (m *RecursiveContextManager) HandleContext(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		contextID := r.Header.Get("X-MCP-Parent-Context-ID")

		if contextID != "" {
			session, exists := m.GetSession(contextID)
			if exists {
				// Inject the session data into the request context
				ctx := context.WithValue(r.Context(), RecursiveContextDataKey, session.Data)
				r = r.WithContext(ctx)
				logging.GetLogger().Debug("Injected recursive context", "context_id", contextID)
			} else {
				logging.GetLogger().Warn("Recursive context session not found or expired", "context_id", contextID)
			}
		}

		next.ServeHTTP(w, r)
	})
}
