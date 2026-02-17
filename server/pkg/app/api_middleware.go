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
				logging.GetLogger().Error("failed to get global settings", "error", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}
			if settings == nil {
				settings = configv1.GlobalSettings_builder{}.Build()
			}
			middlewares := settings.GetMiddlewares()
			w.Header().Set("Content-Type", "application/json")
			opts := protojson.MarshalOptions{UseProtoNames: true, EmitUnpopulated: true}

			var buf []byte
			buf = append(buf, '[')
			for i, m := range middlewares {
				if i > 0 {
					buf = append(buf, ',')
				}
				mb, err := opts.Marshal(m)
				if err != nil {
					logging.GetLogger().Error("failed to marshal middleware", "error", err)
					continue
				}
				buf = append(buf, mb...)
			}
			buf = append(buf, ']')
			_, _ = w.Write(buf)

		case http.MethodPost:
			// Expects a list of Middleware
			body, err := readBodyWithLimit(w, r, 1048576)
			if err != nil {
				return
			}

			// We need to unmarshal a JSON array of Middleware objects.
			// protojson doesn't support unmarshaling arrays directly.
			// So we unmarshal to []json.RawMessage, then unmarshal each.
			var rawMessages []json.RawMessage
			if err := json.Unmarshal(body, &rawMessages); err != nil {
				http.Error(w, "Invalid JSON array: "+err.Error(), http.StatusBadRequest)
				return
			}

			var newMiddlewares []*configv1.Middleware
			for _, raw := range rawMessages {
				var m configv1.Middleware
				if err := protojson.Unmarshal(raw, &m); err != nil {
					http.Error(w, "Invalid middleware object: "+err.Error(), http.StatusBadRequest)
					return
				}
				newMiddlewares = append(newMiddlewares, &m)
			}

			// Update Global Settings
			// We need to fetch, update, save.
			// Use a transaction if possible? Storage interface doesn't expose it easily here.
			// Optimistic locking?
			settings, err := store.GetGlobalSettings(r.Context())
			if err != nil {
				logging.GetLogger().Error("failed to get global settings for update", "error", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}
			if settings == nil {
				settings = configv1.GlobalSettings_builder{}.Build()
			}

			settings.SetMiddlewares(newMiddlewares)

			if err := store.SaveGlobalSettings(r.Context(), settings); err != nil {
				logging.GetLogger().Error("failed to save global settings", "error", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}

			// Reload config
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
