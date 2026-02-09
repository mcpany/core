// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/mcpany/core/server/pkg/audit"
)

func (a *Application) handleAuditLogs(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse filtering from query params
	filter := audit.Filter{}
	if start := r.URL.Query().Get("start_time"); start != "" {
		if t, err := time.Parse(time.RFC3339, start); err == nil {
			filter.StartTime = &t
		}
	}
	if end := r.URL.Query().Get("end_time"); end != "" {
		if t, err := time.Parse(time.RFC3339, end); err == nil {
			filter.EndTime = &t
		}
	}
	filter.ToolName = r.URL.Query().Get("tool_name")
	filter.UserID = r.URL.Query().Get("user_id")
	filter.ProfileID = r.URL.Query().Get("profile_id")

	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil {
			filter.Limit = l
		}
	} else {
		filter.Limit = 100 // Default limit
	}

	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil {
			filter.Offset = o
		}
	}

	// Get the audit store from standard middlewares
	if a.standardMiddlewares == nil || a.standardMiddlewares.Audit == nil {
		http.Error(w, "Audit store not configured", http.StatusServiceUnavailable)
		return
	}

	entries, err := a.standardMiddlewares.Audit.Read(r.Context(), filter)
	if err != nil {
		http.Error(w, "Failed to read audit logs: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	// Wrap in { entries: [...] } response object
	resp := map[string]any{
		"entries": entries,
	}
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}
