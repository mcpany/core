// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"encoding/json"
	"io"
	"net/http"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/auth"
	"github.com/mcpany/core/server/pkg/logging"
	"github.com/mcpany/core/server/pkg/storage"
)

// handleUserPreferences handles GET and POST requests for user preferences.
func (a *Application) handleUserPreferences(store storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		userID, ok := auth.UserFromContext(ctx)
		if !ok {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Ensure user exists in DB (for implicit system-admin)
		user, err := store.GetUser(ctx, userID)
		if err != nil {
			logging.GetLogger().Error("failed to get user", "error", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		if user == nil {
			// If it's the implicit system-admin, create a placeholder
			if userID == "system-admin" {
				user = &configv1.User{
					// Use setter to ensure protobuf internals are set if needed, though struct literal is usually fine for pointers
				}
				user.SetId("system-admin")
				user.SetRoles([]string{"admin"})
				user.SetPreferences(make(map[string]string))

				if err := store.CreateUser(ctx, user); err != nil {
					logging.GetLogger().Error("failed to create system-admin user", "error", err)
					http.Error(w, "Internal Server Error", http.StatusInternalServerError)
					return
				}
				// Reload config to sync AuthManager with new DB state
				if err := a.ReloadConfig(ctx, a.fs, a.configPaths); err != nil {
					logging.GetLogger().Error("failed to reload config after system-admin creation", "error", err)
				}
			} else {
				// For other users, if they are authenticated but not in DB (e.g. from config file)
				if u, found := a.AuthManager.GetUser(userID); found {
					// Promote config user to DB user to enable persistence
					// We must copy existing fields to ensure we don't lose auth/roles when DB overrides Config
					user = &configv1.User{}
					user.SetId(u.GetId())
					user.SetAuthentication(u.GetAuthentication())
					user.SetProfileIds(u.GetProfileIds())
					user.SetRoles(u.GetRoles())
					user.SetPreferences(make(map[string]string))

					if err := store.CreateUser(ctx, user); err != nil {
						logging.GetLogger().Error("failed to persist config user to db", "error", err)
						http.Error(w, "Internal Server Error", http.StatusInternalServerError)
						return
					}
					// Reload config to sync AuthManager with new DB state
					if err := a.ReloadConfig(ctx, a.fs, a.configPaths); err != nil {
						logging.GetLogger().Error("failed to reload config after user promotion", "error", err)
					}
				} else {
					http.NotFound(w, r)
					return
				}
			}
		}

		switch r.Method {
		case http.MethodGet:
			prefs := user.GetPreferences()
			if prefs == nil {
				prefs = make(map[string]string)
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(prefs)

		case http.MethodPost:
			// Expecting map[string]string JSON body
			var newPrefs map[string]string
			if err := json.NewDecoder(r.Body).Decode(&newPrefs); err != nil && err != io.EOF {
				http.Error(w, "Invalid JSON", http.StatusBadRequest)
				return
			}

			// Merge with existing preferences
			currentPrefs := user.GetPreferences()
			if currentPrefs == nil {
				currentPrefs = make(map[string]string)
			}
			for k, v := range newPrefs {
				currentPrefs[k] = v
			}
			user.SetPreferences(currentPrefs)

			if err := store.UpdateUser(ctx, user); err != nil {
				logging.GetLogger().Error("failed to update user preferences", "error", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(currentPrefs)

		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}
}
