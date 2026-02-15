// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"encoding/json"
	"net/http"
	"os"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/logging"
	"github.com/mcpany/core/server/pkg/storage"
	"github.com/mcpany/core/server/pkg/util"
	"google.golang.org/protobuf/encoding/protojson"
)

// handleDebugSeed handles requests to reset and seed the database.
//
// Summary: Clears all data and seeds new data.
//
// Parameters:
//   - store: storage.Storage. The storage backend.
//
// Returns:
//   - http.HandlerFunc: The handler function.
func (a *Application) handleDebugSeed(store storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if os.Getenv("MCPANY_ENABLE_DEBUG_API") != util.TrueStr {
			http.Error(w, "debug API not enabled", http.StatusForbidden)
			return
		}

		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Use RawMessage to defer unmarshalling of proto messages
		var req struct {
			Services    []json.RawMessage `json:"services"`
			Secrets     []json.RawMessage `json:"secrets"`
			Users       []json.RawMessage `json:"users"`
			Credentials []json.RawMessage `json:"credentials"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}

		ctx := r.Context()

		// 1. Clear existing data
		// Services
		services, err := store.ListServices(ctx)
		if err != nil {
			logging.GetLogger().Error("failed to list services for cleanup", "error", err)
			http.Error(w, "failed to list services", http.StatusInternalServerError)
			return
		}
		for _, svc := range services {
			name := svc.GetName()
			if name == "" {
				name = svc.GetId()
			}
			if name != "" {
				if err := store.DeleteService(ctx, name); err != nil {
					logging.GetLogger().Error("failed to delete service during seed cleanup", "name", name, "error", err)
				}
			}
		}

		// Secrets
		secrets, err := store.ListSecrets(ctx)
		if err != nil {
			logging.GetLogger().Error("failed to list secrets for cleanup", "error", err)
		} else {
			for _, s := range secrets {
				if err := store.DeleteSecret(ctx, s.GetId()); err != nil {
					logging.GetLogger().Error("failed to delete secret during seed cleanup", "id", s.GetId(), "error", err)
				}
			}
		}

		// Users
		users, err := store.ListUsers(ctx)
		if err != nil {
			logging.GetLogger().Error("failed to list users for cleanup", "error", err)
		} else {
			for _, u := range users {
				if err := store.DeleteUser(ctx, u.GetId()); err != nil {
					logging.GetLogger().Error("failed to delete user during seed cleanup", "id", u.GetId(), "error", err)
				}
			}
		}

		// Credentials
		creds, err := store.ListCredentials(ctx)
		if err != nil {
			logging.GetLogger().Error("failed to list credentials for cleanup", "error", err)
		} else {
			for _, c := range creds {
				if err := store.DeleteCredential(ctx, c.GetId()); err != nil {
					logging.GetLogger().Error("failed to delete credential during seed cleanup", "id", c.GetId(), "error", err)
				}
			}
		}

		// 2. Insert new data
		// Services
		for _, raw := range req.Services {
			var svc configv1.UpstreamServiceConfig
			if err := protojson.Unmarshal(raw, &svc); err != nil {
				logging.GetLogger().Error("failed to unmarshal service proto", "error", err)
				http.Error(w, "invalid service proto: "+err.Error(), http.StatusBadRequest)
				return
			}
			if err := store.SaveService(ctx, &svc); err != nil {
				logging.GetLogger().Error("failed to save service during seed", "service", svc.GetName(), "error", err)
				http.Error(w, "failed to save service: "+err.Error(), http.StatusInternalServerError)
				return
			}
		}

		// Credentials
		for _, raw := range req.Credentials {
			var c configv1.Credential
			if err := protojson.Unmarshal(raw, &c); err != nil {
				logging.GetLogger().Error("failed to unmarshal credential proto", "error", err)
				http.Error(w, "invalid credential proto: "+err.Error(), http.StatusBadRequest)
				return
			}
			if err := store.SaveCredential(ctx, &c); err != nil {
				logging.GetLogger().Error("failed to save credential during seed", "credential", c.GetId(), "error", err)
				http.Error(w, "failed to save credential: "+err.Error(), http.StatusInternalServerError)
				return
			}
		}

		// Secrets
		for _, raw := range req.Secrets {
			var s configv1.Secret
			if err := protojson.Unmarshal(raw, &s); err != nil {
				logging.GetLogger().Error("failed to unmarshal secret proto", "error", err)
				http.Error(w, "invalid secret proto: "+err.Error(), http.StatusBadRequest)
				return
			}
			if err := store.SaveSecret(ctx, &s); err != nil {
				logging.GetLogger().Error("failed to save secret during seed", "secret", s.GetId(), "error", err)
				http.Error(w, "failed to save secret: "+err.Error(), http.StatusInternalServerError)
				return
			}
		}

		// Users
		for _, raw := range req.Users {
			var u configv1.User
			if err := protojson.Unmarshal(raw, &u); err != nil {
				logging.GetLogger().Error("failed to unmarshal user proto", "error", err)
				http.Error(w, "invalid user proto: "+err.Error(), http.StatusBadRequest)
				return
			}
			if err := store.CreateUser(ctx, &u); err != nil {
				logging.GetLogger().Error("failed to create user during seed", "user", u.GetId(), "error", err)
				http.Error(w, "failed to create user: "+err.Error(), http.StatusInternalServerError)
				return
			}
		}

		// Trigger reload
		if err := a.ReloadConfig(ctx, a.fs, a.configPaths); err != nil {
			logging.GetLogger().Error("failed to reload config after seed", "error", err)
		}

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("{}"))
	}
}
