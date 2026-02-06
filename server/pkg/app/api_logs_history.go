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

// handleLogsHistory returns a handler for querying log history.
func (a *Application) handleLogsHistory(store storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		query := r.URL.Query()
		filter := logging.LogFilter{
			Level:  query.Get("level"),
			Source: query.Get("source"),
			Search: query.Get("search"),
		}

		if limitStr := query.Get("limit"); limitStr != "" {
			if limit, err := strconv.Atoi(limitStr); err == nil {
				filter.Limit = limit
			}
		} else {
			filter.Limit = 100 // Default
		}

		if offsetStr := query.Get("offset"); offsetStr != "" {
			if offset, err := strconv.Atoi(offsetStr); err == nil {
				filter.Offset = offset
			}
		}

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

		logs, err := store.ListLogs(r.Context(), filter)
		if err != nil {
			http.Error(w, "failed to query logs: "+err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(logs)
	}
}
