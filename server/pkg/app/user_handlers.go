// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"encoding/json"
	"net/http"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/auth"
	"github.com/mcpany/core/server/pkg/logging"
	"google.golang.org/protobuf/proto"
)

// handleGetUserPreferences returns the user preferences.
func (a *Application) handleGetUserPreferences(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID, ok := auth.UserFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// First check DB
	user, err := a.Storage.GetUser(r.Context(), userID)
	if err != nil {
		logging.GetLogger().Error("Failed to get user from storage", "user_id", userID, "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	if user == nil {
		// Not in DB. Check AuthManager (config users)
		if authUser, found := a.AuthManager.GetUser(userID); found {
			// Found in config. Return its preferences (likely empty)
			w.Header().Set("Content-Type", "application/json")
			if authUser.GetPreferences() == nil {
				_ = json.NewEncoder(w).Encode(map[string]string{})
			} else {
				_ = json.NewEncoder(w).Encode(authUser.GetPreferences())
			}
			return
		}

		// If user doesn't exist (e.g. system-admin implicit), we return empty preferences
		if userID == "system-admin" {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]string{})
			return
		}

		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if user.GetPreferences() == nil {
		_ = json.NewEncoder(w).Encode(map[string]string{})
	} else {
		_ = json.NewEncoder(w).Encode(user.GetPreferences())
	}
}

// handleUpdateUserPreferences updates the user preferences.
func (a *Application) handleUpdateUserPreferences(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID, ok := auth.UserFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var prefs map[string]string
	if err := json.NewDecoder(r.Body).Decode(&prefs); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Ensure user exists in DB
	user, err := a.Storage.GetUser(r.Context(), userID)
	if err != nil {
		logging.GetLogger().Error("Failed to get user from storage", "user_id", userID, "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	if user == nil {
		// Not in DB. Check AuthManager (config users)
		if authUser, found := a.AuthManager.GetUser(userID); found {
			// Found in config. Promote to DB user.
			// Clone it to persist
			user = proto.Clone(authUser).(*configv1.User)

			// Update preferences
			currentPrefs := user.GetPreferences()
			if currentPrefs == nil {
				currentPrefs = make(map[string]string)
			}
			for k, v := range prefs {
				currentPrefs[k] = v
			}
			user.SetPreferences(currentPrefs)

			if err := a.Storage.CreateUser(r.Context(), user); err != nil {
				logging.GetLogger().Error("Failed to create user from config", "user_id", userID, "error", err)
				http.Error(w, "Failed to update preferences", http.StatusInternalServerError)
				return
			}
		} else if userID == "system-admin" {
			// Implicit admin, not in AuthManager
			user = configv1.User_builder{
				Id:          proto.String(userID),
				Preferences: prefs,
				Roles:       []string{"admin"},
			}.Build()

			if err := a.Storage.CreateUser(r.Context(), user); err != nil {
				logging.GetLogger().Error("Failed to create system-admin user", "error", err)
				http.Error(w, "Failed to update preferences", http.StatusInternalServerError)
				return
			}
		} else {
			// User really not found
			logging.GetLogger().Error("User not found for update", "user_id", userID)
			http.Error(w, "User not found", http.StatusNotFound)
			return
		}
	} else {
		// Update existing DB user
		currentPrefs := user.GetPreferences()
		if currentPrefs == nil {
			currentPrefs = make(map[string]string)
		}
		for k, v := range prefs {
			currentPrefs[k] = v
		}
		user.SetPreferences(currentPrefs)

		if err := a.Storage.UpdateUser(r.Context(), user); err != nil {
			logging.GetLogger().Error("Failed to update user", "user_id", userID, "error", err)
			http.Error(w, "Failed to update preferences", http.StatusInternalServerError)
			return
		}
	}

	w.WriteHeader(http.StatusOK)
}
