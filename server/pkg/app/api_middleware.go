// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"encoding/json"
	"net/http"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/logging"
	"github.com/mcpany/core/server/pkg/storage"
	"google.golang.org/protobuf/encoding/protojson"
)

func (a *Application) handleMiddleware(store storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			settings, err := store.GetGlobalSettings(r.Context())
			if err != nil {
				logging.GetLogger().Error("failed to get global settings for middleware", "error", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}
			if settings == nil {
				settings = configv1.GlobalSettings_builder{}.Build()
			}
			middlewares := settings.GetMiddlewares()
			if middlewares == nil {
				middlewares = []*configv1.Middleware{}
			}
			w.Header().Set("Content-Type", "application/json")

			// Marshal slice using protojson
			var buf []byte
			buf = append(buf, '[')
			opts := protojson.MarshalOptions{UseProtoNames: true, EmitUnpopulated: true}
			for i, m := range middlewares {
				if i > 0 {
					buf = append(buf, ',')
				}
				b, err := opts.Marshal(m)
				if err != nil {
					logging.GetLogger().Error("failed to marshal middleware", "error", err)
					continue
				}
				buf = append(buf, b...)
			}
			buf = append(buf, ']')
			_, _ = w.Write(buf)

		case http.MethodPost:
			var rawList []json.RawMessage
			if err := json.NewDecoder(r.Body).Decode(&rawList); err != nil {
				http.Error(w, "invalid request body", http.StatusBadRequest)
				return
			}

			var middlewares []*configv1.Middleware
			for _, raw := range rawList {
				var m configv1.Middleware
				if err := protojson.Unmarshal(raw, &m); err != nil {
					http.Error(w, "invalid middleware format: "+err.Error(), http.StatusBadRequest)
					return
				}
				middlewares = append(middlewares, &m)
			}

			// Load existing settings to preserve other fields
			settings, err := store.GetGlobalSettings(r.Context())
			if err != nil {
				logging.GetLogger().Error("failed to get global settings", "error", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}
			if settings == nil {
				settings = configv1.GlobalSettings_builder{}.Build()
			}

			// Update middlewares
			settings.SetMiddlewares(middlewares)

			if err := store.SaveGlobalSettings(r.Context(), settings); err != nil {
				logging.GetLogger().Error("failed to save global settings with new middlewares", "error", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}

			// Trigger reload
			if err := a.ReloadConfig(r.Context(), a.fs, a.configPaths); err != nil {
				logging.GetLogger().Error("failed to reload config after middleware update", "error", err)
			}

			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("{}"))
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	}
}
