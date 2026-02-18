// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"net/http"
	"os"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/logging"
	"github.com/mcpany/core/server/pkg/storage"
	"github.com/mcpany/core/server/pkg/util"
)

func (a *Application) handleDebugReset(store storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		if os.Getenv("MCPANY_ENABLE_DEBUG_RESET") != util.TrueStr {
			http.Error(w, "Debug reset disabled", http.StatusForbidden)
			return
		}

		ctx := r.Context()
		log := logging.GetLogger()

		// 1. Delete all services
		services, err := store.ListServices(ctx)
		if err == nil {
			for _, s := range services {
				_ = store.DeleteService(ctx, s.GetName())
			}
		}

		// 2. Delete all users
		users, err := store.ListUsers(ctx)
		if err == nil {
			for _, u := range users {
				_ = store.DeleteUser(ctx, u.GetId())
			}
		}

		// 3. Delete all profiles
		profiles, err := store.ListProfiles(ctx)
		if err == nil {
			for _, p := range profiles {
				_ = store.DeleteProfile(ctx, p.GetName())
			}
		}

		// 4. Delete all collections
		cols, err := store.ListServiceCollections(ctx)
		if err == nil {
			for _, c := range cols {
				_ = store.DeleteServiceCollection(ctx, c.GetName())
			}
		}

		// 5. Delete all secrets
		secrets, err := store.ListSecrets(ctx)
		if err == nil {
			for _, s := range secrets {
				_ = store.DeleteSecret(ctx, s.GetId())
			}
		}

		// 6. Delete all credentials
		creds, err := store.ListCredentials(ctx)
		if err == nil {
			for _, c := range creds {
				_ = store.DeleteCredential(ctx, c.GetId())
			}
		}

		// 7. Reset Global Settings to empty (or minimal default)
		emptyGS := configv1.GlobalSettings_builder{}.Build()
		_ = store.SaveGlobalSettings(ctx, emptyGS)

		// 8. Re-seed
		if err := a.performDatabaseSeeding(ctx, store); err != nil {
			log.Error("Failed to re-seed database", "error", err)
			http.Error(w, "Failed to re-seed: "+err.Error(), http.StatusInternalServerError)
			return
		}

		// Reload config to apply changes
		if err := a.ReloadConfig(ctx, a.fs, a.configPaths); err != nil {
			log.Error("Failed to reload after reset", "error", err)
		}

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("Database reset and seeded."))
	}
}
