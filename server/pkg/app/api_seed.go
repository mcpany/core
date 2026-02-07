// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"encoding/json"
	"io"
	"net/http"
	"os"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/logging"
	"github.com/mcpany/core/server/pkg/util"
	"github.com/mcpany/core/server/pkg/util/passhash"
	"google.golang.org/protobuf/proto"
)

// SeedRequest represents the data to seed.
type SeedRequest struct {
	Users       []*configv1.User       `json:"users"`
	Credentials []*configv1.Credential `json:"credentials"`
	Secrets     []*configv1.Secret     `json:"secrets"`
}

// handleSeed handles the database seeding request.
// POST /api/v1/debug/seed
func (a *Application) handleSeed(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Safety check: Only allow in debug mode or if explicitly enabled
	if os.Getenv("MCPANY_DEBUG") != util.TrueStr {
		logging.GetLogger().Warn("Blocked seed attempt in non-debug mode")
		http.Error(w, "Seeding only allowed in debug mode", http.StatusForbidden)
		return
	}

	ctx := r.Context()
	log := logging.GetLogger()

	// Parse request
	var req SeedRequest
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "failed to read body", http.StatusBadRequest)
		return
	}
	if len(body) > 0 {
		if err := json.Unmarshal(body, &req); err != nil {
			http.Error(w, "invalid json: "+err.Error(), http.StatusBadRequest)
			return
		}
	}

	// Clear existing data
	// 1. Users
	users, err := a.Storage.ListUsers(ctx)
	if err != nil {
		http.Error(w, "failed to list users: "+err.Error(), http.StatusInternalServerError)
		return
	}
	for _, u := range users {
		if err := a.Storage.DeleteUser(ctx, u.GetId()); err != nil {
			log.Error("failed to delete user during seed", "id", u.GetId(), "error", err)
		}
	}

	// 2. Credentials
	creds, err := a.Storage.ListCredentials(ctx)
	if err != nil {
		http.Error(w, "failed to list credentials: "+err.Error(), http.StatusInternalServerError)
		return
	}
	for _, c := range creds {
		if err := a.Storage.DeleteCredential(ctx, c.GetId()); err != nil {
			log.Error("failed to delete credential during seed", "id", c.GetId(), "error", err)
		}
	}

	// 3. Secrets
	secrets, err := a.Storage.ListSecrets(ctx)
	if err != nil {
		http.Error(w, "failed to list secrets: "+err.Error(), http.StatusInternalServerError)
		return
	}
	for _, s := range secrets {
		if err := a.Storage.DeleteSecret(ctx, s.GetId()); err != nil {
			log.Error("failed to delete secret during seed", "id", s.GetId(), "error", err)
		}
	}

	// Seed new data
	// Defaults if empty
	if len(req.Users) == 0 {
		// Default Admin
		hash, _ := passhash.Password("password")
		admin := configv1.User_builder{
			Id:    proto.String("admin"),
			Roles: []string{"admin"},
			Authentication: configv1.Authentication_builder{
				BasicAuth: configv1.BasicAuth_builder{
					Username:     proto.String("admin"),
					PasswordHash: proto.String(hash),
				}.Build(),
			}.Build(),
		}.Build()
		req.Users = append(req.Users, admin)
	}

	for _, u := range req.Users {
		// Ensure password hash if provided plain
		if u.GetAuthentication() != nil && u.GetAuthentication().GetBasicAuth() != nil {
			basicAuth := u.GetAuthentication().GetBasicAuth()
			plain := basicAuth.GetPasswordHash()
			if plain != "" && len(plain) < 60 { // Assume plain if short
				hash, _ := passhash.Password(plain)
				basicAuth.SetPasswordHash(hash)
			}
		}
		if err := a.Storage.CreateUser(ctx, u); err != nil {
			log.Error("failed to create user", "id", u.GetId(), "error", err)
			http.Error(w, "failed to seed user: "+err.Error(), http.StatusInternalServerError)
			return
		}
	}

	for _, c := range req.Credentials {
		if err := a.Storage.SaveCredential(ctx, c); err != nil {
			http.Error(w, "failed to seed credential: "+err.Error(), http.StatusInternalServerError)
			return
		}
	}

	for _, s := range req.Secrets {
		if err := a.Storage.SaveSecret(ctx, s); err != nil {
			http.Error(w, "failed to seed secret: "+err.Error(), http.StatusInternalServerError)
			return
		}
	}

	// Reload config to apply changes (Users update)
	if err := a.ReloadConfig(ctx, a.fs, a.configPaths); err != nil {
		log.Error("failed to reload config after seed", "error", err)
	}

	w.WriteHeader(http.StatusOK)
	writeJSON(w, http.StatusOK, map[string]string{"status": "seeded"})
}
