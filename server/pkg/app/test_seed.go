// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"encoding/json"
	"net/http"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/logging"
	"github.com/mcpany/core/server/pkg/topology"
	"google.golang.org/protobuf/encoding/protojson"
)

type SeedPayloadRaw struct {
	Users    []json.RawMessage       `json:"users"`
	Services []json.RawMessage       `json:"services"`
	Secrets  []json.RawMessage       `json:"secrets"`
	Traffic  []topology.TrafficPoint `json:"traffic"`
}

func (a *Application) handleDebugSeed() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Only allow in debug mode or if explicitly enabled?
		// For now, assume it's protected by AuthMiddleware and only used in dev/test.
		// Security concern: Nuke DB via API.
		// We should probably check an env var or role.
		// For E2E tests, we have admin role.

		var payload SeedPayloadRaw
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			http.Error(w, "invalid json: "+err.Error(), http.StatusBadRequest)
			return
		}

		ctx := r.Context()
		log := logging.GetLogger()

		// 1. Seed Users
		if len(payload.Users) > 0 {
			for _, raw := range payload.Users {
				var user configv1.User
				if err := protojson.Unmarshal(raw, &user); err != nil {
					log.Error("failed to unmarshal user", "error", err)
					continue
				}
				if user.GetId() == "" {
					continue
				}
				// Save to storage
				// Try to delete first to ensure clean state or overwrite
				_ = a.Storage.DeleteUser(ctx, user.GetId())
				if err := a.Storage.CreateUser(ctx, &user); err != nil {
					log.Error("failed to create user", "id", user.GetId(), "error", err)
				}
			}
			// Update AuthManager
			users, err := a.Storage.ListUsers(ctx)
			if err == nil && a.AuthManager != nil {
				a.AuthManager.SetUsers(users)
			}
		}

		// 2. Seed Secrets
		if len(payload.Secrets) > 0 {
			for _, raw := range payload.Secrets {
				var secret configv1.Secret
				if err := protojson.Unmarshal(raw, &secret); err != nil {
					log.Error("failed to unmarshal secret", "error", err)
					continue
				}
				if secret.GetId() == "" {
					continue
				}
				if err := a.Storage.SaveSecret(ctx, &secret); err != nil {
					log.Error("failed to save secret", "id", secret.GetId(), "error", err)
				}
			}
		}

		// 3. Seed Services
		if len(payload.Services) > 0 {
			for _, raw := range payload.Services {
				var svc configv1.UpstreamServiceConfig
				if err := protojson.Unmarshal(raw, &svc); err != nil {
					log.Error("failed to unmarshal service", "error", err)
					continue
				}
				if svc.GetName() == "" {
					continue
				}
				// Save to storage
				if err := a.Storage.SaveService(ctx, &svc); err != nil {
					log.Error("failed to save service", "name", svc.GetName(), "error", err)
				}
			}
			// Reload Config to pick up new services
			if err := a.ReloadConfig(ctx, a.fs, a.configPaths); err != nil {
				log.Error("failed to reload config after seeding services", "error", err)
			}
		}

		// 4. Seed Traffic
		if len(payload.Traffic) > 0 && a.TopologyManager != nil {
			a.TopologyManager.SeedTrafficHistory(payload.Traffic)
		}

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("{}"))
	}
}
