// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"encoding/json"
	"net/http"

	v1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/auth"
	"github.com/mcpany/core/server/pkg/logging"
)

// handleGetUserPreferences returns the preferences for the authenticated user.
func (a *Application) handleGetUserPreferences(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, ok := auth.UserFromContext(ctx)
	if !ok || userID == "" {
		// Should be caught by auth middleware, but safety check
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	user, err := a.Storage.GetUser(ctx, userID)
	if err != nil || user == nil {
		// If user doesn't exist in DB (e.g. system-admin implicit), return empty preferences
		// Log debug just in case
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

// handleUpdateUserPreferences updates the preferences for the authenticated user.
func (a *Application) handleUpdateUserPreferences(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, ok := auth.UserFromContext(ctx)
	if !ok || userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var newPrefs map[string]string
	if err := json.NewDecoder(r.Body).Decode(&newPrefs); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Fetch existing user to merge or create
	user, err := a.Storage.GetUser(ctx, userID)

	// Storage.GetUser returns (nil, nil) if not found in some implementations (e.g. memory)
	if err != nil || user == nil {
		// User might not exist (e.g. system-admin), create it
		logging.GetLogger().Info("User not found, creating new user record for preferences", "user_id", userID)

		// Use builder or just new and setters
		user = &v1.User{}
		user.SetId(userID)
		user.SetPreferences(newPrefs)

		// If we can get user details from AuthManager, let's populate them to be safe
		if authUser, found := a.AuthManager.GetUser(userID); found {
			user.SetRoles(authUser.GetRoles())
			user.SetProfileIds(authUser.GetProfileIds())
			// We don't copy authentication usually as it might be sensitive or irrelevant for DB record if config-driven
		}

		if createErr := a.Storage.CreateUser(ctx, user); createErr != nil {
			logging.GetLogger().Error("Failed to create user for preferences", "user_id", userID, "error", createErr)
			http.Error(w, "Failed to save preferences", http.StatusInternalServerError)
			return
		}
	} else {
		// Update existing user
		currentPrefs := user.GetPreferences()
		if currentPrefs == nil {
			currentPrefs = make(map[string]string)
		}

		// Merge or Replace?
		// For dashboard layout, usually full replace of that specific key.
		// But here we accept a map.
		// Let's assume the client sends the *delta* or the full map of keys they want to update.
		for k, v := range newPrefs {
			currentPrefs[k] = v
		}
		user.SetPreferences(currentPrefs)

		if updateErr := a.Storage.UpdateUser(ctx, user); updateErr != nil {
			logging.GetLogger().Error("Failed to update user preferences", "user_id", userID, "error", updateErr)
			http.Error(w, "Failed to save preferences", http.StatusInternalServerError)
			return
		}
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(user.GetPreferences())
}
