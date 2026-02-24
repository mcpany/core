// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"github.com/mcpany/core/server/pkg/logging"
	"github.com/mcpany/core/server/pkg/storage"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(_ *http.Request) bool {
		return true // Allow all origins for now, should be configurable
	},
}

// handleListLogs retrieves historical logs from storage.
func (a *Application) handleListLogs(store storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		limit := 1000
		offset := 0

		if l := r.URL.Query().Get("limit"); l != "" {
			fmt.Sscanf(l, "%d", &limit)
		}
		if o := r.URL.Query().Get("offset"); o != "" {
			fmt.Sscanf(o, "%d", &offset)
		}

		// Cap limit
		if limit > 5000 {
			limit = 5000
		}

		logs, err := store.ListLogs(r.Context(), limit, offset)
		if err != nil {
			logging.GetLogger().Error("failed to list logs", "error", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		// Reverse logs to be chronological?
		// ListLogs returns DESC (newest first).
		// UI expects to prepend? Or append?
		// If UI does `setLogs(initialLogs)` and then WS appends.
		// If initialLogs is [Newest, ..., Oldest].
		// And then WS appends [Even Newer].
		// The list becomes [Newest...Oldest, Even Newer].
		// This is weird.
		// Usually logs are [Oldest, ..., Newest].
		// So we should reverse the list here.

		// Optimization: Reverse in place
		for i, j := 0, len(logs)-1; i < j; i, j = i+1, j-1 {
			logs[i], logs[j] = logs[j], logs[i]
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(logs); err != nil {
			logging.GetLogger().Error("failed to encode logs", "error", err)
		}
	}
}

// handleLogsWS handles WebSocket connections for log streaming.
func (a *Application) handleLogsWS() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			logging.GetLogger().Error("failed to upgrade to websocket", "error", err)
			return
		}
		defer func() {
			if err := conn.Close(); err != nil {
				logging.GetLogger().Error("failed to close websocket connection", "error", err)
			}
		}()

		// Subscribe to logs with history
		// Note: We might want to DISABLE history from Broadcaster if we rely on REST API for initial load?
		// But REST API + WS race condition might cause missing logs in between.
		// Correct approach:
		// 1. Fetch REST API (up to T1)
		// 2. Connect WS.
		// WS sends history (last 1000).
		// Frontend merges and dedupes.
		// Since LogEntry has ID, dedup works.
		// So keeping SubscribeWithHistory is fine and robust.

		logCh, history := logging.GlobalBroadcaster.SubscribeWithHistory()
		defer logging.GlobalBroadcaster.Unsubscribe(logCh)

		// Set write deadline
		if err := conn.SetWriteDeadline(time.Now().Add(10 * time.Second)); err != nil {
			logging.GetLogger().Error("failed to set write deadline", "error", err)
			return
		}
		conn.SetPongHandler(func(string) error {
			return conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
		})

		// Send history
		for _, msg := range history {
			if err := conn.SetWriteDeadline(time.Now().Add(10 * time.Second)); err != nil {
				logging.GetLogger().Error("failed to set write deadline", "error", err)
				return
			}
			// ⚡ BOLT: Write struct directly to WebSocket (marshals internally)
			if err := conn.WriteJSON(msg); err != nil {
				logging.GetLogger().Error("failed to write history log message to websocket", "error", err)
				return
			}
		}

		// Send ping periodically
		go func() {
			ticker := time.NewTicker(5 * time.Second)
			defer ticker.Stop()
			for range ticker.C {
				if err := conn.WriteControl(websocket.PingMessage, []byte{}, time.Now().Add(time.Second)); err != nil {
					return
				}
			}
		}()

		for msg := range logCh {
			// ⚡ BOLT: Write struct directly to WebSocket (marshals internally)
			if err := conn.WriteJSON(msg); err != nil {
				logging.GetLogger().Error("failed to write log message to websocket", "error", err)
				return
			}
			if err := conn.SetWriteDeadline(time.Now().Add(10 * time.Second)); err != nil {
				logging.GetLogger().Error("failed to set write deadline", "error", err)
				return
			}
		}
	}
}
