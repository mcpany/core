// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"
)

// MultiAgentSessionManager provides handlers for multi-agent session coordination.
// It leverages the existing RecursiveContextManager (Blackboard) to persist state
// across multiple agents.
type MultiAgentSessionManager struct {
	contextManager *RecursiveContextManager
}

// NewMultiAgentSessionManager creates a new MultiAgentSessionManager.
func NewMultiAgentSessionManager(contextManager *RecursiveContextManager) *MultiAgentSessionManager {
	return &MultiAgentSessionManager{
		contextManager: contextManager,
	}
}

// APIHandler constructs an HTTP handler function for managing multi-agent sessions.
func (m *MultiAgentSessionManager) APIHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path

		switch {
		case r.Method == http.MethodPost && path == "/session/init":
			m.handleInitSession(w, r)
		case r.Method == http.MethodPost && strings.HasPrefix(path, "/session/") && strings.HasSuffix(path, "/handoff"):
			m.handleHandoffSession(w, r)
		case r.Method == http.MethodGet && strings.HasPrefix(path, "/session/") && strings.HasSuffix(path, "/state"):
			m.handleGetState(w, r)
		default:
			http.Error(w, "Not found", http.StatusNotFound)
		}
	}
}

func (m *MultiAgentSessionManager) handleInitSession(w http.ResponseWriter, r *http.Request) {
	var req struct {
		InitialState map[string]interface{} `json:"initial_state"`
		TTL          int                    `json:"ttl,omitempty"` // Time to live in seconds
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// default to 1 hour
	if req.TTL == 0 {
		req.TTL = 3600
	}

	importTime := time.Duration(req.TTL) * time.Second
	session := m.contextManager.CreateSession(req.InitialState, importTime)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"session_id": session.ID,
		"state":      session.Data,
	})
}

func (m *MultiAgentSessionManager) handleHandoffSession(w http.ResponseWriter, r *http.Request) {
	// Extract the session ID from path "/session/{id}/handoff"
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 4 {
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}
	sessionID := parts[2]

	var req struct {
		TargetAgent string                 `json:"target_agent"`
		AddedState  map[string]interface{} `json:"added_state"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	session, exists := m.contextManager.GetSession(sessionID)
	if !exists {
		http.Error(w, "Session not found", http.StatusNotFound)
		return
	}

	session.Mu.Lock()
	if req.AddedState != nil {
		for k, v := range req.AddedState {
			session.Data[k] = v
		}
	}
	session.Data["current_agent"] = req.TargetAgent
	session.Mu.Unlock()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"session_id": session.ID,
		"status":     "handoff_successful",
		"state":      session.Data,
	})
}

func (m *MultiAgentSessionManager) handleGetState(w http.ResponseWriter, r *http.Request) {
	// Extract the session ID from path "/session/{id}/state"
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 4 {
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}
	sessionID := parts[2]

	session, exists := m.contextManager.GetSession(sessionID)
	if !exists {
		http.Error(w, "Session not found", http.StatusNotFound)
		return
	}

	session.Mu.RLock()
	stateCopy := make(map[string]interface{})
	for k, v := range session.Data {
		stateCopy[k] = v
	}
	session.Mu.RUnlock()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"session_id": session.ID,
		"state":      stateCopy,
	})
}
