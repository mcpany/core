// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/websocket"
	"github.com/mcpany/core/server/pkg/logging"
	"github.com/mcpany/core/server/pkg/storage"
)

// LogLister defines the interface for listing logs.
type LogLister interface {
	ListLogs(ctx context.Context, limit, offset int, level, source, search string) ([]*logging.LogEntry, error)
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(_ *http.Request) bool {
		return true // Allow all origins for now, should be configurable
	},
}

// handleListLogs handles GET /api/v1/logs.
func (a *Application) handleListLogs(store storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		lister, ok := store.(LogLister)
		if !ok {
			// Return empty list if not supported, or error?
			// Client expects array.
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte("[]"))
			return
		}

		// Parse query params
		limitStr := r.URL.Query().Get("limit")
		limit := 100
		if limitStr != "" {
			if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 1000 {
				limit = l
			}
		}

		offsetStr := r.URL.Query().Get("offset")
		offset := 0
		if offsetStr != "" {
			if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
				offset = o
			}
		}

		level := r.URL.Query().Get("level")
		source := r.URL.Query().Get("source")
		search := r.URL.Query().Get("search")

		logs, err := lister.ListLogs(r.Context(), limit, offset, level, source, search)
		if err != nil {
			logging.GetLogger().Error("failed to list logs", "error", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if logs == nil {
			_, _ = w.Write([]byte("[]"))
			return
		}
		_ = json.NewEncoder(w).Encode(logs)
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
