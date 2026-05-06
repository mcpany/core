// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"encoding/json"
	"net/http"
	"sync"

	"github.com/mcpany/core/server/pkg/logging"
)

// PCTRMiddleware provides Privilege-Constrained Token Rotation.
type PCTRMiddleware struct {
	mu sync.RWMutex
	// Active tokens and their current scopes
	activeTokens map[string][]string
}

// NewPCTRMiddleware creates a new PCTRMiddleware instance.
func NewPCTRMiddleware() *PCTRMiddleware {
	return &PCTRMiddleware{
		activeTokens: make(map[string][]string),
	}
}

// RegisterToken registers a new token with its scopes (for testing/setup).
func (m *PCTRMiddleware) RegisterToken(token string, scopes []string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.activeTokens[token] = scopes
}

type RotateRequest struct {
	OldToken        string   `json:"old_token"`
	RequestedScopes []string `json:"requested_scopes"`
}

type RotateResponse struct {
	NewToken string   `json:"new_token"`
	Scopes   []string `json:"scopes"`
}

// APIHandler returns an http.HandlerFunc for the token rotation endpoint.
func (m *PCTRMiddleware) APIHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req RotateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}

		m.mu.RLock()
		currentScopes, ok := m.activeTokens[req.OldToken]
		m.mu.RUnlock()

		if !ok {
			http.Error(w, "invalid token", http.StatusUnauthorized)
			return
		}

		// Mathematical subset validation
		if !isSubset(req.RequestedScopes, currentScopes) {
			logging.GetLogger().Warn("PCTR: rotation request rejected due to scope escalation attempt", "requested", req.RequestedScopes, "current", currentScopes)
			http.Error(w, "forbidden: requested scopes exceed current authority", http.StatusForbidden)
			return
		}

		// Success: Issue new token (simulated by appending suffix)
		newToken := req.OldToken + "_rotated"

		m.mu.Lock()
		m.activeTokens[newToken] = req.RequestedScopes
		// Atomically invalidate old token
		delete(m.activeTokens, req.OldToken)
		m.mu.Unlock()

		resp := RotateResponse{
			NewToken: newToken,
			Scopes:   req.RequestedScopes,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}
}

// isSubset checks if requested is a strict subset of current.
func isSubset(requested, current []string) bool {
	currentMap := make(map[string]bool)
	for _, s := range current {
		currentMap[s] = true
	}
	for _, r := range requested {
		if !currentMap[r] {
			return false
		}
	}
	return true
}
