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

const systemAdminUserID = "system-admin"

// HandleGetUserPreferences returns the preferences for the current user.
func (a *Application) HandleGetUserPreferences(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, ok := auth.UserFromContext(ctx)
	if !ok || userID == "" {
		http.Error(w, "User not found in context", http.StatusUnauthorized)
		return
	}

	log := logging.GetLogger()

	// Handle system-admin special case
	if userID == systemAdminUserID {
		// Try to fetch from storage first
		user, err := a.Storage.GetUser(ctx, userID)
		// Check for nil user explicitly as SQLite store returns nil, nil for not found
		if err != nil || user == nil {
			// If not found, return empty preferences
			// Or check specifically for NotFound error?
			// Assuming generic error might mean not found or DB error.
			// For system-admin, if not persisted, we return empty default.
			log.Debug("system-admin user not found in storage, returning empty preferences")
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]string{})
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(user.GetPreferences())
		return
	}

	// For regular users, they must exist in AuthManager, but maybe not in Storage if loaded from file?
	// AuthManager loads users from Config + Storage.
	// We need to fetch the LATEST state from Storage if possible, or fall back to AuthManager's copy.
	// But AuthManager's copy might be stale if updated via API?
	// Actually, AuthManager is updated on reload.
	// But Storage is the source of truth for persistent updates.

	user, err := a.Storage.GetUser(ctx, userID)
	// Check for nil user explicitly as SQLite store returns nil, nil for not found
	if err != nil || user == nil {
		// If user is in AuthManager (file config) but not in Storage (DB),
		// we should return what's in AuthManager (likely empty preferences unless defined in config).
		if u, found := a.AuthManager.GetUser(userID); found {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(u.GetPreferences())
			return
		}
		if err != nil {
			log.Error("User not found in storage or auth manager", "user_id", userID, "error", err)
		}
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(user.GetPreferences())
}

// HandleUpdateUserPreferences updates the preferences for the current user.
func (a *Application) HandleUpdateUserPreferences(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, ok := auth.UserFromContext(ctx)
	if !ok || userID == "" {
		http.Error(w, "User not found in context", http.StatusUnauthorized)
		return
	}

	var newPrefs map[string]string
	if err := json.NewDecoder(r.Body).Decode(&newPrefs); err != nil {
		http.Error(w, "Invalid JSON body", http.StatusBadRequest)
		return
	}

	log := logging.GetLogger()
	log.Info("Updating user preferences", "user_id", userID)

	// Fetch existing user to merge/update
	user, err := a.Storage.GetUser(ctx, userID)
	// Check for nil user explicitly as SQLite store returns nil, nil for not found
	if err != nil || user == nil {
		// User not in DB.
		// If it's system-admin, create it.
		// If it's a file-config user, we create a DB override/entry for them.
		// NOTE: Creating a DB entry for a file-config user effectively "shadows" or "extends" it depending on how LoadServices merges.
		// LoadServices appends DB users. If ID matches, it might result in duplicates in the list?
		// AuthManager.SetUsers does `users = append(users, dbUsers...)`.
		// If ID matches, AuthManager stores both?
		// server/pkg/auth/manager.go:
		// func (m *Manager) SetUsers(users []*configv1.User) { ... }
		// It iterates and sets in map. Last one wins?
		// Usually yes. So DB user (appended last) should win.
		// Let's verify `server/pkg/app/server.go`:
		// users = cfg.GetUsers() (File)
		// users = append(users, dbUsers...) (DB)
		// So DB users come AFTER file users.
		// If AuthManager uses a map keyed by ID, the last one overwrites.
		// So persisting to DB is the correct strategy to override file config.

		// Check if user exists in AuthManager (to get roles/etc to copy if we are "promoting" a file user to DB)
		authUser, found := a.AuthManager.GetUser(userID)
		if !found && userID != systemAdminUserID {
			http.Error(w, "User does not exist", http.StatusNotFound)
			return
		}

		if user == nil {
			// Create new user object
			user = configv1.User_builder{
				Id:          &userID,
				Preferences: newPrefs,
			}.Build()

			if found {
				// Copy fields from in-memory user to persist them
				// We don't want to lose roles/auth if we shadow it.
				// However, if we save to DB, next time we load, we load this DB user.
				// If the DB user is incomplete (missing roles), we might lose access?
				// But we are "merging" in AuthManager?
				// No, if AuthManager overwrites, it replaces the object.
				// So we MUST copy existing attributes.
				user.SetAuthentication(authUser.GetAuthentication())
				user.SetRoles(authUser.GetRoles())
				user.SetProfileIds(authUser.GetProfileIds())
			} else if userID == systemAdminUserID {
				// Default system admin
				user.SetRoles([]string{"admin"})
			}

			if err := a.Storage.CreateUser(ctx, user); err != nil {
				log.Error("Failed to create user in storage", "error", err)
				http.Error(w, "Failed to save preferences", http.StatusInternalServerError)
				return
			}
		}
	} else {
		// Update existing user
		// Merge preferences
		currentPrefs := user.GetPreferences()
		if currentPrefs == nil {
			currentPrefs = make(map[string]string)
		}
		for k, v := range newPrefs {
			currentPrefs[k] = v
		}
		user.SetPreferences(currentPrefs)

		if err := a.Storage.UpdateUser(ctx, user); err != nil {
			log.Error("Failed to update user in storage", "error", err)
			http.Error(w, "Failed to save preferences", http.StatusInternalServerError)
			return
		}
	}

	// Reload/Update AuthManager to reflect changes immediately
	// We just updated the DB. On next restart it loads fine.
	// But for current session?
	// AuthManager.SetUsers re-reads everything? No.
	// We should manually update the user in AuthManager.
	// But AuthManager doesn't expose UpdateUser public method except SetUsers.
	// Wait, we can just let it be?
	// Preferences are likely not used for Auth decisions (Roles/Profiles).
	// They are just data.
	// So updating AuthManager might not be strictly necessary if we only read preferences from this API endpoint
	// which reads from Storage (or falls back to AuthManager).
	// Since `HandleGetUserPreferences` reads from Storage first, we are good!

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(user.GetPreferences())
}
