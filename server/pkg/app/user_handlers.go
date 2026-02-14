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

// handleGetUserPreferences returns the current user's preferences.
func (a *Application) handleGetUserPreferences(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, ok := auth.UserFromContext(ctx)
	if !ok || userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Try to get from storage (DB) first
	user, err := a.Storage.GetUser(ctx, userID)
	if err != nil {
		// If not in DB, check in-memory (Config)
		if a.AuthManager != nil {
			if memUser, found := a.AuthManager.GetUser(userID); found {
				w.Header().Set("Content-Type", "application/json")
				prefs := memUser.GetPreferences()
				if prefs == nil {
					prefs = make(map[string]string)
				}
				_ = json.NewEncoder(w).Encode(prefs)
				return
			}
		}

		// If not found anywhere (e.g. implicit system-admin), return empty
		logging.GetLogger().Debug("User not found in storage or memory, returning empty preferences", "user_id", userID)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	prefs := user.GetPreferences()
	if prefs == nil {
		prefs = make(map[string]string)
	}
	_ = json.NewEncoder(w).Encode(prefs)
}

// handleUpdateUserPreferences updates the current user's preferences.
func (a *Application) handleUpdateUserPreferences(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, ok := auth.UserFromContext(ctx)
	if !ok || userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var newPrefs map[string]string
	if err := json.NewDecoder(r.Body).Decode(&newPrefs); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Fetch existing user from STORAGE first
	user, err := a.Storage.GetUser(ctx, userID)
	if err != nil || user == nil {
		// User not in DB. Check AuthManager for in-memory definition (from config).
		// We need to create a new DB record, preserving existing config if any.

		// Start with basic user
		userBuilder := configv1.User_builder{
			Id:          proto.String(userID),
			Preferences: make(map[string]string),
		}

		if a.AuthManager != nil {
			if memUser, found := a.AuthManager.GetUser(userID); found {
				// Copy fields from memory user to persist them in DB
				// Use proto.Clone to be safe, but we need to cast back
				cloned := proto.Clone(memUser).(*configv1.User)
				userBuilder.Authentication = cloned.GetAuthentication()
				userBuilder.ProfileIds = cloned.GetProfileIds()
				userBuilder.Roles = cloned.GetRoles()
				// Merge existing preferences from config
				if cloned.GetPreferences() != nil {
					for k, v := range cloned.GetPreferences() {
						userBuilder.Preferences[k] = v
					}
				}
			}
		}

		// Apply new preferences
		for k, v := range newPrefs {
			userBuilder.Preferences[k] = v
		}

		user = userBuilder.Build()

		logging.GetLogger().Info("Creating user record for preferences", "user_id", userID)
		if err := a.Storage.CreateUser(ctx, user); err != nil {
			logging.GetLogger().Error("Failed to create user for preferences", "user_id", userID, "error", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
	} else {
		// Update existing user in DB
		currentPrefs := user.GetPreferences()
		if currentPrefs == nil {
			currentPrefs = make(map[string]string)
		}
		// Merge new preferences
		for k, v := range newPrefs {
			currentPrefs[k] = v
		}
		user.SetPreferences(currentPrefs)

		if err := a.Storage.UpdateUser(ctx, user); err != nil {
			logging.GetLogger().Error("Failed to update user preferences", "user_id", userID, "error", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
	}

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(user.GetPreferences())
}
