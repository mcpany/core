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

// SeedData represents the data structure for seeding the database.
type SeedData struct {
	Credentials []json.RawMessage `json:"credentials"`
	Services    []json.RawMessage `json:"services"`
	Secrets     []json.RawMessage `json:"secrets"`
	Profiles    []json.RawMessage `json:"profiles"`
}

// handleDebugSeed handles the seeding of data for testing/debugging purposes.
func (a *Application) handleDebugSeed() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Parse the request body
		var seedData SeedData
		if err := json.NewDecoder(r.Body).Decode(&seedData); err != nil {
			http.Error(w, "Invalid JSON: "+err.Error(), http.StatusBadRequest)
			return
		}

		log := logging.GetLogger()
		store := a.Storage

		if store == nil {
			http.Error(w, "Storage not initialized", http.StatusInternalServerError)
			return
		}

		ctx := r.Context()
		unmarshalOpts := protojson.UnmarshalOptions{DiscardUnknown: true}

		// Seed Credentials
		for _, raw := range seedData.Credentials {
			var cred configv1.Credential
			if err := unmarshalOpts.Unmarshal(raw, &cred); err != nil {
				log.Error("Failed to unmarshal credential", "error", err)
				http.Error(w, "Invalid credential data", http.StatusBadRequest)
				return
			}
			if err := store.SaveCredential(ctx, &cred); err != nil {
				log.Error("Failed to save credential", "id", cred.GetId(), "error", err)
				http.Error(w, "Failed to save credential", http.StatusInternalServerError)
				return
			}
		}

		// Seed Services
		for _, raw := range seedData.Services {
			var svc configv1.UpstreamServiceConfig
			if err := unmarshalOpts.Unmarshal(raw, &svc); err != nil {
				log.Error("Failed to unmarshal service", "error", err)
				http.Error(w, "Invalid service data", http.StatusBadRequest)
				return
			}
			if err := store.SaveService(ctx, &svc); err != nil {
				log.Error("Failed to save service", "name", svc.GetName(), "error", err)
				http.Error(w, "Failed to save service", http.StatusInternalServerError)
				return
			}
		}

		// Seed Secrets
		for _, raw := range seedData.Secrets {
			var secret configv1.Secret
			if err := unmarshalOpts.Unmarshal(raw, &secret); err != nil {
				log.Error("Failed to unmarshal secret", "error", err)
				http.Error(w, "Invalid secret data", http.StatusBadRequest)
				return
			}
			if err := store.SaveSecret(ctx, &secret); err != nil {
				log.Error("Failed to save secret", "id", secret.GetId(), "error", err)
				http.Error(w, "Failed to save secret", http.StatusInternalServerError)
				return
			}
		}

		// Seed Profiles
		for _, raw := range seedData.Profiles {
			var profile configv1.ProfileDefinition
			if err := unmarshalOpts.Unmarshal(raw, &profile); err != nil {
				log.Error("Failed to unmarshal profile", "error", err)
				http.Error(w, "Invalid profile data", http.StatusBadRequest)
				return
			}
			if err := store.SaveProfile(ctx, &profile); err != nil {
				log.Error("Failed to save profile", "name", profile.GetName(), "error", err)
				http.Error(w, "Failed to save profile", http.StatusInternalServerError)
				return
			}
		}

		// Reload configuration to apply changes (especially for services/profiles)
		if err := a.ReloadConfig(ctx, a.fs, a.configPaths); err != nil {
			log.Error("Failed to reload config after seeding", "error", err)
			// We don't fail the request, but we log it.
		}

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status": "ok"}`))
	}
}
