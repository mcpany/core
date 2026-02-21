// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/logging"
	"google.golang.org/protobuf/encoding/protojson"
)

// SeedRequest defines the payload for seeding the database.
// We use json.RawMessage to manually unmarshal using protojson, ensuring correct Protobuf handling.
type SeedRequest struct {
	ServicesRaw    []json.RawMessage `json:"upstream_services"`
	CredentialsRaw []json.RawMessage `json:"credentials"`
	SecretsRaw     []json.RawMessage `json:"secrets"`
	ProfilesRaw    []json.RawMessage `json:"profiles"`
	UsersRaw       []json.RawMessage `json:"users"`
}

// handleDebugSeed creates a handler to seed the database with data.
// It clears existing data before inserting new data.
// handleDebugSeed creates a handler to seed the database with data.
// It clears existing data before inserting new data.
func (a *Application) handleDebugSeed() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req SeedRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		ctx := r.Context()
		log := logging.GetLogger()

		if err := a.clearData(ctx, log); err != nil {
			log.Error("Failed to clear data", "error", err)
			http.Error(w, "Failed to clear data: "+err.Error(), http.StatusInternalServerError)
			return
		}

		if err := a.seedData(ctx, req); err != nil {
			log.Error("Failed to seed data", "error", err)
			if err.Error() == "invalid json" {
				http.Error(w, "Invalid JSON in seed data", http.StatusBadRequest)
			} else {
				http.Error(w, "Failed to seed data: "+err.Error(), http.StatusInternalServerError)
			}
			return
		}

		// Trigger reload to update in-memory state (ServiceRegistry, AuthManager, etc.)
		go func() {
			if err := a.ReloadConfig(context.Background(), a.fs, a.configPaths); err != nil {
				log.Error("Failed to reload config after seeding", "error", err)
			}
		}()

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status": "ok"}`))
	}
}

func (a *Application) clearData(ctx context.Context, log *slog.Logger) error {
	// Services
	services, err := a.Storage.ListServices(ctx)
	if err != nil {
		return fmt.Errorf("failed to list services: %w", err)
	}
	for _, s := range services {
		if err := a.Storage.DeleteService(ctx, s.GetName()); err != nil {
			log.Error("Failed to delete service", "name", s.GetName(), "error", err)
		}
	}

	// Credentials
	creds, err := a.Storage.ListCredentials(ctx)
	if err != nil {
		log.Error("Failed to list credentials for clearing", "error", err)
	} else {
		for _, c := range creds {
			if err := a.Storage.DeleteCredential(ctx, c.GetId()); err != nil {
				log.Error("Failed to delete credential", "id", c.GetId(), "error", err)
			}
		}
	}

	// Secrets
	secrets, err := a.Storage.ListSecrets(ctx)
	if err != nil {
		log.Error("Failed to list secrets for clearing", "error", err)
	} else {
		for _, s := range secrets {
			if err := a.Storage.DeleteSecret(ctx, s.GetId()); err != nil {
				log.Error("Failed to delete secret", "id", s.GetId(), "error", err)
			}
		}
	}

	// Profiles
	profiles, err := a.Storage.ListProfiles(ctx)
	if err != nil {
		log.Error("Failed to list profiles for clearing", "error", err)
	} else {
		for _, p := range profiles {
			if err := a.Storage.DeleteProfile(ctx, p.GetName()); err != nil {
				log.Error("Failed to delete profile", "name", p.GetName(), "error", err)
			}
		}
	}

	// Users
	users, err := a.Storage.ListUsers(ctx)
	if err != nil {
		log.Error("Failed to list users for clearing", "error", err)
	} else {
		for _, u := range users {
			if err := a.Storage.DeleteUser(ctx, u.GetId()); err != nil {
				log.Error("Failed to delete user", "id", u.GetId(), "error", err)
			}
		}
	}
	return nil
}

func (a *Application) seedData(ctx context.Context, req SeedRequest) error {
	for _, raw := range req.ServicesRaw {
		s := configv1.UpstreamServiceConfig_builder{}.Build()
		if err := protojson.Unmarshal(raw, s); err != nil {
			return fmt.Errorf("invalid json")
		}
		if err := a.Storage.SaveService(ctx, s); err != nil {
			return fmt.Errorf("failed to save service %s: %w", s.GetName(), err)
		}
	}
	for _, raw := range req.CredentialsRaw {
		c := configv1.Credential_builder{}.Build()
		if err := protojson.Unmarshal(raw, c); err != nil {
			return fmt.Errorf("invalid json")
		}
		if err := a.Storage.SaveCredential(ctx, c); err != nil {
			return fmt.Errorf("failed to save credential %s: %w", c.GetId(), err)
		}
	}
	for _, raw := range req.SecretsRaw {
		s := configv1.Secret_builder{}.Build()
		if err := protojson.Unmarshal(raw, s); err != nil {
			return fmt.Errorf("invalid json")
		}
		if err := a.Storage.SaveSecret(ctx, s); err != nil {
			return fmt.Errorf("failed to save secret %s: %w", s.GetId(), err)
		}
	}
	for _, raw := range req.ProfilesRaw {
		p := configv1.ProfileDefinition_builder{}.Build()
		if err := protojson.Unmarshal(raw, p); err != nil {
			return fmt.Errorf("invalid json")
		}
		if err := a.Storage.SaveProfile(ctx, p); err != nil {
			return fmt.Errorf("failed to save profile %s: %w", p.GetName(), err)
		}
	}
	for _, raw := range req.UsersRaw {
		u := configv1.User_builder{}.Build()
		if err := protojson.Unmarshal(raw, u); err != nil {
			return fmt.Errorf("invalid json")
		}
		if err := a.Storage.CreateUser(ctx, u); err != nil {
			return fmt.Errorf("failed to create user %s: %w", u.GetId(), err)
		}
	}
	return nil
}
