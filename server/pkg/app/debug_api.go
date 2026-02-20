// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/config"
	"github.com/mcpany/core/server/pkg/logging"
	"github.com/mcpany/core/server/pkg/storage"
	"google.golang.org/protobuf/encoding/protojson"
)

// handleDebugOAuthAuthorize handles the OAuth authorization step for testing.
// It redirects back to the provided redirect_uri with a fixed code.
func (a *Application) handleDebugOAuthAuthorize() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		redirectURI := r.URL.Query().Get("redirect_uri")
		state := r.URL.Query().Get("state")
		if redirectURI == "" {
			http.Error(w, "redirect_uri required", http.StatusBadRequest)
			return
		}
		// Redirect back immediately with code
		target := fmt.Sprintf("%s?code=debug-code&state=%s", redirectURI, state)
		http.Redirect(w, r, target, http.StatusFound)
	}
}

// handleDebugOAuthToken handles the OAuth token exchange for testing.
// It returns a fixed access token.
func (a *Application) handleDebugOAuthToken() http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"access_token": "debug-token", "token_type": "Bearer", "expires_in": 3600}`))
	}
}

// handleDebugSeed seeds data into the storage.
// It accepts a JSON payload with lists of credentials, services, secrets, and profiles.
func (a *Application) handleDebugSeed() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Read body with limit
		body, err := readBodyWithLimit(w, r, 10*1024*1024) // 10MB limit
		if err != nil {
			return
		}

		// We decode into a map of RawMessages to handle lists of proto messages
		var rawData map[string]json.RawMessage
		if err := json.Unmarshal(body, &rawData); err != nil {
			logging.GetLogger().Error("failed to unmarshal seed payload", "error", err)
			http.Error(w, "invalid json: "+err.Error(), http.StatusBadRequest)
			return
		}

		store := a.Storage // Use application storage directly
		if store == nil {
			http.Error(w, "storage not initialized", http.StatusInternalServerError)
			return
		}

		ctx := r.Context()
		opts := protojson.UnmarshalOptions{DiscardUnknown: true}

		processSeedData(ctx, store, rawData, opts)

		// Reload config to apply changes
		if err := a.ReloadConfig(ctx, a.fs, a.configPaths); err != nil {
			logging.GetLogger().Error("failed to reload config after seed", "error", err)
			http.Error(w, "Failed to reload config: "+err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("{}"))
	}
}

func processSeedData(ctx context.Context, store storage.Storage, rawData map[string]json.RawMessage, opts protojson.UnmarshalOptions) {
	if rawCreds, ok := rawData["credentials"]; ok {
		seedCredentials(ctx, store, rawCreds, opts)
	}

	if rawServices, ok := rawData["services"]; ok {
		seedServices(ctx, store, rawServices, opts)
	}

	if rawSecrets, ok := rawData["secrets"]; ok {
		seedSecrets(ctx, store, rawSecrets, opts)
	}

	if rawProfiles, ok := rawData["profiles"]; ok {
		seedProfiles(ctx, store, rawProfiles, opts)
	}

	if rawUsers, ok := rawData["users"]; ok {
		seedUsers(ctx, store, rawUsers, opts)
	}
}

func seedCredentials(ctx context.Context, store storage.Storage, raw json.RawMessage, opts protojson.UnmarshalOptions) {
	var rawList []json.RawMessage
	if err := json.Unmarshal(raw, &rawList); err != nil {
		logging.GetLogger().Error("failed to unmarshal credentials list", "error", err)
		return
	}
	for _, r := range rawList {
		var cred configv1.Credential
		if err := opts.Unmarshal(r, &cred); err == nil {
			if err := store.SaveCredential(ctx, &cred); err != nil {
				logging.GetLogger().Error("failed to save credential", "id", cred.GetId(), "error", err)
			}
		} else {
			logging.GetLogger().Error("failed to unmarshal credential proto", "error", err)
		}
	}
}

func seedServices(ctx context.Context, store storage.Storage, raw json.RawMessage, opts protojson.UnmarshalOptions) {
	var rawList []json.RawMessage
	if err := json.Unmarshal(raw, &rawList); err != nil {
		logging.GetLogger().Error("failed to unmarshal services list", "error", err)
		return
	}
	for _, r := range rawList {
		var svc configv1.UpstreamServiceConfig
		if err := opts.Unmarshal(r, &svc); err == nil {
			// Validate service config
			if err := config.ValidateOrError(ctx, &svc); err != nil {
				logging.GetLogger().Error("skipping invalid service seed", "name", svc.GetName(), "error", err)
				continue
			}
			if err := store.SaveService(ctx, &svc); err != nil {
				logging.GetLogger().Error("failed to save service", "name", svc.GetName(), "error", err)
			}
		} else {
			logging.GetLogger().Error("failed to unmarshal service proto", "error", err)
		}
	}
}

func seedSecrets(ctx context.Context, store storage.Storage, raw json.RawMessage, opts protojson.UnmarshalOptions) {
	var rawList []json.RawMessage
	if err := json.Unmarshal(raw, &rawList); err != nil {
		logging.GetLogger().Error("failed to unmarshal secrets list", "error", err)
		return
	}
	for _, r := range rawList {
		var secret configv1.Secret
		if err := opts.Unmarshal(r, &secret); err == nil {
			if err := store.SaveSecret(ctx, &secret); err != nil {
				logging.GetLogger().Error("failed to save secret", "id", secret.GetId(), "error", err)
			}
		} else {
			logging.GetLogger().Error("failed to unmarshal secret proto", "error", err)
		}
	}
}

func seedProfiles(ctx context.Context, store storage.Storage, raw json.RawMessage, opts protojson.UnmarshalOptions) {
	var rawList []json.RawMessage
	if err := json.Unmarshal(raw, &rawList); err != nil {
		logging.GetLogger().Error("failed to unmarshal profiles list", "error", err)
		return
	}
	for _, r := range rawList {
		var profile configv1.ProfileDefinition
		if err := opts.Unmarshal(r, &profile); err == nil {
			if err := store.SaveProfile(ctx, &profile); err != nil {
				logging.GetLogger().Error("failed to save profile", "name", profile.GetName(), "error", err)
			}
		} else {
			logging.GetLogger().Error("failed to unmarshal profile proto", "error", err)
		}
	}
}

func seedUsers(ctx context.Context, store storage.Storage, raw json.RawMessage, opts protojson.UnmarshalOptions) {
	var rawList []json.RawMessage
	if err := json.Unmarshal(raw, &rawList); err != nil {
		logging.GetLogger().Error("failed to unmarshal users list", "error", err)
		return
	}
	for _, r := range rawList {
		var user configv1.User
		if err := opts.Unmarshal(r, &user); err == nil {
			// Check if exists to update or create
			existing, _ := store.GetUser(ctx, user.GetId())
			var err error
			if existing != nil {
				err = store.UpdateUser(ctx, &user)
			} else {
				err = store.CreateUser(ctx, &user)
			}
			if err != nil {
				logging.GetLogger().Error("failed to save user", "id", user.GetId(), "error", err)
			}
		} else {
			logging.GetLogger().Error("failed to unmarshal user proto", "error", err)
		}
	}
}
