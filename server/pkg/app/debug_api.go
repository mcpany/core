// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"encoding/json"
	"net/http"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/config"
	"github.com/mcpany/core/server/pkg/logging"
	"github.com/mcpany/core/server/pkg/storage"
	"google.golang.org/protobuf/encoding/protojson"
)

// handleDebugSeed handles the seeding of database data for testing purposes.
func (a *Application) handleDebugSeed(store storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Security Check: Only allow if Debugger is enabled
		if !config.GlobalSettings().IsDebug() {
			http.Error(w, "Forbidden: Debugger not enabled", http.StatusForbidden)
			return
		}

		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		body, err := readBodyWithLimit(w, r, 5*1024*1024) // 5MB limit
		if err != nil {
			return
		}

		// We use map[string]json.RawMessage to handle partial seeding and use protojson for specific types
		var rawMap map[string]json.RawMessage
		if err := json.Unmarshal(body, &rawMap); err != nil {
			http.Error(w, "invalid json: "+err.Error(), http.StatusBadRequest)
			return
		}

		ctx := r.Context()
		logger := logging.GetLogger()

		// 1. Credentials
		if rawCreds, ok := rawMap["credentials"]; ok {
			var list []json.RawMessage
			if err := json.Unmarshal(rawCreds, &list); err == nil {
				for _, item := range list {
					var cred configv1.Credential
					if err := protojson.Unmarshal(item, &cred); err == nil {
						if err := store.SaveCredential(ctx, &cred); err != nil {
							logger.Error("Failed to seed credential", "id", cred.GetId(), "error", err)
						}
					} else {
						logger.Error("Failed to unmarshal credential", "error", err)
					}
				}
			}
		}

		// 2. Services
		if rawServices, ok := rawMap["services"]; ok {
			var list []json.RawMessage
			if err := json.Unmarshal(rawServices, &list); err == nil {
				for _, item := range list {
					var svc configv1.UpstreamServiceConfig
					if err := protojson.Unmarshal(item, &svc); err == nil {
						if err := store.SaveService(ctx, &svc); err != nil {
							logger.Error("Failed to seed service", "name", svc.GetName(), "error", err)
						}
					} else {
						logger.Error("Failed to unmarshal service", "error", err)
					}
				}
			}
		}

		// 3. Secrets
		if rawSecrets, ok := rawMap["secrets"]; ok {
			var list []json.RawMessage
			if err := json.Unmarshal(rawSecrets, &list); err == nil {
				for _, item := range list {
					var secret configv1.Secret
					if err := protojson.Unmarshal(item, &secret); err == nil {
						if err := store.SaveSecret(ctx, &secret); err != nil {
							logger.Error("Failed to seed secret", "id", secret.GetId(), "error", err)
						}
					} else {
						logger.Error("Failed to unmarshal secret", "error", err)
					}
				}
			}
		}

		// 4. Profiles
		if rawProfiles, ok := rawMap["profiles"]; ok {
			var list []json.RawMessage
			if err := json.Unmarshal(rawProfiles, &list); err == nil {
				for _, item := range list {
					var profile configv1.ProfileDefinition
					if err := protojson.Unmarshal(item, &profile); err == nil {
						if err := store.SaveProfile(ctx, &profile); err != nil {
							logger.Error("Failed to seed profile", "name", profile.GetName(), "error", err)
						}
					} else {
						logger.Error("Failed to unmarshal profile", "error", err)
					}
				}
			}
		}

		// 5. Collections
		if rawCollections, ok := rawMap["collections"]; ok {
			var list []json.RawMessage
			if err := json.Unmarshal(rawCollections, &list); err == nil {
				for _, item := range list {
					var col configv1.Collection
					if err := protojson.Unmarshal(item, &col); err == nil {
						if err := store.SaveServiceCollection(ctx, &col); err != nil {
							logger.Error("Failed to seed collection", "name", col.GetName(), "error", err)
						}
					} else {
						logger.Error("Failed to unmarshal collection", "error", err)
					}
				}
			}
		}

		// Reload config to apply changes (especially for services)
		if err := a.ReloadConfig(ctx, a.fs, a.configPaths); err != nil {
			logger.Error("Failed to reload config after seeding", "error", err)
		}

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status": "seeded"}`))
	}
}
