// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"github.com/mcpany/core/server/pkg/logging"
)

func (a *Application) handleTraces() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Prefer TraceExporter if available
		if a.TraceExporter != nil {
			traces := a.TraceExporter.GetTraces()
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(traces)
			return
		}

		// Fallback to empty if not configured
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte("[]"))
	}
}

func (a *Application) handleTracesWS() http.HandlerFunc {
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

		// WebSocket support for TraceExporter is not yet implemented.
		// For now, we send a close message.
		// The previous implementation relied on audit log broadcasting which is flat.
		// To support nested trace updates live, we need a broadcaster in TraceExporter.

		_ = conn.WriteControl(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, "Live updates not supported yet"), time.Now().Add(time.Second))
	}
}
