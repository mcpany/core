// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"net/http"

	"github.com/mcpany/core/server/pkg/auth"
	"github.com/mcpany/core/server/pkg/logging"
	"github.com/mcpany/core/server/pkg/storage"
	"github.com/mcpany/core/server/pkg/util"
)

// handleMe returns the current authenticated user.
// GET /api/v1/users/me
func (a *Application) handleMe(store storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Get current user ID from context
		userID, ok := auth.UserFromContext(r.Context())
		if !ok {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		user, err := store.GetUser(r.Context(), userID)
		if err != nil {
			logging.GetLogger().Error("failed to get current user", "id", userID, "error", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		if user == nil {
			http.Error(w, "User not found", http.StatusNotFound)
			return
		}

		writeJSON(w, http.StatusOK, util.SanitizeUser(user))
	}
}
