package app

import (
	"encoding/csv"
	"fmt"
	"net/http"
	"time"

	"github.com/mcpany/core/server/pkg/audit"
)

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
