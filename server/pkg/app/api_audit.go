// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/mcpany/core/server/pkg/audit"
)

func (a *Application) handleListAuditLogs(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

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

	if limit := r.URL.Query().Get("limit"); limit != "" {
		if l, err := strconv.Atoi(limit); err == nil {
			filter.Limit = l
		}
	}
	if offset := r.URL.Query().Get("offset"); offset != "" {
		if o, err := strconv.Atoi(offset); err == nil {
			filter.Offset = o
		}
	}

	if a.standardMiddlewares.Audit == nil {
		http.Error(w, "Audit store not configured", http.StatusServiceUnavailable)
		return
	}

	entries, err := a.standardMiddlewares.Audit.Read(r.Context(), filter)
	if err != nil {
		http.Error(w, "Failed to read audit logs: "+err.Error(), http.StatusInternalServerError)
		return
	}

	type jsonAuditEntry struct {
		Timestamp  string `json:"timestamp"`
		ToolName   string `json:"toolName"`
		UserID     string `json:"userId"`
		ProfileID  string `json:"profileId"`
		Arguments  string `json:"arguments"`
		Result     string `json:"result"`
		Error      string `json:"error"`
		Duration   string `json:"duration"`
		DurationMs int64  `json:"durationMs"`
	}

	var jsonEntries []jsonAuditEntry
	for _, e := range entries {
		var resultStr string
		if s, ok := e.Result.(string); ok {
			resultStr = s
		} else {
			b, _ := json.Marshal(e.Result)
			resultStr = string(b)
		}

		jsonEntries = append(jsonEntries, jsonAuditEntry{
			Timestamp:  e.Timestamp.Format(time.RFC3339Nano),
			ToolName:   e.ToolName,
			UserID:     e.UserID,
			ProfileID:  e.ProfileID,
			Arguments:  string(e.Arguments),
			Result:     resultStr,
			Error:      e.Error,
			Duration:   fmt.Sprintf("%dms", e.DurationMs),
			DurationMs: e.DurationMs,
		})
	}

	if jsonEntries == nil {
		jsonEntries = []jsonAuditEntry{}
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"entries": jsonEntries,
	})
}

func (a *Application) handleAuditExport(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Basic filtering from query params
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

	// Get the audit store from standard middlewares
	// Note: We need to ensure standardMiddlewares is accessible.
	// Based on server.go discovery, it is a field on Application.
	if a.standardMiddlewares.Audit == nil {
		http.Error(w, "Audit store not configured", http.StatusServiceUnavailable)
		return
	}

	entries, err := a.standardMiddlewares.Audit.Read(r.Context(), filter)
	if err != nil {
		http.Error(w, "Failed to read audit logs: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=audit_export_%s.csv", time.Now().Format("20060102_150405")))

	writer := csv.NewWriter(w)
	defer writer.Flush()

	// Header
	_ = writer.Write([]string{"Timestamp", "ToolName", "UserID", "ProfileID", "Arguments", "Result", "Error", "DurationMs"})

	for _, entry := range entries {
		_ = writer.Write([]string{
			entry.Timestamp.Format(time.RFC3339Nano),
			entry.ToolName,
			entry.UserID,
			entry.ProfileID,
			string(entry.Arguments),
			fmt.Sprintf("%v", entry.Result),
			entry.Error,
			fmt.Sprintf("%d", entry.DurationMs),
		})
	}
}
