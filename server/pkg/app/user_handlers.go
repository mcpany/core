// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"encoding/json"
	"net/http"

	config_v1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/auth"
	"github.com/mcpany/core/server/pkg/logging"
)

// HandleGetUserPreferences returns the preferences for the authenticated user.
func (a *Application) HandleGetUserPreferences(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, ok := auth.UserFromContext(ctx)
	if !ok || userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	user, err := a.Storage.GetUser(ctx, userID)
	if err != nil {
		// If user doesn't exist in DB (e.g. system-admin implicit), return empty preferences
		logging.GetLogger().Debug("User not found in storage, returning empty preferences", "user_id", userID, "error", err)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{})
		return
	}

	prefs := user.GetPreferences()
	if prefs == nil {
		prefs = make(map[string]string)
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(prefs)
}

// HandleUpdateUserPreferences updates the preferences for the authenticated user.
func (a *Application) HandleUpdateUserPreferences(w http.ResponseWriter, r *http.Request) {
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

	// Try to get existing user
	user, err := a.Storage.GetUser(ctx, userID)
	if err != nil || user == nil {
		// Assume user doesn't exist, create one
		logging.GetLogger().Info("Creating new user entry for preferences", "user_id", userID)

		var roles []string
		var profileIDs []string
		var authConfig *config_v1.Authentication

		// Basic attempt to preserve roles if we can get them from AuthManager
		if existingUser, found := a.AuthManager.GetUser(userID); found {
			roles = existingUser.GetRoles()
			profileIDs = existingUser.GetProfileIds()
			authConfig = existingUser.GetAuthentication()
		}

		user = config_v1.User_builder{
			Id:             &userID,
			Preferences:    newPrefs,
			Roles:          roles,
			ProfileIds:     profileIDs,
			Authentication: authConfig,
		}.Build()

		if err := a.Storage.CreateUser(ctx, user); err != nil {
			logging.GetLogger().Error("Failed to create user", "error", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
	} else {
		// User exists, update preferences
		// Since we can't easily mutate the map in place if it's hidden/opaque without using Setters,
		// and SetPreferences takes a map, we should prepare the map.
		currentPrefs := user.GetPreferences()
		if currentPrefs == nil {
			currentPrefs = make(map[string]string)
		}
		for k, v := range newPrefs {
			currentPrefs[k] = v
		}
		user.SetPreferences(currentPrefs)

		if err := a.Storage.UpdateUser(ctx, user); err != nil {
			logging.GetLogger().Error("Failed to update user", "error", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
	}

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(user.GetPreferences())
}
