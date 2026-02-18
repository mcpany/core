// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	config_v1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/logging"
)

// DebugSeedRequest defines the payload for the seeding endpoint.
type DebugSeedRequest struct {
	Reset              bool                               `json:"reset"`
	Services           []*config_v1.UpstreamServiceConfig `json:"services"`
	Users              []*config_v1.User                  `json:"users"`
	Secrets            []*config_v1.Secret                `json:"secrets"`
	Credentials        []*config_v1.Credential            `json:"credentials"`
	Profiles           []*config_v1.ProfileDefinition     `json:"profiles"`
	ServiceCollections []*config_v1.Collection            `json:"service_collections"`
}

// handleDebugSeed returns a handler that seeds the database with the provided data.
func (a *Application) handleDebugSeed() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req DebugSeedRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			logging.GetLogger().Error("Failed to decode seed request", "error", err)
			http.Error(w, "Invalid JSON: "+err.Error(), http.StatusBadRequest)
			return
		}

		ctx := r.Context()

		if req.Reset {
			a.resetData(ctx)
		}

		if err := a.seedData(ctx, req); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte(`{"status": "seeded", "message": "Database seeded successfully"}`)); err != nil {
			logging.GetLogger().Error("Failed to write response", "error", err)
		}
	}
}

func (a *Application) resetData(ctx context.Context) {
	log := logging.GetLogger()
	log.Info("Resetting database state...")

	// Clear Services
	if services, err := a.Storage.ListServices(ctx); err == nil {
		for _, s := range services {
			if err := a.Storage.DeleteService(ctx, s.GetName()); err != nil {
				log.Error("Failed to delete service", "name", s.GetName(), "error", err)
			}
			if a.ServiceRegistry != nil {
				_ = a.ServiceRegistry.UnregisterService(ctx, s.GetName())
			}
		}
	}

	// Clear Users
	if users, err := a.Storage.ListUsers(ctx); err == nil {
		for _, u := range users {
			if err := a.Storage.DeleteUser(ctx, u.GetId()); err != nil {
				log.Error("Failed to delete user", "id", u.GetId(), "error", err)
			}
		}
	}

	// Clear Secrets
	if secrets, err := a.Storage.ListSecrets(ctx); err == nil {
		for _, s := range secrets {
			if err := a.Storage.DeleteSecret(ctx, s.GetId()); err != nil {
				log.Error("Failed to delete secret", "id", s.GetId(), "error", err)
			}
		}
	}

	// Clear Credentials
	if creds, err := a.Storage.ListCredentials(ctx); err == nil {
		for _, c := range creds {
			if err := a.Storage.DeleteCredential(ctx, c.GetId()); err != nil {
				log.Error("Failed to delete credential", "id", c.GetId(), "error", err)
			}
		}
	}

	// Clear Profiles
	if profiles, err := a.Storage.ListProfiles(ctx); err == nil {
		for _, p := range profiles {
			if err := a.Storage.DeleteProfile(ctx, p.GetName()); err != nil {
				log.Error("Failed to delete profile", "name", p.GetName(), "error", err)
			}
		}
	}

	// Clear Collections
	if colls, err := a.Storage.ListServiceCollections(ctx); err == nil {
		for _, c := range colls {
			if err := a.Storage.DeleteServiceCollection(ctx, c.GetName()); err != nil {
				log.Error("Failed to delete collection", "name", c.GetName(), "error", err)
			}
		}
	}
}

func (a *Application) seedData(ctx context.Context, req DebugSeedRequest) error {
	log := logging.GetLogger()

	for _, svc := range req.Services {
		if err := a.Storage.SaveService(ctx, svc); err != nil {
			return fmt.Errorf("failed to save service: %w", err)
		}
		if a.ServiceRegistry != nil {
			if _, _, _, err := a.ServiceRegistry.RegisterService(ctx, svc); err != nil {
				log.Error("Failed to register service in memory", "name", svc.GetName(), "error", err)
			}
		}
	}

	for _, user := range req.Users {
		if err := a.Storage.CreateUser(ctx, user); err != nil {
			if updateErr := a.Storage.UpdateUser(ctx, user); updateErr != nil {
				log.Error("Failed to create and update user", "create_err", err, "update_err", updateErr)
				if req.Reset {
					return fmt.Errorf("failed to create user: %w", err)
				}
			}
		}
	}
	if a.AuthManager != nil {
		if users, err := a.Storage.ListUsers(ctx); err == nil {
			a.AuthManager.SetUsers(users)
		}
	}

	for _, s := range req.Secrets {
		if err := a.Storage.SaveSecret(ctx, s); err != nil {
			return fmt.Errorf("failed to save secret: %w", err)
		}
	}

	for _, c := range req.Credentials {
		if err := a.Storage.SaveCredential(ctx, c); err != nil {
			return fmt.Errorf("failed to save credential: %w", err)
		}
	}

	for _, p := range req.Profiles {
		if err := a.Storage.SaveProfile(ctx, p); err != nil {
			return fmt.Errorf("failed to save profile: %w", err)
		}
	}
	if a.ProfileManager != nil {
		if profiles, err := a.Storage.ListProfiles(ctx); err == nil {
			a.ProfileManager.Update(profiles)
		}
	}

	for _, c := range req.ServiceCollections {
		if err := a.Storage.SaveServiceCollection(ctx, c); err != nil {
			return fmt.Errorf("failed to save collection: %w", err)
		}
	}

	return nil
}
