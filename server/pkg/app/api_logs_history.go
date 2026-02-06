// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/mcpany/core/server/pkg/logging"
	"github.com/mcpany/core/server/pkg/storage"
)

func (a *Application) handleLogsHistory(store storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		filter := logging.LogFilter{
			Level:  r.URL.Query().Get("level"),
			Source: r.URL.Query().Get("source"),
			Search: r.URL.Query().Get("search"),
			Limit:  1000,
		}

		if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
			if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
				filter.Limit = l
			}
		}
		if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
			if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
				filter.Offset = o
			}
		}

		if startTimeStr := r.URL.Query().Get("start_time"); startTimeStr != "" {
			if t, err := time.Parse(time.RFC3339, startTimeStr); err == nil {
				filter.StartTime = &t
			}
		}
		if endTimeStr := r.URL.Query().Get("end_time"); endTimeStr != "" {
			if t, err := time.Parse(time.RFC3339, endTimeStr); err == nil {
				filter.EndTime = &t
			}
		}

		logs, total, err := store.QueryLogs(r.Context(), filter)
		if err != nil {
			logging.GetLogger().Error("failed to query logs", "error", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		if logs == nil {
			logs = []logging.LogEntry{}
		}

		resp := map[string]any{
			"logs":  logs,
			"total": total,
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}
}
