// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/mcpany/core/server/pkg/middleware"
)

// A2AMessagingHub represents the A2A messaging hub.
type A2AMessagingHub struct {
	bridge *middleware.A2ABridgeMiddleware
}

// NewA2AMessagingHub creates a new A2AMessagingHub.
func NewA2AMessagingHub(bridge *middleware.A2ABridgeMiddleware) *A2AMessagingHub {
	return &A2AMessagingHub{
		bridge: bridge,
	}
}

// ProposeHandler handles A2A task proposals.
func (h *A2AMessagingHub) ProposeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var proposal map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&proposal); err != nil {
		http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}

	// Basic validation (simulate Proof-of-Intent check)
	if _, ok := proposal["intent"]; !ok {
		http.Error(w, "Missing Proof-of-Intent (intent field)", http.StatusForbidden)
		return
	}

	// In a real implementation, this would route to the subagent
	// via the A2ABridgeMiddleware or a similar component.
	// For now, we simulate success.

	response := map[string]string{
		"status":  "accepted",
		"message": "Task proposal accepted and routed.",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(response)
}

// MailboxHandler handles real-time task delivery (SSE).
func (h *A2AMessagingHub) MailboxHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	agentID := r.URL.Query().Get("agent_id")
	if agentID == "" {
		http.Error(w, "Missing agent_id", http.StatusBadRequest)
		return
	}

	// Set headers for Server-Sent Events
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	ctx := r.Context()

	// Simulate an SSE stream
	for {
		select {
		case <-ctx.Done():
			return
		default:
			// In a real implementation, this would read from the Blackboard or a queue.
			// Here we just keep the connection open.
			fmt.Fprintf(w, ": keepalive\n\n")
			if flusher, ok := w.(http.Flusher); ok {
				flusher.Flush()
			}
			time.Sleep(15 * time.Second)
		}
	}
}

// RegisterRoutes registers the A2A messaging hub routes with an HTTP mux.
func (h *A2AMessagingHub) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/v1/a2a/propose", h.ProposeHandler)
	mux.HandleFunc("/v1/a2a/mailbox", h.MailboxHandler)
}
