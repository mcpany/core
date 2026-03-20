// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"encoding/json"
	"net/http"

	"github.com/mcpany/core/server/pkg/auth"
	"github.com/mcpany/core/server/pkg/logging"
	"github.com/mcpany/core/server/pkg/storage"
	"github.com/mcpany/core/server/pkg/util"
)

// handleUserMe returns an HTTP handler for retrieving the current user's profile.
//
// Summary: Retrieves the authenticated user's profile.
//
// Parameters:
//   - store: storage.Storage. The storage interface.
//
// Returns:
//   - http.HandlerFunc: The HTTP handler function.
func (a *Application) handleUserMe(store storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		userID, ok := auth.UserFromContext(r.Context())
		if !ok {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Handle "system-admin" or similar virtual users if any
		if userID == "system-admin" {
			// Return a virtual admin user
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"id":    "system-admin",
				"name":  "System Admin",
				"email": "admin@localhost",
				"roles": []string{"admin"},
			})
			return
		}

		user, err := store.GetUser(r.Context(), userID)
		if err != nil {
			logging.GetLogger().Error("failed to get current user", "id", userID, "error", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		if user == nil {
			// This is weird: authenticated but user not found in DB?
			// Maybe deleted? Or auth provider sync issue.
			http.Error(w, "User not found", http.StatusNotFound)
			return
		}

		writeJSON(w, http.StatusOK, util.SanitizeUser(user))
	}
}
