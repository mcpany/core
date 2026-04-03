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

// SessionState represents the public SessionState entity.
//
// Summary: Defines the structured data model representing a state.
//
// Parameters:
//   - None.
//
// Returns:
//   - None.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
type SessionState struct {
	ID        string                 `json:"id"`
	Data      map[string]interface{} `json:"data"`
	CreatedAt time.Time              `json:"created_at"`
	ExpiresAt time.Time              `json:"expires_at"`
}

// RecursiveContextManager represents the public RecursiveContextManager entity.
//
// Summary: Coordinates operations and orchestrates lifecycle events for the context manager components.
//
// Parameters:
//   - None.
//
// Returns:
//   - None.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
type RecursiveContextManager struct {
	mu       sync.RWMutex
	sessions map[string]*SessionState
}

// NewRecursiveContextManager serves as a public interface for interacting with NewRecursiveContextManager.
//
// Summary: Constructs and returns an initialized recursive context manager ready for consumption.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the successfully computed domain model or execution state.
//
// Errors:
//   - No explicit errors are thrown by this operation.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
func NewRecursiveContextManager() *RecursiveContextManager {
	return &RecursiveContextManager{
		sessions: make(map[string]*SessionState),
	}
}

// CreateSession serves as a public interface for interacting with CreateSession.
//
// Summary: Create the session appropriately based on current system conditions.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the successfully computed domain model or execution state.
//
// Errors:
//   - No explicit errors are thrown by this operation.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
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

// GetSession serves as a public interface for interacting with GetSession.
//
// Summary: Fetches and returns the underlying session from the system state.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the successfully computed domain model or execution state.
//
// Errors:
//   - No explicit errors are thrown by this operation.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
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

// APIHandler serves as a public interface for interacting with APIHandler.
//
// Summary: Api the handler appropriately based on current system conditions.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the successfully computed domain model or execution state.
//
// Errors:
//   - No explicit errors are thrown by this operation.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
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

// RecursiveContextKeyType represents the public RecursiveContextKeyType entity.
//
// Summary: Defines the structured data model representing a context key type.
//
// Parameters:
//   - None.
//
// Returns:
//   - None.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
type RecursiveContextKeyType string

const (
	// RecursiveContextDataKey is the key used to store the recursive context data in the request context.
	// Summary: Defines RecursiveContextDataKe.
	RecursiveContextDataKey RecursiveContextKeyType = "recursive_context_data"
)

// HandleContext serves as a public interface for interacting with HandleContext.
//
// Summary: Handle the context appropriately based on current system conditions.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the successfully computed domain model or execution state.
//
// Errors:
//   - No explicit errors are thrown by this operation.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
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
