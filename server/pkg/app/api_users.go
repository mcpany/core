// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/auth"
	"github.com/mcpany/core/server/pkg/logging"
	"github.com/mcpany/core/server/pkg/storage"
	"github.com/mcpany/core/server/pkg/util"
	"github.com/mcpany/core/server/pkg/util/passhash"
	"google.golang.org/protobuf/encoding/protojson"
)

func (a *Application) handleUsers(store storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Authorization: Only admins can list or create users
		if !auth.NewRBACEnforcer().HasRoleInContext(r.Context(), "admin") {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}

		switch r.Method {
		case http.MethodGet:
			users, err := store.ListUsers(r.Context())
			if err != nil {
				logging.GetLogger().Error("failed to list users", "error", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			opts := protojson.MarshalOptions{UseProtoNames: true}
			var buf []byte
			buf = append(buf, '[')
			for i, u := range users {
				if i > 0 {
					buf = append(buf, ',')
				}
				b, _ := opts.Marshal(util.SanitizeUser(u))
				buf = append(buf, b...)
			}
			buf = append(buf, ']')
			_, _ = w.Write(buf)

		case http.MethodPost:
			// Limit 1MB
			r.Body = http.MaxBytesReader(w, r.Body, 1048576)
			body, err := io.ReadAll(r.Body)
			if err != nil {
				http.Error(w, "failed to read body", http.StatusBadRequest)
				return
			}

			// We expect { user: {...} } wrapper from client.ts
			var tempMap map[string]json.RawMessage
			if err := json.Unmarshal(body, &tempMap); err != nil {
				http.Error(w, "invalid json", http.StatusBadRequest)
				return
			}

			var user configv1.User
			if userRaw, ok := tempMap["user"]; ok {
				if err := protojson.Unmarshal(userRaw, &user); err != nil {
					http.Error(w, "invalid user proto: "+err.Error(), http.StatusBadRequest)
					return
				}
			} else {
				// Maybe body IS the user?
				if err := protojson.Unmarshal(body, &user); err != nil {
					http.Error(w, "missing user field or invalid body", http.StatusBadRequest)
					return
				}
			}

			if user.GetId() == "" {
				http.Error(w, "id is required", http.StatusBadRequest)
				return
			}

			if err := hashUserPassword(r.Context(), &user, store); err != nil {
				logging.GetLogger().Error("failed to hash password", "error", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}

			if err := store.CreateUser(r.Context(), &user); err != nil {
				logging.GetLogger().Error("failed to create user", "error", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}

			// Reload auth manager
			a.AuthManager.SetUsers([]*configv1.User{&user}) // Wait, this replaces ALL users?
			// We need to reload usage from config. But ListUsers comes from Storage.
			// AuthManager might be using config-based users OR storage-based users.
			// api.go ReloadConfig: a.AuthManager.SetUsers(cfg.GetUsers())
			// LoadServices loads from store too.

			if err := a.ReloadConfig(r.Context(), a.fs, a.configPaths); err != nil {
				logging.GetLogger().Error("failed to reload config after user create", "error", err)
			}

			w.WriteHeader(http.StatusCreated)
			writeJSON(w, http.StatusCreated, util.SanitizeUser(&user))

		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	}
}

func (a *Application) handleUserDetail(store storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := strings.TrimPrefix(r.URL.Path, "/users/")
		if id == "" {
			http.Error(w, "id required", http.StatusBadRequest)
			return
		}

		// Authorization: Users can only access their own profile, unless they are admin
		currentUserID, ok := auth.UserFromContext(r.Context())
		if !ok {
			// This might happen if auth middleware is disabled or bypassed.
			// Fail secure.
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		isAdmin := auth.NewRBACEnforcer().HasRoleInContext(r.Context(), "admin")
		if currentUserID != id && !isAdmin {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}

		switch r.Method {
		case http.MethodGet:
			user, err := store.GetUser(r.Context(), id)
			if err != nil {
				logging.GetLogger().Error("failed to get user", "id", id, "error", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}
			if user == nil {
				http.NotFound(w, r)
				return
			}
			writeJSON(w, http.StatusOK, util.SanitizeUser(user))

		case http.MethodPut:
			// Expect wrapper { user: ... }
			body, err := io.ReadAll(r.Body)
			if err != nil {
				http.Error(w, "failed to read body", http.StatusBadRequest)
				return
			}

			var tempMap map[string]json.RawMessage
			if err := json.Unmarshal(body, &tempMap); err != nil {
				http.Error(w, "invalid json", http.StatusBadRequest)
				return
			}

			var user configv1.User
			if userRaw, ok := tempMap["user"]; ok {
				if err := protojson.Unmarshal(userRaw, &user); err != nil {
					http.Error(w, "invalid user proto: "+err.Error(), http.StatusBadRequest)
					return
				}
			} else {
				// Maybe body IS the user?
				if err := protojson.Unmarshal(body, &user); err != nil {
					http.Error(w, "missing user field or invalid body", http.StatusBadRequest)
					return
				}
			}

			if user.GetId() != "" && user.GetId() != id {
				http.Error(w, "id mismatch", http.StatusBadRequest)
				return
			}
			user.SetId(id)

			if err := hashUserPassword(r.Context(), &user, store); err != nil {
				logging.GetLogger().Error("failed to hash password", "error", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}

			if err := store.UpdateUser(r.Context(), &user); err != nil {
				logging.GetLogger().Error("failed to update user", "error", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}

			if err := a.ReloadConfig(r.Context(), a.fs, a.configPaths); err != nil {
				logging.GetLogger().Error("failed to reload config after user update", "error", err)
			}

			writeJSON(w, http.StatusOK, util.SanitizeUser(&user))

		case http.MethodDelete:
			if err := store.DeleteUser(r.Context(), id); err != nil {
				logging.GetLogger().Error("failed to delete user", "error", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}

			if err := a.ReloadConfig(r.Context(), a.fs, a.configPaths); err != nil {
				logging.GetLogger().Error("failed to reload config after user delete", "error", err)
			}
			w.WriteHeader(http.StatusNoContent)

		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	}
}

// hashUserPassword hashes the user's password if it is provided in plain text.
// It handles the case where the password is "REDACTED" by fetching the existing user and restoring the hash.
func hashUserPassword(ctx context.Context, user *configv1.User, store storage.Storage) error {
	if user.GetAuthentication() != nil && user.GetAuthentication().GetBasicAuth() != nil {
		basicAuth := user.GetAuthentication().GetBasicAuth()
		plain := basicAuth.GetPasswordHash()

		// Case 1: Password is REDACTED (from SanitizeUser). Restore existing hash.
		if plain == util.RedactedString {
			existingUser, err := store.GetUser(ctx, user.GetId())
			if err != nil {
				return err
			}
			if existingUser != nil && existingUser.GetAuthentication() != nil && existingUser.GetAuthentication().GetBasicAuth() != nil {
				existingHash := existingUser.GetAuthentication().GetBasicAuth().GetPasswordHash()
				basicAuth.SetPasswordHash(existingHash)
			} else {
				// If no existing user or auth, clear the REDACTED value to avoid saving it
				basicAuth.SetPasswordHash("")
			}
			return nil
		}

		// Case 2: Password is provided (likely plain text from UI). Hash it.
		// We assume that if it's not REDACTED and not empty, it's a new password.
		// We verify if it is already a bcrypt hash to avoid double hashing, although UI sends plain text.
		// Bcrypt hashes start with $2a$, $2b$, $2y$ and are 60 chars long.
		if plain != "" {
			if len(plain) == 60 && strings.HasPrefix(plain, "$2") {
				// It looks like a hash, keep it.
				return nil
			}

			// Hash it
			hash, err := passhash.Password(plain)
			if err != nil {
				return err
			}
			basicAuth.SetPasswordHash(hash)
		}
	}
	return nil
}
