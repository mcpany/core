// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"fmt"
	"net/http"
	"os"

	"github.com/mcpany/core/server/pkg/logging"
	"github.com/mcpany/core/server/pkg/storage"
)

func (a *Application) handleDebugReset(store storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Security Check: Only allow reset if explicitly enabled
		if os.Getenv("MCPANY_ENABLE_DEBUG_RESET") != "true" {
			logging.GetLogger().Warn("Blocked database reset attempt: MCPANY_ENABLE_DEBUG_RESET not set")
			http.Error(w, "Debug reset is disabled", http.StatusForbidden)
			return
		}

		ctx := r.Context()
		log := logging.GetLogger()
		log.Warn("RESETTING DATABASE via debug endpoint")

		// 1. Clear Data
		// Services
		services, err := store.ListServices(ctx)
		if err == nil {
			for _, s := range services {
				_ = store.DeleteService(ctx, s.GetName())
			}
		}

		// Users
		users, err := store.ListUsers(ctx)
		if err == nil {
			for _, u := range users {
				_ = store.DeleteUser(ctx, u.GetId())
			}
		}

		// Secrets
		secrets, err := store.ListSecrets(ctx)
		if err == nil {
			for _, s := range secrets {
				_ = store.DeleteSecret(ctx, s.GetId())
			}
		}

		// Profiles
		profiles, err := store.ListProfiles(ctx)
		if err == nil {
			for _, p := range profiles {
				_ = store.DeleteProfile(ctx, p.GetName())
			}
		}

		// Collections
		collections, err := store.ListServiceCollections(ctx)
		if err == nil {
			for _, c := range collections {
				_ = store.DeleteServiceCollection(ctx, c.GetName())
			}
		}

		// Global Settings (Reset to empty/default)
		// We can't "delete" global settings easily as it's a singleton in some stores,
		// but we can overwrite it in the re-init step.

		// 2. Re-initialize
		if err := a.initializeDatabase(ctx, store); err != nil {
			log.Error("Failed to re-initialize database", "error", err)
			http.Error(w, fmt.Sprintf("Failed to re-initialize: %v", err), http.StatusInternalServerError)
			return
		}

		// 3. Reload Config
		if err := a.ReloadConfig(ctx, a.fs, a.configPaths); err != nil {
			log.Error("Failed to reload config after reset", "error", err)
			http.Error(w, fmt.Sprintf("Failed to reload: %v", err), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("Database reset successful"))
	}
}
