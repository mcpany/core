// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/mcpany/core/server/pkg/logging"
)

// handleGetLogs returns a list of logs with optional filtering.
func (a *Application) handleGetLogs() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		if a.LogStore == nil {
			// Fallback to in-memory history if persistence is not available
			history := logging.GlobalBroadcaster.GetHistory()
			// Convert [][]byte to []LogEntry
			var entries []logging.LogEntry
			for _, b := range history {
				if b == nil {
					continue
				}
				var l logging.LogEntry
				if err := json.Unmarshal(b, &l); err == nil {
					entries = append(entries, l)
				}
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(entries)
			return
		}

		limitStr := r.URL.Query().Get("limit")
		limit := 1000
		if limitStr != "" {
			if l, err := strconv.Atoi(limitStr); err == nil {
				limit = l
			}
		}

		offsetStr := r.URL.Query().Get("offset")
		offset := 0
		if offsetStr != "" {
			if o, err := strconv.Atoi(offsetStr); err == nil {
				offset = o
			}
		}

		logs, err := a.LogStore.Query(limit, offset)
		if err != nil {
			logging.GetLogger().Error("Failed to query logs", "error", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(logs)
	}
}
