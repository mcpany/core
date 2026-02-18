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

			// Extract middlewares
			middlewares := settings.GetMiddlewares()

			w.Header().Set("Content-Type", "application/json")
			// We return just the list of middlewares
			opts := protojson.MarshalOptions{UseProtoNames: true, EmitUnpopulated: true}

			var buf []byte
			buf = append(buf, '[')
			for i, m := range middlewares {
				if i > 0 {
					buf = append(buf, ',')
				}
				b, _ := opts.Marshal(m)
				buf = append(buf, b...)
			}
			buf = append(buf, ']')
			_, _ = w.Write(buf)

		case http.MethodPost:
			var middlewares []*configv1.Middleware
			body, err := readBodyWithLimit(w, r, 1048576)
			if err != nil {
				return
			}

			// Unmarshal list of middlewares
			// protojson doesn't support unmarshaling a list directly easily without a wrapper message.
			// But we can unmarshal as generic JSON first and then convert, or expect a wrapper?
			// The client sends an array.
			// Let's use json to unmarshal into a temporary struct or map, then map to proto.

			var rawMiddlewares []map[string]any
			if err := json.Unmarshal(body, &rawMiddlewares); err != nil {
				http.Error(w, "Invalid JSON", http.StatusBadRequest)
				return
			}

			for _, rm := range rawMiddlewares {
				// We re-marshal each item to protojson
				b, _ := json.Marshal(rm)
				var m configv1.Middleware
				if err := protojson.Unmarshal(b, &m); err != nil {
					http.Error(w, "Invalid middleware format: "+err.Error(), http.StatusBadRequest)
					return
				}
				middlewares = append(middlewares, &m)
			}

			// Update Global Settings
			settings, err := store.GetGlobalSettings(r.Context())
			if err != nil {
				logging.GetLogger().Error("failed to get global settings for middleware update", "error", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}
			if settings == nil {
				settings = configv1.GlobalSettings_builder{}.Build()
			}

			settings.SetMiddlewares(middlewares)

			if err := store.SaveGlobalSettings(r.Context(), settings); err != nil {
				logging.GetLogger().Error("failed to save global settings with middlewares", "error", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}

			if err := a.ReloadConfig(r.Context(), a.fs, a.configPaths); err != nil {
				logging.GetLogger().Error("failed to reload config after middleware save", "error", err)
			}

			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("{}"))
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	}
}
