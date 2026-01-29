// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/auth"
	"github.com/mcpany/core/server/pkg/storage/memory"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

// TestUserSecurity_IDOR tests that users cannot access or modify other users' data.
func TestUserSecurity_IDOR(t *testing.T) {
	app := NewApplication()
	app.fs = afero.NewMemMapFs()
	app.AuthManager = auth.NewManager()
	store := memory.NewStore()
	app.Storage = store

	userDetailHandler := app.handleUserDetail(store)

	// Setup: Create "victim" user
	victim := &configv1.User{Id: proto.String("victim")}
	require.NoError(t, store.CreateUser(context.Background(), victim))

	t.Run("Attacker cannot update Victim", func(t *testing.T) {
		updatedUser := &configv1.User{
			Id: proto.String("victim"),
			Authentication: &configv1.Authentication{
				AuthMethod: &configv1.Authentication_BasicAuth{
					BasicAuth: &configv1.BasicAuth{
						PasswordHash: proto.String("hacked"),
					},
				},
			},
		}
		opts := protojson.MarshalOptions{UseProtoNames: true}
		userBytes, _ := opts.Marshal(updatedUser)
		bodyMap := map[string]json.RawMessage{
			"user": json.RawMessage(userBytes),
		}
		body, _ := json.Marshal(bodyMap)

		req := httptest.NewRequest(http.MethodPut, "/users/victim", bytes.NewReader(body))
		// Authenticate as "attacker"
		ctx := auth.ContextWithUser(req.Context(), "attacker")
		req = req.WithContext(ctx)

		w := httptest.NewRecorder()
		userDetailHandler(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code, "Expected 403 Forbidden for IDOR attempt")
	})

	t.Run("Attacker cannot delete Victim", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, "/users/victim", nil)
		// Authenticate as "attacker"
		ctx := auth.ContextWithUser(req.Context(), "attacker")
		req = req.WithContext(ctx)

		w := httptest.NewRecorder()
		userDetailHandler(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code, "Expected 403 Forbidden for IDOR attempt")
	})

	t.Run("Attacker cannot get Victim details", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/users/victim", nil)
		// Authenticate as "attacker"
		ctx := auth.ContextWithUser(req.Context(), "attacker")
		req = req.WithContext(ctx)

		w := httptest.NewRecorder()
		userDetailHandler(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code, "Expected 403 Forbidden for IDOR attempt")
	})
}

// TestUserSecurity_AccessControl tests that only admins can perform privileged actions.
func TestUserSecurity_AccessControl(t *testing.T) {
	app := NewApplication()
	app.fs = afero.NewMemMapFs()
	app.AuthManager = auth.NewManager()
	store := memory.NewStore()
	app.Storage = store

	usersHandler := app.handleUsers(store)

	t.Run("Regular user cannot list users", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/users", nil)
		// Authenticate as "regular"
		ctx := auth.ContextWithUser(req.Context(), "regular")
		req = req.WithContext(ctx)

		w := httptest.NewRecorder()
		usersHandler(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code, "Expected 403 Forbidden for user listing")
	})

	t.Run("Regular user cannot create user", func(t *testing.T) {
		newUser := &configv1.User{Id: proto.String("newuser")}
		opts := protojson.MarshalOptions{UseProtoNames: true}
		userBytes, _ := opts.Marshal(newUser)
		bodyMap := map[string]json.RawMessage{
			"user": json.RawMessage(userBytes),
		}
		body, _ := json.Marshal(bodyMap)

		req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewReader(body))
		// Authenticate as "regular"
		ctx := auth.ContextWithUser(req.Context(), "regular")
		req = req.WithContext(ctx)

		w := httptest.NewRecorder()
		usersHandler(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code, "Expected 403 Forbidden for user creation")
	})

	t.Run("Admin can list users", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/users", nil)
		// Authenticate as "admin" with "admin" role
		ctx := auth.ContextWithUser(req.Context(), "admin")
		ctx = auth.ContextWithRoles(ctx, []string{"admin"})
		req = req.WithContext(ctx)

		w := httptest.NewRecorder()
		usersHandler(w, req)

		assert.Equal(t, http.StatusOK, w.Code, "Admin should be able to list users")
	})
}
