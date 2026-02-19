// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"encoding/json"
	"fmt"
	"net/http"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/config"
	"github.com/mcpany/core/server/pkg/logging"
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
	return func(w http.ResponseWriter, r *http.Request) {
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

		opts := protojson.UnmarshalOptions{DiscardUnknown: true}

		// Credentials
		if rawCreds, ok := rawData["credentials"]; ok {
			var rawList []json.RawMessage
			if err := json.Unmarshal(rawCreds, &rawList); err != nil {
				logging.GetLogger().Error("failed to unmarshal credentials list", "error", err)
			} else {
				for _, raw := range rawList {
					var cred configv1.Credential
					if err := opts.Unmarshal(raw, &cred); err == nil {
						if err := store.SaveCredential(r.Context(), &cred); err != nil {
							logging.GetLogger().Error("failed to save credential", "id", cred.GetId(), "error", err)
						}
					} else {
						logging.GetLogger().Error("failed to unmarshal credential proto", "error", err)
					}
				}
			}
		}

		// Services
		if rawServices, ok := rawData["services"]; ok {
			var rawList []json.RawMessage
			if err := json.Unmarshal(rawServices, &rawList); err != nil {
				logging.GetLogger().Error("failed to unmarshal services list", "error", err)
			} else {
				for _, raw := range rawList {
					var svc configv1.UpstreamServiceConfig
					if err := opts.Unmarshal(raw, &svc); err == nil {
						// Validate service config
						if err := config.ValidateOrError(r.Context(), &svc); err != nil {
							logging.GetLogger().Error("skipping invalid service seed", "name", svc.GetName(), "error", err)
							continue
						}
						if err := store.SaveService(r.Context(), &svc); err != nil {
							logging.GetLogger().Error("failed to save service", "name", svc.GetName(), "error", err)
						}
					} else {
						logging.GetLogger().Error("failed to unmarshal service proto", "error", err)
					}
				}
			}
		}

		// Secrets
		if rawSecrets, ok := rawData["secrets"]; ok {
			var rawList []json.RawMessage
			if err := json.Unmarshal(rawSecrets, &rawList); err != nil {
				logging.GetLogger().Error("failed to unmarshal secrets list", "error", err)
			} else {
				for _, raw := range rawList {
					var secret configv1.Secret
					if err := opts.Unmarshal(raw, &secret); err == nil {
						if err := store.SaveSecret(r.Context(), &secret); err != nil {
							logging.GetLogger().Error("failed to save secret", "id", secret.GetId(), "error", err)
						}
					} else {
						logging.GetLogger().Error("failed to unmarshal secret proto", "error", err)
					}
				}
			}
		}

		// Profiles
		if rawProfiles, ok := rawData["profiles"]; ok {
			var rawList []json.RawMessage
			if err := json.Unmarshal(rawProfiles, &rawList); err != nil {
				logging.GetLogger().Error("failed to unmarshal profiles list", "error", err)
			} else {
				for _, raw := range rawList {
					var profile configv1.ProfileDefinition
					if err := opts.Unmarshal(raw, &profile); err == nil {
						if err := store.SaveProfile(r.Context(), &profile); err != nil {
							logging.GetLogger().Error("failed to save profile", "name", profile.GetName(), "error", err)
						}
					} else {
						logging.GetLogger().Error("failed to unmarshal profile proto", "error", err)
					}
				}
			}
		}

		// Users
		if rawUsers, ok := rawData["users"]; ok {
			var rawList []json.RawMessage
			if err := json.Unmarshal(rawUsers, &rawList); err != nil {
				logging.GetLogger().Error("failed to unmarshal users list", "error", err)
			} else {
				for _, raw := range rawList {
					var user configv1.User
					if err := opts.Unmarshal(raw, &user); err == nil {
						// Check if exists to update or create?
						// Store has CreateUser and UpdateUser.
						// We can try GetUser first or Update and if fails Create?
						// Store.CreateUser fails if exists? Implementation says INSERT ...
						// We should probably check if it exists.
						// But Store.CreateUser usually fails on conflict.
						// For seeding, upsert is better.
						// My sqlite implementation for CreateUser: INSERT ...
						// UpdateUser: UPDATE ...
						// So I should try Update, if err, Create.
						// Or check GetUser.
						existing, _ := store.GetUser(r.Context(), user.GetId())
						var err error
						if existing != nil {
							err = store.UpdateUser(r.Context(), &user)
						} else {
							err = store.CreateUser(r.Context(), &user)
						}
						if err != nil {
							logging.GetLogger().Error("failed to save user", "id", user.GetId(), "error", err)
						}
					} else {
						logging.GetLogger().Error("failed to unmarshal user proto", "error", err)
					}
				}
			}
		}

		// Reload config to apply changes
		if err := a.ReloadConfig(r.Context(), a.fs, a.configPaths); err != nil {
			logging.GetLogger().Error("failed to reload config after seed", "error", err)
			http.Error(w, "Failed to reload config: "+err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("{}"))
	}
}
