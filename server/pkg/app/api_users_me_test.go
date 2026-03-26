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
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

func TestHandleUserMe(t *testing.T) {
	app := NewApplication()
	// Minimal setup for handleUserMe
	store := memory.NewStore()
	handler := app.handleUserMe(store)

	t.Run("MethodNotAllowed", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/users/me", nil)
		w := httptest.NewRecorder()
		handler(w, req)
		assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
	})

	t.Run("Unauthorized", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/users/me", nil)
		// No auth context
		w := httptest.NewRecorder()
		handler(w, req)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("SystemAdmin", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/users/me", nil)
		ctx := auth.ContextWithUser(req.Context(), "system-admin")
		req = req.WithContext(ctx)

		w := httptest.NewRecorder()
		handler(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		require.NoError(t, err)
		assert.Equal(t, "system-admin", resp["id"])
		assert.Equal(t, "System Admin", resp["name"])
	})

	t.Run("UserNotFound", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/users/me", nil)
		ctx := auth.ContextWithUser(req.Context(), "non-existent-user")
		req = req.WithContext(ctx)

		w := httptest.NewRecorder()
		handler(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("Success", func(t *testing.T) {
		// Create user in store
		user := configv1.User_builder{
			Id: proto.String("user-1"),
			Authentication: configv1.Authentication_builder{
				BasicAuth: configv1.BasicAuth_builder{
					PasswordHash: proto.String("secret-hash"),
				}.Build(),
			}.Build(),
		}.Build()
		err := store.CreateUser(context.Background(), user)
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodGet, "/users/me", nil)
		ctx := auth.ContextWithUser(req.Context(), "user-1")
		req = req.WithContext(ctx)

		w := httptest.NewRecorder()
		handler(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var u configv1.User
		err = protojson.Unmarshal(w.Body.Bytes(), &u)
		require.NoError(t, err)
		assert.Equal(t, "user-1", u.GetId())

		// Verify sanitization: Authentication field should be cleared
		assert.NotNil(t, u.GetAuthentication())
		assert.NotNil(t, u.GetAuthentication().GetBasicAuth())
		assert.Equal(t, "REDACTED", u.GetAuthentication().GetBasicAuth().GetPasswordHash())
	})
}
