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
	"github.com/mcpany/core/server/pkg/logging"
)

// AuditLogEntry represents a single audit log event for JSON response.
// We use camelCase to match the UI expectations.
type AuditLogEntry struct {
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

func (a *Application) handleAuditLogs(w http.ResponseWriter, r *http.Request) {
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

	// Get the audit store from standard middlewares
	if a.standardMiddlewares == nil || a.standardMiddlewares.Audit == nil {
		http.Error(w, "Audit store not configured", http.StatusServiceUnavailable)
		return
	}

	entries, err := a.standardMiddlewares.Audit.Read(r.Context(), filter)
	if err != nil {
		logging.GetLogger().Error("failed to read audit logs", "error", err)
		http.Error(w, "Failed to read audit logs", http.StatusInternalServerError)
		return
	}

	responseEntries := make([]AuditLogEntry, 0, len(entries))
	for _, entry := range entries {
		// Convert Result to string if possible
		var resultStr string
		if entry.Result != nil {
			if s, ok := entry.Result.(string); ok {
				resultStr = s
			} else {
				b, _ := json.Marshal(entry.Result)
				resultStr = string(b)
			}
		}

		responseEntries = append(responseEntries, AuditLogEntry{
			Timestamp:  entry.Timestamp.Format(time.RFC3339Nano),
			ToolName:   entry.ToolName,
			UserID:     entry.UserID,
			ProfileID:  entry.ProfileID,
			Arguments:  string(entry.Arguments),
			Result:     resultStr,
			Error:      entry.Error,
			Duration:   entry.Duration,
			DurationMs: entry.DurationMs,
		})
	}

	// Wrap in object to match ListAuditLogsResponse
	response := map[string]any{
		"entries": responseEntries,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		logging.GetLogger().Error("failed to encode audit logs response", "error", err)
	}
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
	if a.standardMiddlewares == nil || a.standardMiddlewares.Audit == nil {
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
