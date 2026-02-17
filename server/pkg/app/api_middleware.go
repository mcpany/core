// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"encoding/json"
	"net/http"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/logging"
	"github.com/mcpany/core/server/pkg/storage"
)

func (a *Application) handleMiddleware(store storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			// List Middlewares
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
			_ = json.NewEncoder(w).Encode(middlewares)

		case http.MethodPost:
			// Update Middlewares (Replace list)
			var middlewares []*configv1.Middleware
			if err := json.NewDecoder(r.Body).Decode(&middlewares); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
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
				logging.GetLogger().Error("failed to save global settings", "error", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}

			// Reload to apply changes
			if err := a.ReloadConfig(r.Context(), a.fs, a.configPaths); err != nil {
				logging.GetLogger().Error("failed to reload config after middleware update", "error", err)
			}

			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(middlewares)

		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}
}
