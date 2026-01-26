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

func TestHandleUsers_List(t *testing.T) {
	app := NewApplication()
	app.fs = afero.NewMemMapFs()
	app.AuthManager = auth.NewManager()
	store := memory.NewStore()
	app.Storage = store
	handler := app.handleUsers(store)

	// Create a user first
	user := &configv1.User{Id: proto.String("user1")}
	require.NoError(t, store.CreateUser(context.Background(), user))

	req := httptest.NewRequest(http.MethodGet, "/users", nil)
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
}

func TestHandleUserDetail(t *testing.T) {
	app := NewApplication()
	app.fs = afero.NewMemMapFs()
	app.AuthManager = auth.NewManager()
	store := memory.NewStore()
	app.Storage = store
	handler := app.handleUserDetail(store)

	// Create a user
	user := &configv1.User{Id: proto.String("user1")}
	require.NoError(t, store.CreateUser(context.Background(), user))

	t.Run("Get User", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/users/user1", nil)
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
		w := httptest.NewRecorder()
		handler(w, req)
		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("Update User", func(t *testing.T) {
		updatedUser := &configv1.User{
			Id:       proto.String("user1"),
			Authentication: &configv1.Authentication{
				AuthMethod: &configv1.Authentication_BasicAuth{
					BasicAuth: &configv1.BasicAuth{
						PasswordHash: proto.String("newpass"),
					},
				},
			},
		}
		// Wrap in { user: ... }
		opts := protojson.MarshalOptions{UseProtoNames: true}
		userBytes, _ := opts.Marshal(updatedUser)
		bodyMap := map[string]json.RawMessage{
			"user": json.RawMessage(userBytes),
		}
		body, _ := json.Marshal(bodyMap)

		req := httptest.NewRequest(http.MethodPut, "/users/user1", bytes.NewReader(body))
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
		w := httptest.NewRecorder()
		handler(w, req)
		assert.Equal(t, http.StatusNoContent, w.Code)

		// Verify deletion
		u, err := store.GetUser(context.Background(), "user1")
		require.NoError(t, err)
		assert.Nil(t, u)
	})
}

func TestHashUserPassword_Redaction(t *testing.T) {
	app := NewApplication()
	// Dummy store for testing
	_ = app

	store := memory.NewStore()

	// 1. Create a user with a real hash
	user := &configv1.User{
		Id: proto.String("user-redact"),
		Authentication: &configv1.Authentication{
			AuthMethod: &configv1.Authentication_BasicAuth{
				BasicAuth: &configv1.BasicAuth{
					PasswordHash: proto.String("real-hash"),
				},
			},
		},
	}
	require.NoError(t, store.CreateUser(context.Background(), user))

	// 2. Simulate an update where password_hash is "[REDACTED]"
	updatedUser := &configv1.User{
		Id: proto.String("user-redact"),
		Authentication: &configv1.Authentication{
			AuthMethod: &configv1.Authentication_BasicAuth{
				BasicAuth: &configv1.BasicAuth{
					PasswordHash: proto.String("REDACTED"),
				},
			},
		},
	}

	// 3. Call hashUserPassword
	err := hashUserPassword(context.Background(), updatedUser, store)
	require.NoError(t, err)

	// 4. Verify that the hash was restored to "real-hash"
	assert.Equal(t, "real-hash", updatedUser.GetAuthentication().GetBasicAuth().GetPasswordHash())
}

func TestCreateUser_ColonInUsername(t *testing.T) {
	app := NewApplication()
	app.fs = afero.NewMemMapFs()
	app.AuthManager = auth.NewManager()
	store := memory.NewStore()
	app.Storage = store
	handler := app.handleUsers(store)

	// Attempt to create user with colon
	user := &configv1.User{Id: proto.String("user:name")}
	opts := protojson.MarshalOptions{UseProtoNames: true}
	userBytes, _ := opts.Marshal(user)
	bodyMap := map[string]json.RawMessage{
		"user": json.RawMessage(userBytes),
	}
	body, _ := json.Marshal(bodyMap)

	req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewReader(body))
	w := httptest.NewRecorder()
	handler(w, req)

	// Expect 400 Bad Request
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestHandleUsers_Errors(t *testing.T) {
	app := NewApplication()
	store := memory.NewStore()
	handler := app.handleUsers(store)

	tests := []struct {
		name       string
		method     string
		body       string
		wantStatus int
	}{
		{
			name:       "Invalid JSON",
			method:     http.MethodPost,
			body:       `{invalid}`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "Missing ID",
			method:     http.MethodPost,
			body:       `{"user": {}}`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "Method Not Allowed",
			method:     http.MethodDelete,
			body:       ``,
			wantStatus: http.StatusMethodNotAllowed,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, "/users", bytes.NewReader([]byte(tc.body)))
			w := httptest.NewRecorder()
			handler(w, req)
			assert.Equal(t, tc.wantStatus, w.Code)
		})
	}
}

func TestHandleUserDetail_Errors(t *testing.T) {
	app := NewApplication()
	store := memory.NewStore()
	handler := app.handleUserDetail(store)

	// Pre-create user
	user := &configv1.User{Id: proto.String("user1")}
	require.NoError(t, store.CreateUser(context.Background(), user))

	tests := []struct {
		name       string
		method     string
		path       string
		body       string
		wantStatus int
	}{
		{
			name:       "Missing ID in Path",
			method:     http.MethodGet,
			path:       "/users/",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "Put Invalid JSON",
			method:     http.MethodPut,
			path:       "/users/user1",
			body:       `{invalid}`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "Put ID Mismatch",
			method:     http.MethodPut,
			path:       "/users/user1",
			body:       `{"user": {"id": "user2"}}`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "Put ID With Colon (Update)",
			method:     http.MethodPut,
			path:       "/users/user1",
			body:       `{"user": {"id": "user:2"}}`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "Put ID With Colon (URL)",
			method:     http.MethodPut,
			path:       "/users/user:2", // This triggers the new validation
			body:       `{"user": {"id": "user:2"}}`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "Method Not Allowed",
			method:     http.MethodPost,
			path:       "/users/user1",
			wantStatus: http.StatusMethodNotAllowed,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, tc.path, bytes.NewReader([]byte(tc.body)))
			w := httptest.NewRecorder()
			handler(w, req)
			assert.Equal(t, tc.wantStatus, w.Code)
		})
	}
}
