// Copyright 2026 Author(s) of MCP Any
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

// handleGetUserPreferences retrieves the preferences for the authenticated user.
func (a *Application) handleGetUserPreferences(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, ok := auth.UserFromContext(ctx)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	user, err := a.Storage.GetUser(ctx, userID)
	if err != nil {
		logging.GetLogger().Error("Failed to get user from storage", "user_id", userID, "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	if user == nil {
		// If user not found in DB, return empty preferences.
		logging.GetLogger().Debug("User not found in storage for preferences fetch", "user_id", userID)
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
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var prefs map[string]string
	if err := json.NewDecoder(r.Body).Decode(&prefs); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	user, err := a.Storage.GetUser(ctx, userID)
	if err != nil {
		logging.GetLogger().Error("Failed to get user from storage", "user_id", userID, "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	if user == nil {
		// If user not found (e.g. system-admin), create it using Builder.
		logging.GetLogger().Info("User not found in storage, creating new user for preferences", "user_id", userID)

		user = configv1.User_builder{
			Id:          proto.String(userID),
			Preferences: prefs,
		}.Build()

		if err := a.Storage.CreateUser(ctx, user); err != nil {
			logging.GetLogger().Error("Failed to create user", "error", err)
			http.Error(w, "Failed to save preferences", http.StatusInternalServerError)
			return
		}
	} else {
		// Update existing user.
		currentPrefs := user.GetPreferences()
		if currentPrefs == nil {
			currentPrefs = make(map[string]string)
		}

		newPrefs := make(map[string]string)
		for k, v := range currentPrefs {
			newPrefs[k] = v
		}
		for k, v := range prefs {
			newPrefs[k] = v
		}

		user.SetPreferences(newPrefs)

		if err := a.Storage.UpdateUser(ctx, user); err != nil {
			logging.GetLogger().Error("Failed to update user", "error", err)
			http.Error(w, "Failed to save preferences", http.StatusInternalServerError)
			return
		}
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(user.GetPreferences())
}
