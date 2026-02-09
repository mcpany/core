// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/websocket"
	"github.com/mcpany/core/server/pkg/logging"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(_ *http.Request) bool {
		return true // Allow all origins for now, should be configurable
	},
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
		// We use the broadcaster's in-memory history for quick sync
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
			if err := conn.WriteMessage(websocket.TextMessage, msg); err != nil {
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
			if err := conn.WriteMessage(websocket.TextMessage, msg); err != nil {
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

// handleGetLogs handles the HTTP request for historical logs.
func (a *Application) handleGetLogs() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		store := logging.GetStore()
		if store == nil {
			// If no store is configured, return empty list or error?
			// Empty list is safer to avoid breaking UI.
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode([]logging.LogEntry{})
			return
		}

		limit := 100
		if l := r.URL.Query().Get("limit"); l != "" {
			if val, err := strconv.Atoi(l); err == nil {
				limit = val
			}
		}
		// Cap limit to avoid DoS
		if limit > 1000 {
			limit = 1000
		}

		offset := 0
		if o := r.URL.Query().Get("offset"); o != "" {
			if val, err := strconv.Atoi(o); err == nil {
				offset = val
			}
		}

		opts := logging.LogQueryOptions{
			Limit:  limit,
			Offset: offset,
			Level:  r.URL.Query().Get("level"),
			Source: r.URL.Query().Get("source"),
			Search: r.URL.Query().Get("search"),
		}

		logs, err := store.Read(r.Context(), opts)
		if err != nil {
			logging.GetLogger().Error("Failed to query logs", "error", err)
			http.Error(w, "Failed to query logs", http.StatusInternalServerError)
			return
		}

		if logs == nil {
			logs = []logging.LogEntry{}
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(logs); err != nil {
			logging.GetLogger().Error("Failed to encode logs response", "error", err)
		}
	}
}
