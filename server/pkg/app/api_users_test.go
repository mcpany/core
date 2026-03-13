// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
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

func TestHandleUsers_Enhanced(t *testing.T) {
	app := NewApplication()
	app.fs = afero.NewMemMapFs()
	app.AuthManager = auth.NewManager()
	store := memory.NewStore()
	app.Storage = store
	handler := app.handleUsers(store)

	// Create a user first
	user := configv1.User_builder{Id: proto.String("user1")}.Build()
	require.NoError(t, store.CreateUser(context.Background(), user))

	t.Run("List Users", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/users", nil)
		// Inject admin role
		ctx := auth.ContextWithRoles(req.Context(), []string{"admin"})
		req = req.WithContext(ctx)

		w := httptest.NewRecorder()
		handler(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var users []json.RawMessage
		err := json.Unmarshal(w.Body.Bytes(), &users)
		require.NoError(t, err)
		assert.Len(t, users, 1)

		var u configv1.User
		err = protojson.Unmarshal(users[0], &u)
		require.NoError(t, err)
		assert.Equal(t, "user1", u.GetId())
	})

	t.Run("Create User", func(t *testing.T) {
		newUser := configv1.User_builder{
			Id: proto.String("user2"),
			Authentication: configv1.Authentication_builder{
				BasicAuth: configv1.BasicAuth_builder{
					PasswordHash: proto.String("pass123"),
				}.Build(),
			}.Build(),
		}.Build()
		opts := protojson.MarshalOptions{UseProtoNames: true}
		userBytes, _ := opts.Marshal(newUser)
		bodyMap := map[string]json.RawMessage{
			"user": json.RawMessage(userBytes),
		}
		body, _ := json.Marshal(bodyMap)

		req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewReader(body))
		ctx := auth.ContextWithRoles(req.Context(), []string{"admin"})
		req = req.WithContext(ctx)

		w := httptest.NewRecorder()
		handler(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		u, err := store.GetUser(context.Background(), "user2")
		require.NoError(t, err)
		assert.NotNil(t, u)
		assert.True(t, strings.HasPrefix(u.GetAuthentication().GetBasicAuth().GetPasswordHash(), "$2"))
	})

	t.Run("Create User Conflict", func(t *testing.T) {
		conflictUser := configv1.User_builder{
			Id: proto.String("user1"), // user1 already exists
		}.Build()
		opts := protojson.MarshalOptions{UseProtoNames: true}
		userBytes, _ := opts.Marshal(conflictUser)
		bodyMap := map[string]json.RawMessage{
			"user": json.RawMessage(userBytes),
		}
		body, _ := json.Marshal(bodyMap)

		req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewReader(body))
		ctx := auth.ContextWithRoles(req.Context(), []string{"admin"})
		req = req.WithContext(ctx)

		w := httptest.NewRecorder()
		handler(w, req)

		assert.Equal(t, http.StatusConflict, w.Code)
	})

	t.Run("Create User Missing ID", func(t *testing.T) {
		missingIDUser := configv1.User_builder{
			Roles: []string{"admin"}, // No ID
		}.Build()
		opts := protojson.MarshalOptions{UseProtoNames: true}
		userBytes, _ := opts.Marshal(missingIDUser)
		bodyMap := map[string]json.RawMessage{
			"user": json.RawMessage(userBytes),
		}
		body, _ := json.Marshal(bodyMap)

		req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewReader(body))
		ctx := auth.ContextWithRoles(req.Context(), []string{"admin"})
		req = req.WithContext(ctx)

		w := httptest.NewRecorder()
		handler(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Create User Invalid JSON", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewReader([]byte("{invalid}")))
		ctx := auth.ContextWithRoles(req.Context(), []string{"admin"})
		req = req.WithContext(ctx)

		w := httptest.NewRecorder()
		handler(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Create User Direct Proto JSON", func(t *testing.T) {
		newUser := configv1.User_builder{
			Id: proto.String("user3"),
		}.Build()
		opts := protojson.MarshalOptions{UseProtoNames: true}
		userBytes, _ := opts.Marshal(newUser)

		req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewReader(userBytes))
		ctx := auth.ContextWithRoles(req.Context(), []string{"admin"})
		req = req.WithContext(ctx)

		w := httptest.NewRecorder()
		handler(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
	})

	t.Run("Create User Missing Required Array Field (Invalid Proto JSON)", func(t *testing.T) {
		// e.g. a string where an array or object is expected, protojson will fail
		body := []byte(`{"user": {"roles": "not_an_array"}}`)
		req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewReader(body))
		ctx := auth.ContextWithRoles(req.Context(), []string{"admin"})
		req = req.WithContext(ctx)

		w := httptest.NewRecorder()
		handler(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Unauthorized Access", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/users", nil)
		// No admin role injected
		w := httptest.NewRecorder()
		handler(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("Method Not Allowed", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPut, "/users", nil)
		ctx := auth.ContextWithRoles(req.Context(), []string{"admin"})
		req = req.WithContext(ctx)

		w := httptest.NewRecorder()
		handler(w, req)

		assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
	})
}

func TestHandleUserDetail(t *testing.T) {
	app := NewApplication()
	app.fs = afero.NewMemMapFs()
	app.AuthManager = auth.NewManager()
	store := memory.NewStore()
	app.Storage = store
	handler := app.handleUserDetail(store)

	// Create a user
	user := configv1.User_builder{Id: proto.String("user1")}.Build()
	require.NoError(t, store.CreateUser(context.Background(), user))

	t.Run("Get User", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/users/user1", nil)
		// Inject auth context (user accessing self)
		ctx := auth.ContextWithUser(req.Context(), "user1")
		req = req.WithContext(ctx)

		w := httptest.NewRecorder()
		handler(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var u configv1.User
		err := protojson.Unmarshal(w.Body.Bytes(), &u)
		require.NoError(t, err)
		assert.Equal(t, "user1", u.GetId())
	})

	t.Run("Get Non-Existent User", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/users/unknown", nil)
		// Inject admin role to bypass "own user" check and hit 404
		ctx := auth.ContextWithUser(req.Context(), "admin")
		ctx = auth.ContextWithRoles(ctx, []string{"admin"})
		req = req.WithContext(ctx)

		w := httptest.NewRecorder()
		handler(w, req)
		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("Update User", func(t *testing.T) {
		updatedUser := configv1.User_builder{
			Id: proto.String("user1"),
			Authentication: configv1.Authentication_builder{
				BasicAuth: configv1.BasicAuth_builder{
					PasswordHash: proto.String("newpass"),
				}.Build(),
			}.Build(),
		}.Build()
		// Wrap in { user: ... }
		opts := protojson.MarshalOptions{UseProtoNames: true}
		userBytes, _ := opts.Marshal(updatedUser)
		bodyMap := map[string]json.RawMessage{
			"user": json.RawMessage(userBytes),
		}
		body, _ := json.Marshal(bodyMap)

		req := httptest.NewRequest(http.MethodPut, "/users/user1", bytes.NewReader(body))
		// Inject auth context (user accessing self)
		ctx := auth.ContextWithUser(req.Context(), "user1")
		req = req.WithContext(ctx)

		w := httptest.NewRecorder()
		handler(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		// Verify update in store
		u, err := store.GetUser(context.Background(), "user1")
		require.NoError(t, err)
		// Password should be hashed (not "newpass")
		assert.NotEqual(t, "newpass", u.GetAuthentication().GetBasicAuth().GetPasswordHash())
		assert.True(t, strings.HasPrefix(u.GetAuthentication().GetBasicAuth().GetPasswordHash(), "$2"))
	})

	t.Run("Delete User", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, "/users/user1", nil)
		// Inject auth context (user deleting self)
		ctx := auth.ContextWithUser(req.Context(), "user1")
		req = req.WithContext(ctx)

		w := httptest.NewRecorder()
		handler(w, req)
		assert.Equal(t, http.StatusNoContent, w.Code)

		// Verify deletion
		u, err := store.GetUser(context.Background(), "user1")
		require.NoError(t, err)
		assert.Nil(t, u)
	})

	// Missing coverage paths

	t.Run("Missing ID", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/users/", nil)
		ctx := auth.ContextWithUser(req.Context(), "user1")
		req = req.WithContext(ctx)

		w := httptest.NewRecorder()
		handler(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Unauthenticated", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/users/user1", nil)
		w := httptest.NewRecorder()
		handler(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("Forbidden Access", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/users/user1", nil)
		// Accessing someone else without admin
		ctx := auth.ContextWithUser(req.Context(), "user2")
		req = req.WithContext(ctx)

		w := httptest.NewRecorder()
		handler(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("Update User Invalid JSON", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPut, "/users/user2", bytes.NewReader([]byte("{invalid}")))
		ctx := auth.ContextWithUser(req.Context(), "user2")
		req = req.WithContext(ctx)

		w := httptest.NewRecorder()
		handler(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Update User ID Mismatch", func(t *testing.T) {
		updatedUser := configv1.User_builder{
			Id: proto.String("user3"),
		}.Build()
		opts := protojson.MarshalOptions{UseProtoNames: true}
		userBytes, _ := opts.Marshal(updatedUser)

		req := httptest.NewRequest(http.MethodPut, "/users/user2", bytes.NewReader(userBytes))
		ctx := auth.ContextWithUser(req.Context(), "user2")
		req = req.WithContext(ctx)

		w := httptest.NewRecorder()
		handler(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Update Non-existent User", func(t *testing.T) {
		updatedUser := configv1.User_builder{
			Id: proto.String("unknown"),
		}.Build()
		opts := protojson.MarshalOptions{UseProtoNames: true}
		userBytes, _ := opts.Marshal(updatedUser)

		req := httptest.NewRequest(http.MethodPut, "/users/unknown", bytes.NewReader(userBytes))
		ctx := auth.ContextWithUser(req.Context(), "unknown")
		req = req.WithContext(ctx)

		w := httptest.NewRecorder()
		handler(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("Update User Prevent Escalation", func(t *testing.T) {
		// Create a user2 for this test
		u2 := configv1.User_builder{
			Id: proto.String("user2"),
			Roles: []string{"user"},
		}.Build()
		require.NoError(t, store.CreateUser(context.Background(), u2))

		// Try to escalate to admin
		updatedUser := configv1.User_builder{
			Id: proto.String("user2"),
			Roles: []string{"admin"},
		}.Build()
		opts := protojson.MarshalOptions{UseProtoNames: true}
		userBytes, _ := opts.Marshal(updatedUser)

		req := httptest.NewRequest(http.MethodPut, "/users/user2", bytes.NewReader(userBytes))
		ctx := auth.ContextWithUser(req.Context(), "user2") // non-admin
		req = req.WithContext(ctx)

		w := httptest.NewRecorder()
		handler(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		// Verify role was not escalated
		u, err := store.GetUser(context.Background(), "user2")
		require.NoError(t, err)
		assert.Equal(t, []string{"user"}, u.GetRoles())
	})

	t.Run("Method Not Allowed", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/users/user2", nil)
		ctx := auth.ContextWithUser(req.Context(), "user2")
		req = req.WithContext(ctx)

		w := httptest.NewRecorder()
		handler(w, req)

		assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
	})
}

func TestHashUserPassword_Redaction(t *testing.T) {
	app := NewApplication()
	// Dummy store for testing
	_ = app

	store := memory.NewStore()

	// 1. Create a user with a real hash
	user := configv1.User_builder{
		Id: proto.String("user-redact"),
		Authentication: configv1.Authentication_builder{
			BasicAuth: configv1.BasicAuth_builder{
				PasswordHash: proto.String("real-hash"),
			}.Build(),
		}.Build(),
	}.Build()
	require.NoError(t, store.CreateUser(context.Background(), user))

	// 2. Simulate an update where password_hash is "[REDACTED]"
	updatedUser := configv1.User_builder{
		Id: proto.String("user-redact"),
		Authentication: configv1.Authentication_builder{
			BasicAuth: configv1.BasicAuth_builder{
				PasswordHash: proto.String("REDACTED"),
			}.Build(),
		}.Build(),
	}.Build()

	// 3. Call hashUserPassword
	err := hashUserPassword(context.Background(), updatedUser, store, nil)
	require.NoError(t, err)

	// 4. Verify that the hash was restored to "real-hash"
	assert.Equal(t, "real-hash", updatedUser.GetAuthentication().GetBasicAuth().GetPasswordHash())
}
