// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"encoding/json"
	"net/http"
	"os"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/logging"
	"google.golang.org/protobuf/encoding/protojson"
)

// handleDebugReset resets the database to a clean state (removes all services, users, etc.).
// Only available if MCPANY_E2E_TEST_MODE is true.
func (a *Application) handleDebugReset() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if os.Getenv("MCPANY_E2E_TEST_MODE") != "true" {
			http.Error(w, "Debug endpoints disabled", http.StatusForbidden)
			return
		}

		ctx := r.Context()
		log := logging.GetLogger()

		// 1. Clear Services
		services, err := a.Storage.ListServices(ctx)
		if err != nil {
			log.Error("Failed to list services for reset", "error", err)
			http.Error(w, "Failed to list services", http.StatusInternalServerError)
			return
		}
		for _, s := range services {
			if err := a.Storage.DeleteService(ctx, s.GetName()); err != nil {
				log.Error("Failed to delete service", "service", s.GetName(), "error", err)
			}
		}

		// 2. Clear Users (except admin maybe? No, full reset means FULL reset usually, but we might lose access)
		// If we clear users, we might kill the current session if it depends on user existence.
		// E2E tests usually re-seed admin user immediately after reset.
		users, err := a.Storage.ListUsers(ctx)
		if err != nil {
			log.Error("Failed to list users for reset", "error", err)
		} else {
			for _, u := range users {
				if err := a.Storage.DeleteUser(ctx, u.GetId()); err != nil {
					log.Error("Failed to delete user", "user", u.GetId(), "error", err)
				}
			}
		}

		// 3. Clear Profiles
		profiles, err := a.Storage.ListProfiles(ctx)
		if err != nil {
			log.Error("Failed to list profiles for reset", "error", err)
		} else {
			for _, p := range profiles {
				if err := a.Storage.DeleteProfile(ctx, p.GetName()); err != nil {
					log.Error("Failed to delete profile", "profile", p.GetName(), "error", err)
				}
			}
		}

		// 4. Clear Collections
		collections, err := a.Storage.ListServiceCollections(ctx)
		if err != nil {
			log.Error("Failed to list collections for reset", "error", err)
		} else {
			for _, c := range collections {
				if err := a.Storage.DeleteServiceCollection(ctx, c.GetName()); err != nil {
					log.Error("Failed to delete collection", "collection", c.GetName(), "error", err)
				}
			}
		}

		// 5. Clear Secrets
		secrets, err := a.Storage.ListSecrets(ctx)
		if err != nil {
			log.Error("Failed to list secrets for reset", "error", err)
		} else {
			for _, s := range secrets {
				if err := a.Storage.DeleteSecret(ctx, s.GetId()); err != nil {
					log.Error("Failed to delete secret", "secret", s.GetId(), "error", err)
				}
			}
		}

		// Re-initialize default admin user to ensure accessibility
		if err := a.initializeAdminUser(ctx, a.Storage); err != nil {
			log.Error("Failed to re-initialize admin user after reset", "error", err)
		}

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status": "reset_complete"}`))
	}
}

// handleDebugSeed seeds the database with provided configuration.
// Only available if MCPANY_E2E_TEST_MODE is true.
func (a *Application) handleDebugSeed() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if os.Getenv("MCPANY_E2E_TEST_MODE") != "true" {
			http.Error(w, "Debug endpoints disabled", http.StatusForbidden)
			return
		}

		var raw map[string]json.RawMessage
		if err := json.NewDecoder(r.Body).Decode(&raw); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		ctx := r.Context()
		log := logging.GetLogger()

		// Seed Services
		if val, ok := raw["services"]; ok {
			// This is tricky with protojson because it expects a single message.
			// We can unmarshal as a list of raw messages first.
			var rawList []json.RawMessage
			if err := json.Unmarshal(val, &rawList); err == nil {
				for _, item := range rawList {
					var svc configv1.UpstreamServiceConfig
					if err := protojson.Unmarshal(item, &svc); err == nil {
						if err := a.Storage.SaveService(ctx, &svc); err != nil {
							log.Error("Failed to seed service", "name", svc.GetName(), "error", err)
						}
					}
				}
			}
		}

		// Seed Users
		if val, ok := raw["users"]; ok {
			var rawList []json.RawMessage
			if err := json.Unmarshal(val, &rawList); err == nil {
				for _, item := range rawList {
					var user configv1.User
					if err := protojson.Unmarshal(item, &user); err == nil {
						if err := a.Storage.CreateUser(ctx, &user); err != nil {
							// Try update if create fails (e.g. ID exists)
							if err := a.Storage.UpdateUser(ctx, &user); err != nil {
								log.Error("Failed to seed user", "id", user.GetId(), "error", err)
							}
						}
					}
				}
			}
		}

		// Seed Profiles
		if val, ok := raw["profiles"]; ok {
			var rawList []json.RawMessage
			if err := json.Unmarshal(val, &rawList); err == nil {
				for _, item := range rawList {
					var p configv1.ProfileDefinition
					if err := protojson.Unmarshal(item, &p); err == nil {
						if err := a.Storage.SaveProfile(ctx, &p); err != nil {
							log.Error("Failed to seed profile", "name", p.GetName(), "error", err)
						}
					}
				}
			}
		}

		// Seed Collections
		if val, ok := raw["collections"]; ok {
			var rawList []json.RawMessage
			if err := json.Unmarshal(val, &rawList); err == nil {
				for _, item := range rawList {
					var c configv1.Collection
					if err := protojson.Unmarshal(item, &c); err == nil {
						if err := a.Storage.SaveServiceCollection(ctx, &c); err != nil {
							log.Error("Failed to seed collection", "name", c.GetName(), "error", err)
						}
					}
				}
			}
		}

		// Seed Secrets
		if val, ok := raw["secrets"]; ok {
			var rawList []json.RawMessage
			if err := json.Unmarshal(val, &rawList); err == nil {
				for _, item := range rawList {
					var s configv1.Secret
					if err := protojson.Unmarshal(item, &s); err == nil {
						if err := a.Storage.SaveSecret(ctx, &s); err != nil {
							log.Error("Failed to seed secret", "id", s.GetId(), "error", err)
						}
					}
				}
			}
		}

		// Seed Global Settings
		if val, ok := raw["global_settings"]; ok {
			var gs configv1.GlobalSettings
			if err := protojson.Unmarshal(val, &gs); err == nil {
				if err := a.Storage.SaveGlobalSettings(ctx, &gs); err != nil {
					log.Error("Failed to seed global settings", "error", err)
				}
			}
		}

		// Trigger Reload to apply changes
		if err := a.ReloadConfig(ctx, a.fs, a.configPaths); err != nil {
			log.Error("Failed to reload config after seeding", "error", err)
			http.Error(w, "Seeding done but reload failed", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status": "seeded"}`))
	}
}
