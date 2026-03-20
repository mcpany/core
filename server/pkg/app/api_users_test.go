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

func TestHandleUsers_Api(t *testing.T) {
	app := NewApplication()
	app.fs = afero.NewMemMapFs()
	app.AuthManager = auth.NewManager()
	store := memory.NewStore()
	app.Storage = store
	handler := app.handleUsers(store)

	// Pre-populate store with a user
	user := configv1.User_builder{Id: proto.String("existing-user")}.Build()
	require.NoError(t, store.CreateUser(context.Background(), user))

	tests := []struct {
		name           string
		method         string
		roles          []string
		body           string // valid JSON string
		expectedStatus int
		verifyResponse func(t *testing.T, w *httptest.ResponseRecorder)
	}{
		{
			name:           "Forbidden - No Admin Role",
			method:         http.MethodGet,
			roles:          []string{"user"},
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "GET - Success List",
			method:         http.MethodGet,
			roles:          []string{"admin"},
			expectedStatus: http.StatusOK,
			verifyResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var users []json.RawMessage
				err := json.Unmarshal(w.Body.Bytes(), &users)
				require.NoError(t, err)
				assert.NotEmpty(t, users)
			},
		},
		{
			name:           "POST - Invalid JSON",
			method:         http.MethodPost,
			roles:          []string{"admin"},
			body:           `{ invalid json`,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "POST - Invalid User Proto in Wrapper",
			method:         http.MethodPost,
			roles:          []string{"admin"},
			body:           `{"user": {"id": 123}}`, // id is supposed to be a string
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "POST - Invalid User Proto Direct",
			method:         http.MethodPost,
			roles:          []string{"admin"},
			body:           `{"id": 123}`, // id is supposed to be a string
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "POST - Missing ID Direct",
			method:         http.MethodPost,
			roles:          []string{"admin"},
			body:           `{"authentication": {}}`,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "POST - Conflict Existing User",
			method:         http.MethodPost,
			roles:          []string{"admin"},
			body:           `{"user": {"id": "existing-user"}}`,
			expectedStatus: http.StatusConflict,
		},
		{
			name:           "POST - Success Wrapped User",
			method:         http.MethodPost,
			roles:          []string{"admin"},
			body:           `{"user": {"id": "new-user-1"}}`,
			expectedStatus: http.StatusCreated,
			verifyResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var u configv1.User
				err := protojson.Unmarshal(w.Body.Bytes(), &u)
				require.NoError(t, err)
				assert.Equal(t, "new-user-1", u.GetId())
			},
		},
		{
			name:           "POST - Success Direct User",
			method:         http.MethodPost,
			roles:          []string{"admin"},
			body:           `{"id": "new-user-2"}`,
			expectedStatus: http.StatusCreated,
			verifyResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var u configv1.User
				err := protojson.Unmarshal(w.Body.Bytes(), &u)
				require.NoError(t, err)
				assert.Equal(t, "new-user-2", u.GetId())
			},
		},
		{
			name:           "POST - Password Hashing",
			method:         http.MethodPost,
			roles:          []string{"admin"},
			body:           `{"user": {"id": "new-user-3", "authentication": {"basic_auth": {"password_hash": "plaintext"}}}}`,
			expectedStatus: http.StatusCreated,
			verifyResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				// The returned user is sanitized (redacted)
				var u configv1.User
				err := protojson.Unmarshal(w.Body.Bytes(), &u)
				require.NoError(t, err)
				assert.Equal(t, "new-user-3", u.GetId())
				assert.Equal(t, "REDACTED", u.GetAuthentication().GetBasicAuth().GetPasswordHash())

				// Check real hash in DB
				dbUser, _ := store.GetUser(context.Background(), "new-user-3")
				assert.NotEqual(t, "plaintext", dbUser.GetAuthentication().GetBasicAuth().GetPasswordHash())
				assert.True(t, strings.HasPrefix(dbUser.GetAuthentication().GetBasicAuth().GetPasswordHash(), "$2"))
			},
		},
		{
			name:           "PUT - Method Not Allowed",
			method:         http.MethodPut,
			roles:          []string{"admin"},
			expectedStatus: http.StatusMethodNotAllowed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var req *http.Request
			if tt.body != "" {
				req = httptest.NewRequest(tt.method, "/users", bytes.NewBufferString(tt.body))
			} else {
				req = httptest.NewRequest(tt.method, "/users", nil)
			}
			ctx := auth.ContextWithRoles(req.Context(), tt.roles)
			req = req.WithContext(ctx)

			w := httptest.NewRecorder()
			handler(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code, w.Body.String())
			if tt.verifyResponse != nil {
				tt.verifyResponse(t, w)
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

	// Create a user for test data
	user := configv1.User_builder{
		Id: proto.String("user1"),
		Roles: []string{"developer"},
	}.Build()
	require.NoError(t, store.CreateUser(context.Background(), user))

	tests := []struct {
		name           string
		method         string
		targetID       string
		contextUser    string
		roles          []string
		body           string // valid JSON string
		expectedStatus int
		verifyResponse func(t *testing.T, w *httptest.ResponseRecorder)
	}{
		{
			name:           "Missing ID",
			method:         http.MethodGet,
			targetID:       "",
			contextUser:    "user1",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Unauthorized - No Context User",
			method:         http.MethodGet,
			targetID:       "user1",
			contextUser:    "", // Missing user context
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "Forbidden - Not Same User and Not Admin",
			method:         http.MethodGet,
			targetID:       "user1",
			contextUser:    "user2",
			roles:          []string{"user"},
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "GET - Success (Same User)",
			method:         http.MethodGet,
			targetID:       "user1",
			contextUser:    "user1",
			roles:          []string{"user"},
			expectedStatus: http.StatusOK,
			verifyResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var u configv1.User
				err := protojson.Unmarshal(w.Body.Bytes(), &u)
				require.NoError(t, err)
				assert.Equal(t, "user1", u.GetId())
			},
		},
		{
			name:           "GET - Success (Admin Accessing Other User)",
			method:         http.MethodGet,
			targetID:       "user1",
			contextUser:    "admin-user",
			roles:          []string{"admin"},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "GET - Not Found",
			method:         http.MethodGet,
			targetID:       "unknown",
			contextUser:    "admin",
			roles:          []string{"admin"},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "PUT - Invalid JSON",
			method:         http.MethodPut,
			targetID:       "user1",
			contextUser:    "user1",
			body:           `{ invalid json }`,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "PUT - Invalid Proto in Wrapper",
			method:         http.MethodPut,
			targetID:       "user1",
			contextUser:    "user1",
			body:           `{"user": {"id": 123}}`,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "PUT - Invalid Proto Direct",
			method:         http.MethodPut,
			targetID:       "user1",
			contextUser:    "user1",
			body:           `{"id": 123}`,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "PUT - ID Mismatch",
			method:         http.MethodPut,
			targetID:       "user1",
			contextUser:    "user1",
			body:           `{"id": "user2"}`,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "PUT - IDOR Protection (Non-Admin cannot escalate roles)",
			method:         http.MethodPut,
			targetID:       "user1",
			contextUser:    "user1",
			roles:          []string{"developer"},
			body:           `{"user": {"id": "user1", "roles": ["admin"]}}`, // Trying to escalate
			expectedStatus: http.StatusOK,
			verifyResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				// The returned user should STILL only have "developer"
				var u configv1.User
				err := protojson.Unmarshal(w.Body.Bytes(), &u)
				require.NoError(t, err)
				assert.Equal(t, []string{"developer"}, u.GetRoles())

				// Check DB as well
				dbUser, _ := store.GetUser(context.Background(), "user1")
				assert.Equal(t, []string{"developer"}, dbUser.GetRoles())
			},
		},
		{
			name:           "PUT - Admin can change roles",
			method:         http.MethodPut,
			targetID:       "user1",
			contextUser:    "admin-user",
			roles:          []string{"admin"},
			body:           `{"user": {"id": "user1", "roles": ["developer", "admin"]}}`,
			expectedStatus: http.StatusOK,
			verifyResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var u configv1.User
				err := protojson.Unmarshal(w.Body.Bytes(), &u)
				require.NoError(t, err)
				assert.Equal(t, []string{"developer", "admin"}, u.GetRoles())
			},
		},
		{
			name:           "PUT - Success Wrapped User",
			method:         http.MethodPut,
			targetID:       "user1",
			contextUser:    "user1",
			roles:          []string{"admin"}, // now it has admin
			body:           `{"user": {"id": "user1", "authentication": {"basic_auth": {"password_hash": "newpass"}}}}`,
			expectedStatus: http.StatusOK,
			verifyResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				dbUser, _ := store.GetUser(context.Background(), "user1")
				assert.NotEqual(t, "newpass", dbUser.GetAuthentication().GetBasicAuth().GetPasswordHash())
				assert.True(t, strings.HasPrefix(dbUser.GetAuthentication().GetBasicAuth().GetPasswordHash(), "$2"))
			},
		},
		{
			name:           "PUT - Success Direct User",
			method:         http.MethodPut,
			targetID:       "user1",
			contextUser:    "user1",
			roles:          []string{"admin"},
			body:           `{"id": "user1", "authentication": {"basic_auth": {"password_hash": "newpass2"}}}`,
			expectedStatus: http.StatusOK,
			verifyResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				dbUser, _ := store.GetUser(context.Background(), "user1")
				assert.NotEqual(t, "newpass2", dbUser.GetAuthentication().GetBasicAuth().GetPasswordHash())
				assert.True(t, strings.HasPrefix(dbUser.GetAuthentication().GetBasicAuth().GetPasswordHash(), "$2"))
			},
		},
		{
			name:           "DELETE - Success",
			method:         http.MethodDelete,
			targetID:       "user1",
			contextUser:    "user1",
			roles:          []string{"admin"},
			expectedStatus: http.StatusNoContent,
			verifyResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				dbUser, _ := store.GetUser(context.Background(), "user1")
				assert.Nil(t, dbUser)
			},
		},
		{
			name:           "POST - Method Not Allowed",
			method:         http.MethodPost,
			targetID:       "user1",
			contextUser:    "user1",
			expectedStatus: http.StatusMethodNotAllowed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var req *http.Request
			if tt.body != "" {
				req = httptest.NewRequest(tt.method, "/users/"+tt.targetID, bytes.NewBufferString(tt.body))
			} else {
				req = httptest.NewRequest(tt.method, "/users/"+tt.targetID, nil)
			}

			// Setup Auth Context
			ctx := req.Context()
			if tt.contextUser != "" {
				ctx = auth.ContextWithUser(ctx, tt.contextUser)
			}
			if len(tt.roles) > 0 {
				ctx = auth.ContextWithRoles(ctx, tt.roles)
			}
			req = req.WithContext(ctx)

			w := httptest.NewRecorder()
			handler(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code, w.Body.String())
			if tt.verifyResponse != nil {
				tt.verifyResponse(t, w)
			}
		})
	}
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
