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

func TestHandleUsers(t *testing.T) {
	fs := afero.NewMemMapFs()
	app := NewApplication()
	app.fs = fs
	// We need configPaths to be set for ReloadConfig to work without crashing,
	// although NewMemMapFs is empty, so it will load nothing, which is fine.
	app.configPaths = []string{}
	app.AuthManager = auth.NewManager()

	store := memory.NewStore()
	app.Storage = store

	handler := app.handleUsers(store)

	t.Run("CreateUser", func(t *testing.T) {
		loc := configv1.APIKeyAuth_HEADER
		user := &configv1.User{
			Id: proto.String("user1"),
			Authentication: &configv1.Authentication{
				AuthMethod: &configv1.Authentication_ApiKey{
					ApiKey: &configv1.APIKeyAuth{
						VerificationValue: proto.String("secret-key"),
						ParamName:         proto.String("X-API-Key"),
						In:                &loc,
					},
				},
			},
		}

		userBytes, err := protojson.Marshal(user)
		require.NoError(t, err)

		wrapper := map[string]json.RawMessage{
			"user": json.RawMessage(userBytes),
		}
		body, err := json.Marshal(wrapper)
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewReader(body))
		w := httptest.NewRecorder()

		handler(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)

		// Verify user created in store
		savedUser, err := store.GetUser(context.Background(), "user1")
		require.NoError(t, err)
		assert.Equal(t, "user1", savedUser.GetId())
		assert.Equal(t, "secret-key", savedUser.GetAuthentication().GetApiKey().GetVerificationValue())
	})

	t.Run("ListUsers", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/users", nil)
		w := httptest.NewRecorder()

		handler(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var users []*configv1.User
		err := json.Unmarshal(w.Body.Bytes(), &users)
		require.NoError(t, err)
		require.Len(t, users, 1)
		assert.Equal(t, "user1", users[0].GetId())
	})

	t.Run("CreateUser_InvalidJSON", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewReader([]byte("{invalid-json")))
		w := httptest.NewRecorder()

		handler(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("CreateUser_MissingID", func(t *testing.T) {
		user := &configv1.User{
			// Id: missing
			Authentication: &configv1.Authentication{
				AuthMethod: &configv1.Authentication_ApiKey{
					ApiKey: &configv1.APIKeyAuth{
						VerificationValue: proto.String("secret-key"),
					},
				},
			},
		}
		userBytes, err := protojson.Marshal(user)
		require.NoError(t, err)

		wrapper := map[string]json.RawMessage{
			"user": json.RawMessage(userBytes),
		}
		body, err := json.Marshal(wrapper)
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewReader(body))
		w := httptest.NewRecorder()

		handler(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("MethodNotAllowed", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPut, "/users", nil)
		w := httptest.NewRecorder()

		handler(w, req)

		assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
	})
}

func TestHandleUserDetail(t *testing.T) {
	fs := afero.NewMemMapFs()
	app := NewApplication()
	app.fs = fs
	app.configPaths = []string{}
	app.AuthManager = auth.NewManager()

	store := memory.NewStore()
	app.Storage = store

	handler := app.handleUserDetail(store)

	// Setup: Create a user first
	user := &configv1.User{
		Id: proto.String("user1"),
		Authentication: &configv1.Authentication{
			AuthMethod: &configv1.Authentication_ApiKey{
				ApiKey: &configv1.APIKeyAuth{
					VerificationValue: proto.String("initial-key"),
				},
			},
		},
	}
	err := store.CreateUser(context.Background(), user)
	require.NoError(t, err)

	t.Run("GetUser", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/users/user1", nil)
		w := httptest.NewRecorder()

		handler(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var fetchedUser configv1.User
		err := protojson.Unmarshal(w.Body.Bytes(), &fetchedUser)
		require.NoError(t, err)
		assert.Equal(t, "user1", fetchedUser.GetId())
	})

	t.Run("GetUser_NotFound", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/users/nonexistent", nil)
		w := httptest.NewRecorder()

		handler(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("UpdateUser", func(t *testing.T) {
		updatedUser := &configv1.User{
			Id: proto.String("user1"),
			Authentication: &configv1.Authentication{
				AuthMethod: &configv1.Authentication_ApiKey{
					ApiKey: &configv1.APIKeyAuth{
						VerificationValue: proto.String("updated-key"),
					},
				},
			},
		}
		userBytes, err := protojson.Marshal(updatedUser)
		require.NoError(t, err)

		wrapper := map[string]json.RawMessage{
			"user": json.RawMessage(userBytes),
		}
		body, err := json.Marshal(wrapper)
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodPut, "/users/user1", bytes.NewReader(body))
		w := httptest.NewRecorder()

		handler(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		// Verify update in store
		savedUser, err := store.GetUser(context.Background(), "user1")
		require.NoError(t, err)
		assert.Equal(t, "updated-key", savedUser.GetAuthentication().GetApiKey().GetVerificationValue())
	})

	t.Run("UpdateUser_IDMismatch", func(t *testing.T) {
		updatedUser := &configv1.User{
			Id: proto.String("user2"), // Mismatch
		}
		userBytes, err := protojson.Marshal(updatedUser)
		require.NoError(t, err)

		wrapper := map[string]json.RawMessage{
			"user": json.RawMessage(userBytes),
		}
		body, err := json.Marshal(wrapper)
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodPut, "/users/user1", bytes.NewReader(body))
		w := httptest.NewRecorder()

		handler(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("DeleteUser", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, "/users/user1", nil)
		w := httptest.NewRecorder()

		handler(w, req)

		assert.Equal(t, http.StatusNoContent, w.Code)

		// Verify delete
		savedUser, err := store.GetUser(context.Background(), "user1")
		require.NoError(t, err)
		assert.Nil(t, savedUser)
	})

	t.Run("MethodNotAllowed", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/users/user1", nil)
		w := httptest.NewRecorder()

		handler(w, req)

		assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
	})
}
