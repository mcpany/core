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

func TestHandleUsers_Table(t *testing.T) {
	app := NewApplication()
	app.fs = afero.NewMemMapFs()
	app.AuthManager = auth.NewManager()
	store := memory.NewStore()
	app.Storage = store
	handler := app.handleUsers(store)

	// Create a user first to test listing and conflicts
	user := configv1.User_builder{Id: proto.String("user1")}.Build()
	require.NoError(t, store.CreateUser(context.Background(), user))

	tests := []struct {
		name           string
		method         string
		roles          []string
		body           map[string]interface{}
		rawBody        string
		expectedStatus int
		verify         func(t *testing.T, w *httptest.ResponseRecorder)
	}{
		{
			name:           "Forbidden - Missing Admin Role",
			method:         http.MethodGet,
			roles:          []string{"user"},
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "GET - Success",
			method:         http.MethodGet,
			roles:          []string{"admin"},
			expectedStatus: http.StatusOK,
			verify: func(t *testing.T, w *httptest.ResponseRecorder) {
				var users []json.RawMessage
				err := json.Unmarshal(w.Body.Bytes(), &users)
				require.NoError(t, err)
				assert.Len(t, users, 1)

				var u configv1.User
				err = protojson.Unmarshal(users[0], &u)
				require.NoError(t, err)
				assert.Equal(t, "user1", u.GetId())
			},
		},
		{
			name:           "POST - Invalid JSON",
			method:         http.MethodPost,
			roles:          []string{"admin"},
			rawBody:        `{ invalid json }`,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "POST - Missing ID",
			method:         http.MethodPost,
			roles:          []string{"admin"},
			body:           map[string]interface{}{"user": map[string]interface{}{}},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "POST - Invalid User Proto",
			method:         http.MethodPost,
			roles:          []string{"admin"},
			body:           map[string]interface{}{"user": map[string]interface{}{"invalid_field": "value"}},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "POST - Conflict (User Exists)",
			method:         http.MethodPost,
			roles:          []string{"admin"},
			body:           map[string]interface{}{"user": map[string]interface{}{"id": "user1"}},
			expectedStatus: http.StatusConflict,
		},
		{
			name:           "POST - Success Wrapper",
			method:         http.MethodPost,
			roles:          []string{"admin"},
			body:           map[string]interface{}{"user": map[string]interface{}{"id": "user2"}},
			expectedStatus: http.StatusCreated,
			verify: func(t *testing.T, w *httptest.ResponseRecorder) {
				u, err := store.GetUser(context.Background(), "user2")
				require.NoError(t, err)
				assert.NotNil(t, u)
			},
		},
		{
			name:           "POST - Success No Wrapper",
			method:         http.MethodPost,
			roles:          []string{"admin"},
			body:           map[string]interface{}{"id": "user3"},
			expectedStatus: http.StatusCreated,
			verify: func(t *testing.T, w *httptest.ResponseRecorder) {
				u, err := store.GetUser(context.Background(), "user3")
				require.NoError(t, err)
				assert.NotNil(t, u)
			},
		},
		{
			name:           "Method Not Allowed",
			method:         http.MethodPut,
			roles:          []string{"admin"},
			expectedStatus: http.StatusMethodNotAllowed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var bodyReader *bytes.Reader
			if tt.rawBody != "" {
				bodyReader = bytes.NewReader([]byte(tt.rawBody))
			} else if tt.body != nil {
				b, _ := json.Marshal(tt.body)
				bodyReader = bytes.NewReader(b)
			} else {
				bodyReader = bytes.NewReader([]byte{})
			}

			req := httptest.NewRequest(tt.method, "/users", bodyReader)
			ctx := auth.ContextWithRoles(req.Context(), tt.roles)
			req = req.WithContext(ctx)

			w := httptest.NewRecorder()
			handler(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.verify != nil {
				tt.verify(t, w)
			}
		})
	}
}

func TestHandleUserDetail(t *testing.T) {
	app := NewApplication()
	app.fs = afero.NewMemMapFs()
	app.AuthManager = auth.NewManager()
	store := memory.NewStore()
	app.Storage = store
	handler := app.handleUserDetail(store)

	// Create a user
	user := configv1.User_builder{
		Id:    proto.String("user1"),
		Roles: []string{"user"},
	}.Build()
	require.NoError(t, store.CreateUser(context.Background(), user))

	updatedUser := configv1.User_builder{
		Id:    proto.String("user1"),
		Roles: []string{"admin"},
		Authentication: configv1.Authentication_builder{
			BasicAuth: configv1.BasicAuth_builder{
				PasswordHash: proto.String("newpass"),
			}.Build(),
		}.Build(),
	}.Build()
	opts := protojson.MarshalOptions{UseProtoNames: true}
	userBytes, _ := opts.Marshal(updatedUser)

	updatedAdminUser := configv1.User_builder{
		Id:    proto.String("user1"),
		Roles: []string{"editor"},
	}.Build()
	adminUserBytes, _ := opts.Marshal(updatedAdminUser)

	tests := []struct {
		name           string
		method         string
		url            string
		userID         string
		roles          []string
		body           map[string]interface{}
		rawBody        string
		expectedStatus int
		verify         func(t *testing.T, w *httptest.ResponseRecorder)
	}{
		{
			name:           "Missing ID in URL",
			method:         http.MethodGet,
			url:            "/users/",
			userID:         "admin",
			roles:          []string{"admin"},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Unauthorized - No User in Context",
			method:         http.MethodGet,
			url:            "/users/user1",
			userID:         "", // simulate missing context
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "Forbidden - Accessing Other User without Admin",
			method:         http.MethodGet,
			url:            "/users/user1",
			userID:         "other_user",
			roles:          []string{"user"},
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "GET - Success (Self Access)",
			method:         http.MethodGet,
			url:            "/users/user1",
			userID:         "user1",
			roles:          []string{"user"},
			expectedStatus: http.StatusOK,
			verify: func(t *testing.T, w *httptest.ResponseRecorder) {
				var u configv1.User
				err := protojson.Unmarshal(w.Body.Bytes(), &u)
				require.NoError(t, err)
				assert.Equal(t, "user1", u.GetId())
			},
		},
		{
			name:           "GET - Not Found",
			method:         http.MethodGet,
			url:            "/users/unknown",
			userID:         "admin",
			roles:          []string{"admin"},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "PUT - Invalid JSON",
			method:         http.MethodPut,
			url:            "/users/user1",
			userID:         "user1",
			rawBody:        `{ invalid json }`,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "PUT - Invalid User Proto",
			method:         http.MethodPut,
			url:            "/users/user1",
			userID:         "user1",
			body:           map[string]interface{}{"user": map[string]interface{}{"invalid_field": "value"}},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "PUT - ID Mismatch",
			method:         http.MethodPut,
			url:            "/users/user1",
			userID:         "user1",
			body:           map[string]interface{}{"user": map[string]interface{}{"id": "mismatch"}},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "PUT - Success Wrapper (Self, Privilege Escalation Prevented)",
			method:         http.MethodPut,
			url:            "/users/user1",
			userID:         "user1",
			roles:          []string{"user"},
			rawBody:        `{"user": ` + string(userBytes) + `}`,
			expectedStatus: http.StatusOK,
			verify: func(t *testing.T, w *httptest.ResponseRecorder) {
				u, err := store.GetUser(context.Background(), "user1")
				require.NoError(t, err)
				assert.Equal(t, []string{"user"}, u.GetRoles(), "Non-admin should not be able to escalate privileges")
				assert.NotEqual(t, "newpass", u.GetAuthentication().GetBasicAuth().GetPasswordHash())
				assert.True(t, strings.HasPrefix(u.GetAuthentication().GetBasicAuth().GetPasswordHash(), "$2"))
			},
		},
		{
			name:           "PUT - Success No Wrapper (Admin modifying user)",
			method:         http.MethodPut,
			url:            "/users/user1",
			userID:         "admin",
			roles:          []string{"admin"},
			rawBody:        string(adminUserBytes),
			expectedStatus: http.StatusOK,
			verify: func(t *testing.T, w *httptest.ResponseRecorder) {
				u, err := store.GetUser(context.Background(), "user1")
				require.NoError(t, err)
				assert.Equal(t, []string{"editor"}, u.GetRoles(), "Admin should be able to change roles")
			},
		},
		{
			name:           "DELETE - Not Found",
			method:         http.MethodDelete,
			url:            "/users/unknown",
			userID:         "admin",
			roles:          []string{"admin"},
			expectedStatus: http.StatusNoContent,
		},
		{
			name:           "DELETE - Success",
			method:         http.MethodDelete,
			url:            "/users/user1",
			userID:         "admin",
			roles:          []string{"admin"},
			expectedStatus: http.StatusNoContent,
			verify: func(t *testing.T, w *httptest.ResponseRecorder) {
				u, err := store.GetUser(context.Background(), "user1")
				require.NoError(t, err)
				assert.Nil(t, u)
			},
		},
		{
			name:           "Method Not Allowed",
			method:         http.MethodPost,
			url:            "/users/user1",
			userID:         "admin",
			roles:          []string{"admin"},
			expectedStatus: http.StatusMethodNotAllowed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var bodyReader *bytes.Reader
			if tt.rawBody != "" {
				bodyReader = bytes.NewReader([]byte(tt.rawBody))
			} else if tt.body != nil {
				b, _ := json.Marshal(tt.body)
				bodyReader = bytes.NewReader(b)
			} else {
				bodyReader = bytes.NewReader([]byte{})
			}

			req := httptest.NewRequest(tt.method, tt.url, bodyReader)
			ctx := req.Context()
			if tt.userID != "" {
				ctx = auth.ContextWithUser(ctx, tt.userID)
			}
			if len(tt.roles) > 0 {
				ctx = auth.ContextWithRoles(ctx, tt.roles)
			}
			req = req.WithContext(ctx)

			w := httptest.NewRecorder()
			handler(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.verify != nil {
				tt.verify(t, w)
			}
		})
	}
}

func TestHashUserPassword(t *testing.T) {
	app := NewApplication()
	_ = app
	store := memory.NewStore()
	ctx := context.Background()

	t.Run("Empty Password", func(t *testing.T) {
		user := configv1.User_builder{
			Id: proto.String("user1"),
			Authentication: configv1.Authentication_builder{
				BasicAuth: configv1.BasicAuth_builder{
					PasswordHash: proto.String(""),
				}.Build(),
			}.Build(),
		}.Build()

		err := hashUserPassword(ctx, user, store, nil)
		require.NoError(t, err)
		assert.Equal(t, "", user.GetAuthentication().GetBasicAuth().GetPasswordHash())
	})

	t.Run("Already Hashed Password", func(t *testing.T) {
		hash := "$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy"
		user := configv1.User_builder{
			Id: proto.String("user1"),
			Authentication: configv1.Authentication_builder{
				BasicAuth: configv1.BasicAuth_builder{
					PasswordHash: proto.String(hash),
				}.Build(),
			}.Build(),
		}.Build()

		err := hashUserPassword(ctx, user, store, nil)
		require.NoError(t, err)
		assert.Equal(t, hash, user.GetAuthentication().GetBasicAuth().GetPasswordHash())
	})

	t.Run("Redacted Missing Existing User", func(t *testing.T) {
		// User is NOT in store
		user := configv1.User_builder{
			Id: proto.String("user-missing"),
			Authentication: configv1.Authentication_builder{
				BasicAuth: configv1.BasicAuth_builder{
					PasswordHash: proto.String("REDACTED"),
				}.Build(),
			}.Build(),
		}.Build()

		err := hashUserPassword(ctx, user, store, nil)
		require.NoError(t, err)
		assert.Equal(t, "", user.GetAuthentication().GetBasicAuth().GetPasswordHash())
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
