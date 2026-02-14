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

// HandleGetUserPreferences returns the preferences for the current user.
func (a *Application) HandleGetUserPreferences(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logging.GetLogger()

	// 1. Identify User
	userID, ok := auth.UserFromContext(ctx)
	if !ok || userID == "" {
		// Should have been caught by auth middleware, but double check
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// 2. Fetch User from Storage
	user, err := a.Storage.GetUser(ctx, userID)
	if err != nil {
		log.Error("Failed to get user for preferences", "user_id", userID, "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	if user == nil {
		// 2a. Special Handling for Implicit "system-admin"
		// If the user is "system-admin" (default global auth) and doesn't exist in DB yet,
		// we return empty preferences instead of 404.
		if userID == "system-admin" {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]string{})
			return
		}

		// Real user not found
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	// 3. Return Preferences
	prefs := user.GetPreferences()
	if prefs == nil {
		prefs = make(map[string]string)
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(prefs); err != nil {
		log.Error("Failed to encode preferences", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// HandleUpdateUserPreferences updates the preferences for the current user.
func (a *Application) HandleUpdateUserPreferences(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logging.GetLogger()

	// 1. Identify User
	userID, ok := auth.UserFromContext(ctx)
	if !ok || userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// 2. Parse Request Body
	var newPrefs map[string]string
	if err := json.NewDecoder(r.Body).Decode(&newPrefs); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// 3. Fetch or Create User
	user, err := a.Storage.GetUser(ctx, userID)
	if err != nil {
		log.Error("Failed to get user for preferences update", "user_id", userID, "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	if user == nil {
		if userID == "system-admin" {
			// Create implicit admin user on first save
			log.Info("Creating implicit system-admin user for preferences storage")
			user = config_v1.User_builder{
				Id:          proto.String("system-admin"),
				Preferences: newPrefs,
				Roles:       []string{"admin"}, // Ensure they keep admin role
			}.Build()

			if err := a.Storage.CreateUser(ctx, user); err != nil {
				log.Error("Failed to create system-admin user", "error", err)
				http.Error(w, "Failed to save preferences", http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusOK)
			return
		}

		// Real user not found
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	// 4. Update Preferences
	currentPrefs := user.GetPreferences()
	if currentPrefs == nil {
		currentPrefs = make(map[string]string)
	}

	for k, v := range newPrefs {
		currentPrefs[k] = v
	}

	user.SetPreferences(currentPrefs)

	// 5. Persist
	if err := a.Storage.UpdateUser(ctx, user); err != nil {
		log.Error("Failed to update user preferences", "user_id", userID, "error", err)
		http.Error(w, "Failed to update preferences", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
