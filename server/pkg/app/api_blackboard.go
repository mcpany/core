// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"encoding/json"
	"net/http"

	"github.com/mcpany/core/server/pkg/logging"
	"github.com/mcpany/core/server/pkg/middleware"
)

// handleBlackboardKeys returns an HTTP handler for retrieving all blackboard keys.
//
// Summary: HTTP handler for retrieving all blackboard keys.
//
// Returns:
//   - http.HandlerFunc: The HTTP handler function.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func (a *Application) handleBlackboardKeys() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		if a.standardMiddlewares == nil || a.standardMiddlewares.Blackboard == nil {
			http.Error(w, "blackboard store not initialized", http.StatusInternalServerError)
			return
		}

		entries, err := a.standardMiddlewares.Blackboard.GetAll(r.Context())
		if err != nil {
			logging.GetLogger().Error("failed to get all blackboard keys", "error", err)
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}

		if entries == nil {
			entries = make([]middleware.BlackboardEntry, 0)
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(entries); err != nil {
			logging.GetLogger().Error("failed to encode blackboard entries", "error", err)
		}
	}
}
