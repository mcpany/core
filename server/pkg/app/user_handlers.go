// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"encoding/json"
	"fmt"
	"net/http"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/auth"
	"github.com/mcpany/core/server/pkg/logging"
)

// HandleGetUserPreferences returns the preferences for the current user.
func (a *Application) HandleGetUserPreferences(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, ok := auth.UserFromContext(ctx)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	user, err := a.Storage.GetUser(ctx, userID)
	if err != nil || user == nil {
		// If user not found (e.g. system-admin), return empty preferences instead of error
		// This allows the UI to function even if the user record doesn't exist yet.
		logging.GetLogger().Debug("User not found in storage, returning empty preferences", "user_id", userID)
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

// HandleUpdateUserPreferences updates the preferences for the current user.
func (a *Application) HandleUpdateUserPreferences(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, ok := auth.UserFromContext(ctx)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var newPrefs map[string]string
	if err := json.NewDecoder(r.Body).Decode(&newPrefs); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Fetch existing user
	user, err := a.Storage.GetUser(ctx, userID)
	if err != nil || user == nil {
		// If user not found, we might need to create it (e.g. for system-admin)
		// We assume that if we are authenticated, we are allowed to have preferences.
		logging.GetLogger().Info("User not found in storage, creating new user record for preferences", "user_id", userID)

		user = &configv1.User{}
		user.SetId(userID)
		user.SetPreferences(newPrefs)
		// We don't set authentication or roles here as they might be managed elsewhere
		// or implied (like system-admin).

		if err := a.Storage.CreateUser(ctx, user); err != nil {
			logging.GetLogger().Error("Failed to create user for preferences", "user_id", userID, "error", err)
			http.Error(w, fmt.Sprintf("Failed to save preferences: %v", err), http.StatusInternalServerError)
			return
		}
	} else {
		// Update existing user
		prefs := user.GetPreferences()
		if prefs == nil {
			prefs = make(map[string]string)
		}

		// Merge preferences
		for k, v := range newPrefs {
			prefs[k] = v
		}
		user.SetPreferences(prefs)

		if err := a.Storage.UpdateUser(ctx, user); err != nil {
			logging.GetLogger().Error("Failed to update user preferences", "user_id", userID, "error", err)
			http.Error(w, fmt.Sprintf("Failed to save preferences: %v", err), http.StatusInternalServerError)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(user.GetPreferences())
}
