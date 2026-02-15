// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"encoding/json"
	"net/http"

	configv1 "github.com/mcpany/core/proto/config/v1"
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

	// Try to get user from storage (DB)
	user, err := a.Storage.GetUser(ctx, userID)
	if err != nil || user == nil {
		// If user not found in DB, check if it exists in memory (e.g. from config)
		if inMemUser, found := a.AuthManager.GetUser(userID); found {
			w.Header().Set("Content-Type", "application/json")
			if inMemUser.GetPreferences() == nil {
				_, _ = w.Write([]byte("{}"))
			} else {
				_ = json.NewEncoder(w).Encode(inMemUser.GetPreferences())
			}
			return
		}

		// If not in memory either, return empty.
		logging.GetLogger().Debug("User not found for preferences", "user_id", userID)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte("{}"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if user.GetPreferences() == nil {
		_, _ = w.Write([]byte("{}"))
	} else {
		_ = json.NewEncoder(w).Encode(user.GetPreferences())
	}
}

// HandleUpdateUserPreferences updates the preferences for the authenticated user.
func (a *Application) HandleUpdateUserPreferences(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, ok := auth.UserFromContext(ctx)
	if !ok || userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var prefs map[string]string
	if err := json.NewDecoder(r.Body).Decode(&prefs); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Try to get user from storage to verify existence and avoid overwriting other fields if possible
	user, err := a.Storage.GetUser(ctx, userID)
	if err != nil || user == nil {
		// User not found in storage. Create it.
		// We MUST copy existing auth config from AuthManager if available, otherwise the user loses access
		// because DB user overrides Config user.
		logging.GetLogger().Info("Creating user in storage for preferences", "user_id", userID)

		newUser := configv1.User_builder{
			Id:          &userID,
			Preferences: prefs,
		}.Build()

		if existingUser, found := a.AuthManager.GetUser(userID); found {
			newUser.SetAuthentication(existingUser.GetAuthentication())
			newUser.SetRoles(existingUser.GetRoles())
			newUser.SetProfileIds(existingUser.GetProfileIds())
		}
		// If user is not in AuthManager either (e.g. implicit system-admin), we create a minimal user.
		// Implicit system-admin has no Auth config (uses global key) and implicit roles.
		// So storing just ID and Preferences is fine for system-admin.

		if err := a.Storage.CreateUser(ctx, newUser); err != nil {
			logging.GetLogger().Error("Failed to create user for preferences", "error", err)
			http.Error(w, "Failed to save preferences", http.StatusInternalServerError)
			return
		}

		// We also need to update AuthManager immediately so the change is reflected without restart?
		// AuthManager.SetUsers overwrites everything.
		// Ideally we should update AuthManager too.
		// But SetUsers takes a list.
		// We can't easily update single user in AuthManager.
		// However, AuthManager.GetUser returns a pointer?
		// "u, ok := am.users[id]" returns pointer.
		// If we modify it in place?
		// "am.users" is protected by RWMutex.
		// We shouldn't modify it directly without lock.
		// For now, persistence is enough. On restart it loads.
		// But for the current session, we want the preferences to be available via GetUserPreferences?
		// HandleGetUserPreferences checks DB first. So it's fine.
	} else {
		// User exists in DB. Update preferences.
		user.SetPreferences(prefs)
		if err := a.Storage.UpdateUser(ctx, user); err != nil {
			logging.GetLogger().Error("Failed to update user preferences", "error", err)
			http.Error(w, "Failed to save preferences", http.StatusInternalServerError)
			return
		}
	}

	w.WriteHeader(http.StatusOK)
}
