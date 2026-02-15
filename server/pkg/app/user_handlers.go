// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"encoding/json"
	"net/http"

	config_v1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/auth"
	"github.com/mcpany/core/server/pkg/logging"
	"google.golang.org/protobuf/proto"
)

// handleGetUserPreferences returns the preferences for the authenticated user.
func (a *Application) handleGetUserPreferences(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, ok := auth.UserFromContext(ctx)
	if !ok || userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	user, err := a.Storage.GetUser(ctx, userID)
	if err != nil {
		// If user retrieval fails, it might be due to user not existing in DB yet (even if auth passed).
		// We treat this as "no preferences" rather than an error to prevent UI breakage.
		logging.GetLogger().Warn("Failed to get user for preferences, treating as empty", "user_id", userID, "error", err)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte("{}"))
		return
	}

	if user == nil {
		// If user doesn't exist in DB (e.g. implicit system-admin), return empty preferences
		logging.GetLogger().Debug("User not found in storage, returning empty preferences", "user_id", userID)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte("{}"))
		return
	}

	prefs := user.GetPreferences()
	if prefs == nil {
		prefs = make(map[string]string)
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(prefs); err != nil {
		logging.GetLogger().Error("Failed to encode preferences", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
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

	user, err := a.Storage.GetUser(ctx, userID)
	if err != nil {
		logging.GetLogger().Error("Failed to get user", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	if user == nil {
		// User might not exist in DB yet (e.g. implicit system-admin or file-based user)
		// We create a new user record in DB to store preferences.
		logging.GetLogger().Info("User not found in storage, creating new record for preferences", "user_id", userID)

		newUser := config_v1.User_builder{
			Id:          proto.String(userID),
			Preferences: newPrefs,
		}.Build()

		if err := a.Storage.CreateUser(ctx, newUser); err != nil {
			logging.GetLogger().Error("Failed to create user for preferences", "error", err)
			http.Error(w, "Failed to save preferences", http.StatusInternalServerError)
			return
		}
	} else {
		// Update existing user
		// We need to merge preferences or replace?
		// The requirement implies persistence of the dashboard layout.
		// If we receive a full map, we can replace or merge.
		// Let's merge for flexibility, or replace if the client sends strict update.
		// For dashboard layout, usually the client sends the full state of specific keys.

		currentPrefs := user.GetPreferences()
		if currentPrefs == nil {
			currentPrefs = make(map[string]string)
		}

		for k, v := range newPrefs {
			currentPrefs[k] = v
		}

		user.SetPreferences(currentPrefs)

		if err := a.Storage.UpdateUser(ctx, user); err != nil {
			logging.GetLogger().Error("Failed to update user preferences", "error", err)
			http.Error(w, "Failed to save preferences", http.StatusInternalServerError)
			return
		}
	}

	w.WriteHeader(http.StatusOK)
}
