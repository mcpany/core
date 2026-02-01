// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	loggingv1 "github.com/mcpany/core/proto/logging/v1"
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

		sentIDs := make(map[string]struct{})

		// 1. Send DB Logs (Persistent History)
		if a.Storage != nil {
			ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
			// Fetch up to 2000 logs from storage to provide context beyond memory buffer
			dbLogs, err := a.Storage.ListLogs(ctx, &loggingv1.LogFilter{Limit: 2000})
			cancel()
			if err != nil {
				logging.GetLogger().Error("failed to list logs from storage", "error", err)
			} else {
				for _, log := range dbLogs {
					if _, sent := sentIDs[log.GetId()]; sent {
						continue
					}

					var metadata map[string]any
					if log.GetMetadataJson() != "" {
						_ = json.Unmarshal([]byte(log.GetMetadataJson()), &metadata)
					}
					entry := logging.LogEntry{
						ID:        log.GetId(),
						Timestamp: log.GetTimestamp(),
						Level:     log.GetLevel(),
						Message:   log.GetMessage(),
						Source:    log.GetSource(),
						Metadata:  metadata,
					}

					bytes, err := json.Marshal(entry)
					if err == nil {
						if err := conn.SetWriteDeadline(time.Now().Add(10 * time.Second)); err != nil {
							return
						}
						if err := conn.WriteMessage(websocket.TextMessage, bytes); err != nil {
							logging.GetLogger().Error("failed to write db log message to websocket", "error", err)
							return
						}
						sentIDs[log.GetId()] = struct{}{}
					}
				}
			}
		}

		// 2. Send Memory History (Recent Live) - Deduplicated
		for _, msg := range history {
			var entry logging.LogEntry
			if err := json.Unmarshal(msg, &entry); err == nil {
				if _, sent := sentIDs[entry.ID]; sent {
					continue
				}
			}

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
