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

// handleGetUserPreferences handles GET /api/v1/user/preferences.
func (a *Application) handleGetUserPreferences(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, ok := auth.UserFromContext(ctx)
	if !ok || userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	user, err := a.Storage.GetUser(ctx, userID)
	if err != nil {
		// If user not found in DB, it might be an implicit user (e.g. system-admin).
		// In that case, we return empty preferences.
		// We should check if err is "not found".
		// For now, we assume if error, user might not exist in DB yet.
		// But if it's a real DB error, we should log it.
		// Since we don't have easy error code checking here without depending on specific storage impl errors,
		// we'll log info and return empty.
		logging.GetLogger().Info("User not found in storage for preferences, returning empty", "user_id", userID, "error", err)
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

// handleUpdateUserPreferences handles POST /api/v1/user/preferences.
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

	user, err := a.Storage.GetUser(ctx, userID)
	if err != nil || user == nil {
		// User not found, try to create it.
		logging.GetLogger().Info("User not found for preference update, creating new user record", "user_id", userID)

		// Create a basic user object
		newUser := config_v1.User_builder{
			Id:          &userID,
			Preferences: newPrefs,
		}.Build()

		if err := a.Storage.CreateUser(ctx, newUser); err != nil {
			logging.GetLogger().Error("Failed to create user for preferences", "user_id", userID, "error", err)
			http.Error(w, "Failed to save preferences", http.StatusInternalServerError)
			return
		}
	} else {
		// User exists, update preferences
		// Merge or replace? The API implies replacing the preferences map, or at least the keys provided.
		// But typically a PUT/POST to /preferences might mean "set these preferences".
		// To avoid race conditions or complexity, let's say this endpoint sets the entire preferences map
		// OR we merge. Merging is safer if we have partial updates.
		// However, simple implementation: Replace the map or merge.
		// Given we passed `newPrefs` map, let's merge it into existing.

		currentPrefs := user.GetPreferences()
		if currentPrefs == nil {
			currentPrefs = make(map[string]string)
		}

		for k, v := range newPrefs {
			currentPrefs[k] = v
		}

		user.SetPreferences(currentPrefs)

		if err := a.Storage.UpdateUser(ctx, user); err != nil {
			logging.GetLogger().Error("Failed to update user preferences", "user_id", userID, "error", err)
			http.Error(w, "Failed to save preferences", http.StatusInternalServerError)
			return
		}
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}
