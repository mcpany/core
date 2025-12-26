// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"io"
	"net/http"
	"strings"

	"encoding/json"
	"time"

	"github.com/gorilla/websocket"
	"github.com/mcpany/core/pkg/logging"
	"github.com/mcpany/core/pkg/logging/stream"
	"github.com/mcpany/core/pkg/storage/sqlite"
	configv1 "github.com/mcpany/core/proto/config/v1"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins for now
	},
}

// createAPIHandler creates a http.Handler for the config API.
func (a *Application) createAPIHandler(store *sqlite.Store) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/logs/stream", func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			logging.GetLogger().Error("failed to upgrade websocket", "error", err)
			return
		}
		defer func() {
			_ = conn.Close()
		}()

		broadcaster := stream.GetBroadcaster()
		ch := broadcaster.Subscribe()
		defer broadcaster.Unsubscribe(ch)

		// Send ping to keep connection alive
		go func() {
			ticker := time.NewTicker(30 * time.Second)
			defer ticker.Stop()
			for range ticker.C {
				if err := conn.WriteMessage(websocket.PingMessage, []byte{}); err != nil {
					return
				}
			}
		}()

		for entry := range ch {
			data, err := json.Marshal(entry)
			if err != nil {
				continue
			}
			if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
				break
			}
		}
	})

	mux.HandleFunc("/services", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			services, err := store.ListServices()
			if err != nil {
				logging.GetLogger().Error("failed to list services", "error", err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			var buf []byte
			buf = append(buf, '[')
			opts := protojson.MarshalOptions{UseProtoNames: true, EmitUnpopulated: false}
			for i, svc := range services {
				if i > 0 {
					buf = append(buf, ',')
				}
				b, err := opts.Marshal(svc)
				if err != nil {
					logging.GetLogger().Error("failed to marshal service", "error", err)
					continue
				}
				buf = append(buf, b...)
			}
			buf = append(buf, ']')
			_, _ = w.Write(buf)

		case http.MethodPost:
			var svc configv1.UpstreamServiceConfig
			body, _ := io.ReadAll(r.Body)
			if err := protojson.Unmarshal(body, &svc); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			if svc.GetName() == "" {
				http.Error(w, "name is required", http.StatusBadRequest)
				return
			}
			// Auto-generate ID if missing? Store handles it if we pass emtpy ID (fallback to name).
			// But creating UUID here might be better? For now name fallback is fine.

			if err := store.SaveService(&svc); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			// Trigger reload
			if err := a.ReloadConfig(a.fs, a.configPaths); err != nil {
				logging.GetLogger().Error("failed to reload config after save", "error", err)
				// Don't fail the request, but log error
			}

			w.WriteHeader(http.StatusCreated)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/services/", func(w http.ResponseWriter, r *http.Request) {
		name := strings.TrimPrefix(r.URL.Path, "/services/")
		if name == "" {
			http.Error(w, "name required", http.StatusBadRequest)
			return
		}

		switch r.Method {
		case http.MethodGet:
			svc, err := store.GetService(name)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			if svc == nil {
				http.NotFound(w, r)
				return
			}
			opts := protojson.MarshalOptions{UseProtoNames: true}
			b, _ := opts.Marshal(svc)
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write(b)
		case http.MethodPut:
			var svc configv1.UpstreamServiceConfig
			body, _ := io.ReadAll(r.Body)
			if err := protojson.Unmarshal(body, &svc); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			svc.Name = proto.String(name) // Force name match
			if err := store.SaveService(&svc); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			if err := a.ReloadConfig(a.fs, a.configPaths); err != nil {
				logging.GetLogger().Error("failed to reload config after update", "error", err)
			}
			w.WriteHeader(http.StatusOK)
		case http.MethodDelete:
			if err := store.DeleteService(name); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			if err := a.ReloadConfig(a.fs, a.configPaths); err != nil {
				logging.GetLogger().Error("failed to reload config after delete", "error", err)
			}
			w.WriteHeader(http.StatusNoContent)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	})

	return mux
}
