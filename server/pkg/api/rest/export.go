// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package rest

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/mcpany/core/server/pkg/audit"
	"github.com/mcpany/core/server/pkg/logging"
	"github.com/mcpany/core/server/pkg/middleware"
)

// ExportAuditLogsHandler returns a handler that exports audit logs as CSV.
func ExportAuditLogsHandler(auditMiddleware *middleware.AuditMiddleware) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if auditMiddleware == nil {
			http.Error(w, "Audit logging is not enabled", http.StatusNotImplemented)
			return
		}

		// Parse query parameters
		filter := audit.Filter{}
		query := r.URL.Query()

		if startStr := query.Get("start_time"); startStr != "" {
			if t, err := time.Parse(time.RFC3339, startStr); err == nil {
				filter.StartTime = &t
			}
		}

		if endStr := query.Get("end_time"); endStr != "" {
			if t, err := time.Parse(time.RFC3339, endStr); err == nil {
				filter.EndTime = &t
			}
		}

		filter.ToolName = query.Get("tool_name")
		filter.UserID = query.Get("user_id")
		filter.ProfileID = query.Get("profile_id")

		if limitStr := query.Get("limit"); limitStr != "" {
			if l, err := strconv.Atoi(limitStr); err == nil {
				filter.Limit = l
			}
		}

		if offsetStr := query.Get("offset"); offsetStr != "" {
			if o, err := strconv.Atoi(offsetStr); err == nil {
				filter.Offset = o
			}
		}

		// Enforce hard limit to prevent OOM
		const maxLimit = 10000
		if filter.Limit <= 0 || filter.Limit > maxLimit {
			filter.Limit = maxLimit
		}

		// Read logs
		logs, err := auditMiddleware.Read(r.Context(), filter)
		if err != nil {
			logging.GetLogger().Error("Failed to read audit logs", "error", err)
			http.Error(w, "Failed to read audit logs", http.StatusInternalServerError)
			return
		}

		// Set headers
		w.Header().Set("Content-Type", "text/csv")
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"audit_logs_%d.csv\"", time.Now().Unix()))

		// Write CSV
		writer := csv.NewWriter(w)
		defer writer.Flush()

		// Header
		header := []string{"Timestamp", "Tool Name", "User ID", "Profile ID", "Duration", "Error", "Arguments", "Result"}
		if err := writer.Write(header); err != nil {
			logging.GetLogger().Error("Failed to write CSV header", "error", err)
			return
		}

		// Rows
		for _, log := range logs {
			argsJSON, _ := json.Marshal(log.Arguments)
			resultJSON, _ := json.Marshal(log.Result)

			row := []string{
				log.Timestamp.Format(time.RFC3339),
				log.ToolName,
				log.UserID,
				log.ProfileID,
				log.Duration,
				log.Error,
				string(argsJSON),
				string(resultJSON),
			}
			if err := writer.Write(row); err != nil {
				logging.GetLogger().Error("Failed to write CSV row", "error", err)
				return
			}
		}
	}
}
