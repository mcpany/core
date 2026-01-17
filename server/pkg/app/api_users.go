// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/logging"
	"github.com/mcpany/core/server/pkg/storage"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

func (a *Application) handleUsers(store storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
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
				b, _ := opts.Marshal(u)
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
					logging.GetLogger().Warn("invalid user proto", "error", err)
					http.Error(w, "invalid user proto", http.StatusBadRequest)
					return
				}
			} else {
				// Maybe body IS the user?
				if err := protojson.Unmarshal(body, &user); err != nil {
					logging.GetLogger().Warn("invalid user body", "error", err)
					http.Error(w, "missing user field or invalid body", http.StatusBadRequest)
					return
				}
			}

			if user.GetId() == "" {
				http.Error(w, "id is required", http.StatusBadRequest)
				return
			}

			if err := store.CreateUser(r.Context(), &user); err != nil {
				logging.GetLogger().Error("failed to create user", "error", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}

			if err := a.ReloadConfig(r.Context(), a.fs, a.configPaths); err != nil {
				logging.GetLogger().Error("failed to reload config after user create", "error", err)
			}

			w.WriteHeader(http.StatusCreated)
			writeJSON(w, http.StatusCreated, &user)

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
			writeJSON(w, http.StatusOK, user)

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
					logging.GetLogger().Warn("invalid user proto", "error", err)
					http.Error(w, "invalid user proto", http.StatusBadRequest)
					return
				}
			} else {
				// Maybe body IS the user?
				if err := protojson.Unmarshal(body, &user); err != nil {
					logging.GetLogger().Warn("invalid user body", "error", err)
					http.Error(w, "missing user field or invalid body", http.StatusBadRequest)
					return
				}
			}

			if user.GetId() != "" && user.GetId() != id {
				http.Error(w, "id mismatch", http.StatusBadRequest)
				return
			}
			user.Id = proto.String(id)

			if err := store.UpdateUser(r.Context(), &user); err != nil {
				logging.GetLogger().Error("failed to update user", "error", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}

			if err := a.ReloadConfig(r.Context(), a.fs, a.configPaths); err != nil {
				logging.GetLogger().Error("failed to reload config after user update", "error", err)
			}

			writeJSON(w, http.StatusOK, &user)

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
