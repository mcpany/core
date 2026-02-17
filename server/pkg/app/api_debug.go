// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"fmt"
	"net/http"

	"github.com/mcpany/core/server/pkg/logging"
	"github.com/mcpany/core/server/pkg/storage"
)

// handleDebugReset resets the database to its initial state.
// This is dangerous and should only be enabled in debug mode or for testing.
func (a *Application) handleDebugReset(store storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Security check: Only allow if explicitly enabled or in dev?
		// For E2E tests, we need this.
		// Maybe check for a specific header or just rely on AuthMiddleware which requires API Key/Admin.
		// Since this endpoint is registered under /api/v1 (which has auth), it's protected by Admin access.

		log := logging.GetLogger()
		log.Warn("⚠️  RESETTING DATABASE VIA DEBUG API ⚠️")

		// 1. Delete all services
		services, _ := store.ListServices(r.Context())
		for _, s := range services {
			_ = store.DeleteService(r.Context(), s.GetName())
		}

		// 2. Delete all secrets
		secrets, _ := store.ListSecrets(r.Context())
		for _, s := range secrets {
			_ = store.DeleteSecret(r.Context(), s.GetId())
		}

		// 3. Delete all profiles (except default?)
		profiles, _ := store.ListProfiles(r.Context())
		for _, p := range profiles {
			_ = store.DeleteProfile(r.Context(), p.GetName())
		}

		// 4. Delete all users?
		// If we delete users, we might kill the current session.
		// Maybe we should keep the admin user?
		// For now, let's keep users to avoid breaking the test runner's auth.

		// 5. Re-seed default data
		if err := a.initializeDatabase(r.Context(), store); err != nil {
			log.Error("Failed to re-initialize database", "error", err)
			http.Error(w, "Failed to re-seed: "+err.Error(), http.StatusInternalServerError)
			return
		}

		// 6. Reload config
		if err := a.ReloadConfig(r.Context(), a.fs, a.configPaths); err != nil {
			log.Error("Failed to reload config after reset", "error", err)
		}

		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprintln(w, "Database reset successful")
	}
}
