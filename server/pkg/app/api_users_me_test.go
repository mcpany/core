// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import (
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

func TestHandleUserMe(t *testing.T) {
	// Setup Environment
	app := NewApplication()
	app.fs = afero.NewMemMapFs()
	app.AuthManager = auth.NewManager()
	store := memory.NewStore()
	app.Storage = store
	handler := app.handleUserMe(store)

	t.Run("Method Not Allowed", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/users/me", nil)
		// Even with valid auth, wrong method should fail
		ctx := auth.ContextWithUser(req.Context(), "user1")
		req = req.WithContext(ctx)

		w := httptest.NewRecorder()
		handler(w, req)

		assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
	})

	t.Run("Unauthorized - No Context", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/users/me", nil)
		// No auth context
		w := httptest.NewRecorder()
		handler(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("System Admin Virtual User", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/users/me", nil)
		ctx := auth.ContextWithUser(req.Context(), "system-admin")
		req = req.WithContext(ctx)

		w := httptest.NewRecorder()
		handler(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

		var resp map[string]any
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		require.NoError(t, err)

		assert.Equal(t, "system-admin", resp["id"])
		assert.Equal(t, "System Admin", resp["name"])
		assert.Equal(t, "admin@localhost", resp["email"])

		roles, ok := resp["roles"].([]any)
		require.True(t, ok)
		assert.Contains(t, roles, "admin")
	})

	t.Run("Regular User - Success", func(t *testing.T) {
		// Create a user in the store
		user := configv1.User_builder{
			Id: proto.String("user-regular"),
			Authentication: configv1.Authentication_builder{
				BasicAuth: configv1.BasicAuth_builder{
					PasswordHash: proto.String("$2a$10$hashedpassword"),
				}.Build(),
			}.Build(),
		}.Build()
		require.NoError(t, store.CreateUser(context.Background(), user))

		req := httptest.NewRequest(http.MethodGet, "/users/me", nil)
		ctx := auth.ContextWithUser(req.Context(), "user-regular")
		req = req.WithContext(ctx)

		w := httptest.NewRecorder()
		handler(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		// Parse response
		var u configv1.User
		err := protojson.Unmarshal(w.Body.Bytes(), &u)
		require.NoError(t, err)

		assert.Equal(t, "user-regular", u.GetId())

		// Verify sanitization (PasswordHash should be empty or redacted, depending on util.SanitizeUser impl)
		// Based on `api_users_test.go` and `util.SanitizeUser` usage, let's assume it removes it or clears it.
		// If `util.SanitizeUser` removes the Authentication block entirely or just the hash.
		// Let's check safely.
		if u.GetAuthentication() != nil && u.GetAuthentication().GetBasicAuth() != nil {
			hash := u.GetAuthentication().GetBasicAuth().GetPasswordHash()
			// It should NOT be the original hash
			assert.NotEqual(t, "$2a$10$hashedpassword", hash)
			// It's likely empty string or specific redacted value
		}
	})

	t.Run("User Not Found In Store", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/users/me", nil)
		// ID exists in context (e.g. from token) but not in DB (deleted)
		ctx := auth.ContextWithUser(req.Context(), "user-deleted")
		req = req.WithContext(ctx)

		w := httptest.NewRecorder()
		handler(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}
