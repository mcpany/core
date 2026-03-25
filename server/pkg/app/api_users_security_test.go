// Copyright 2025 Author(s) of MCP Any
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
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func TestHandleUserDetail_IDOR_Reproduction(t *testing.T) {
	app, store := setupApiTestApp()
	handler := app.handleUserDetail(store)

	victim := configv1.User_builder{Id: proto.String("victim-user"), Roles: []string{"user"}}.Build()
	admin := configv1.User_builder{Id: proto.String("admin-user"), Roles: []string{"admin"}}.Build()

	require.NoError(t, store.CreateUser(context.Background(), victim))
	require.NoError(t, store.CreateUser(context.Background(), admin))

	t.Run("Victim Access Own Profile", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/users/victim-user", nil)
		ctx := auth.ContextWithUser(req.Context(), "victim-user")
		ctx = auth.ContextWithRoles(ctx, []string{"user"})
		req = req.WithContext(ctx)

		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
	})

	t.Run("Victim Access Admin Profile", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/users/admin-user", nil)
		ctx := auth.ContextWithUser(req.Context(), "victim-user")
		ctx = auth.ContextWithRoles(ctx, []string{"user"})
		req = req.WithContext(ctx)

		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("Admin Access Victim Profile", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/users/victim-user", nil)
		ctx := auth.ContextWithUser(req.Context(), "admin-user")
		ctx = auth.ContextWithRoles(ctx, []string{"admin"})
		req = req.WithContext(ctx)

		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})
}

func TestHandleUserDetail_PrivilegeEscalation_Reproduction(t *testing.T) {
	app, store := setupApiTestApp()
	handler := app.handleUserDetail(store)

	victim := configv1.User_builder{Id: proto.String("victim-user"), Roles: []string{"user"}}.Build()
	require.NoError(t, store.CreateUser(context.Background(), victim))

	t.Run("Victim Elevates Own Privileges to Admin", func(t *testing.T) {
		payload := map[string]interface{}{
			"user": map[string]interface{}{
				"id":    "victim-user",
				"roles": []string{"admin"},
			},
		}
		body, err := json.Marshal(payload)
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodPut, "/users/victim-user", bytes.NewReader(body))
		ctx := auth.ContextWithUser(req.Context(), "victim-user")
		ctx = auth.ContextWithRoles(ctx, []string{"user"})
		req = req.WithContext(ctx)

		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		updatedUser, err := store.GetUser(context.Background(), "victim-user")
		require.NoError(t, err)

		assert.NotContains(t, updatedUser.GetRoles(), "admin")
	})

	t.Run("Victim Spoofs JSON payload to update Admin", func(t *testing.T) {
		admin := configv1.User_builder{Id: proto.String("admin-user"), Roles: []string{"admin"}}.Build()
		require.NoError(t, store.CreateUser(context.Background(), admin))

		payload := map[string]interface{}{
			"user": map[string]interface{}{
				"id":    "admin-user",
				"roles": []string{"user"},
			},
		}
		body, err := json.Marshal(payload)
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodPut, "/users/victim-user", bytes.NewReader(body))
		ctx := auth.ContextWithUser(req.Context(), "victim-user")
		ctx = auth.ContextWithRoles(ctx, []string{"user"})
		req = req.WithContext(ctx)

		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		adminUser, err := store.GetUser(context.Background(), "admin-user")
		require.NoError(t, err)

		assert.Contains(t, adminUser.GetRoles(), "admin")
	})

	t.Run("Victim Elevates Own Profile IDs", func(t *testing.T) {
		payload := map[string]interface{}{
			"user": map[string]interface{}{
				"id":          "victim-user",
				"profile_ids": []string{"admin-profile"},
			},
		}
		body, err := json.Marshal(payload)
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodPut, "/users/victim-user", bytes.NewReader(body))
		ctx := auth.ContextWithUser(req.Context(), "victim-user")
		ctx = auth.ContextWithRoles(ctx, []string{"user"})
		req = req.WithContext(ctx)

		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		updatedUser, err := store.GetUser(context.Background(), "victim-user")
		require.NoError(t, err)

		assert.NotContains(t, updatedUser.GetProfileIds(), "admin-profile")
	})
}
