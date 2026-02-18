// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"encoding/json"
	"net/http"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/logging"
	"google.golang.org/protobuf/encoding/protojson"
)

func (a *Application) handleDebugReset() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		ctx := r.Context()
		log := logging.GetLogger()
		log.Warn("RESETTING DATABASE via debug endpoint")

		// 1. Delete all services
		services, err := a.Storage.ListServices(ctx)
		if err != nil {
			http.Error(w, "Failed to list services: "+err.Error(), http.StatusInternalServerError)
			return
		}
		for _, svc := range services {
			if err := a.Storage.DeleteService(ctx, svc.GetName()); err != nil {
				log.Error("Failed to delete service", "name", svc.GetName(), "error", err)
			}
		}

		// 2. Delete all secrets
		secrets, err := a.Storage.ListSecrets(ctx)
		if err != nil {
			log.Error("Failed to list secrets", "error", err)
		} else {
			for _, s := range secrets {
				if err := a.Storage.DeleteSecret(ctx, s.GetId()); err != nil {
					log.Error("Failed to delete secret", "id", s.GetId(), "error", err)
				}
			}
		}

		// 3. Delete all profiles (except maybe default?)
		profiles, err := a.Storage.ListProfiles(ctx)
		if err != nil {
			log.Error("Failed to list profiles", "error", err)
		} else {
			for _, p := range profiles {
				if err := a.Storage.DeleteProfile(ctx, p.GetName()); err != nil {
					log.Error("Failed to delete profile", "name", p.GetName(), "error", err)
				}
			}
		}

		// 4. Delete all collections
		collections, err := a.Storage.ListServiceCollections(ctx)
		if err != nil {
			log.Error("Failed to list collections", "error", err)
		} else {
			for _, c := range collections {
				if err := a.Storage.DeleteServiceCollection(ctx, c.GetName()); err != nil {
					log.Error("Failed to delete collection", "name", c.GetName(), "error", err)
				}
			}
		}

		// Trigger reload to clear in-memory state
		if err := a.ReloadConfig(ctx, a.fs, a.configPaths); err != nil {
			log.Error("Failed to reload config after reset", "error", err)
		}

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("{}"))
	}
}

func (a *Application) handleDebugSeed() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Expect JSON with list of services, etc.
		// For simplicity, let's just accept UpstreamServiceConfig list for now
		// or a wrapper object.

		body, err := readBodyWithLimit(w, r, 5<<20)
		if err != nil {
			return
		}

		// We use json unmarshal to map structure, but protojson for inner objects?
		// protojson.Unmarshal expects a proto message.
		// So we can unmarshal the whole thing if we define a SeedRequest proto,
		// but we don't have one.
		// Let's iterate over a generic map.

		var raw map[string]any
		if err := json.Unmarshal(body, &raw); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		ctx := r.Context()
		log := logging.GetLogger()

		if rawServices, ok := raw["services"].([]any); ok {
			for _, rs := range rawServices {
				b, _ := json.Marshal(rs)
				var svc configv1.UpstreamServiceConfig
				if err := protojson.Unmarshal(b, &svc); err != nil {
					log.Error("Failed to unmarshal service", "error", err)
					continue
				}
				if err := a.Storage.SaveService(ctx, &svc); err != nil {
					log.Error("Failed to save service", "name", svc.GetName(), "error", err)
				}
			}
		}

		if rawSecrets, ok := raw["secrets"].([]any); ok {
			for _, rs := range rawSecrets {
				b, _ := json.Marshal(rs)
				var secret configv1.Secret
				if err := protojson.Unmarshal(b, &secret); err != nil {
					log.Error("Failed to unmarshal secret", "error", err)
					continue
				}
				if err := a.Storage.SaveSecret(ctx, &secret); err != nil {
					log.Error("Failed to save secret", "id", secret.GetId(), "error", err)
				}
			}
		}

		// Reload
		if err := a.ReloadConfig(ctx, a.fs, a.configPaths); err != nil {
			log.Error("Failed to reload config after seed", "error", err)
		}

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("{}"))
	}
}
