// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/auth"
	"github.com/mcpany/core/server/pkg/storage/memory"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func TestHandleUsers_TableDriven(t *testing.T) {
	app := NewApplication()
	app.fs = afero.NewMemMapFs()
	app.AuthManager = auth.NewManager()
	store := memory.NewStore()
	app.Storage = store
	handler := app.handleUsers(store)

	existingUser := &configv1.User{Id: proto.String("existing-user")}
	require.NoError(t, store.CreateUser(context.Background(), existingUser))

	tests := []struct {
		name           string
		method         string
		role           string
		body           []byte
		expectedStatus int
	}{
		{
			name:           "MethodNotAllowed",
			method:         http.MethodPut,
			role:           "admin",
			expectedStatus: http.StatusMethodNotAllowed,
		},
		{
			name:           "Forbidden_NotAdmin_GET",
			method:         http.MethodGet,
			role:           "user",
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "Forbidden_NotAdmin_POST",
			method:         http.MethodPost,
			role:           "user",
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "POST_InvalidJSON",
			method:         http.MethodPost,
			role:           "admin",
			body:           []byte("{invalid json"),
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "POST_MissingID",
			method:         http.MethodPost,
			role:           "admin",
			body:           []byte(`{"user": {}}`),
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "POST_Conflict",
			method:         http.MethodPost,
			role:           "admin",
			body:           []byte(`{"user": {"id": "existing-user"}}`),
			expectedStatus: http.StatusConflict,
		},
		{
			name:           "POST_Success_DirectJSON",
			method:         http.MethodPost,
			role:           "admin",
			body:           []byte(`{"id": "new-user-1"}`),
			expectedStatus: http.StatusCreated,
		},
		{
			name:           "POST_Success_WrappedJSON",
			method:         http.MethodPost,
			role:           "admin",
			body:           []byte(`{"user": {"id": "new-user-2"}}`),
			expectedStatus: http.StatusCreated,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, "/users", bytes.NewReader(tc.body))
			ctx := auth.ContextWithRoles(req.Context(), []string{tc.role})
			req = req.WithContext(ctx)

			w := httptest.NewRecorder()
			handler(w, req)

			assert.Equal(t, tc.expectedStatus, w.Code)
		})
	}
}

func TestHandleUserDetail_TableDriven(t *testing.T) {
	app := NewApplication()
	app.fs = afero.NewMemMapFs()
	app.AuthManager = auth.NewManager()
	store := memory.NewStore()
	app.Storage = store
	handler := app.handleUserDetail(store)

	existingUser := &configv1.User{Id: proto.String("user-test-detail")}
	require.NoError(t, store.CreateUser(context.Background(), existingUser))

	tests := []struct {
		name           string
		method         string
		url            string
		authUser       string
		role           string
		body           []byte
		expectedStatus int
	}{
		{
			name:           "MissingIDInURL",
			method:         http.MethodGet,
			url:            "/users/",
			authUser:       "user-test-detail",
			role:           "admin",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Unauthorized",
			method:         http.MethodGet,
			url:            "/users/user-test-detail",
			authUser:       "", // Missing auth
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "Forbidden",
			method:         http.MethodGet,
			url:            "/users/user-test-detail",
			authUser:       "other-user",
			role:           "user",
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "MethodNotAllowed",
			method:         http.MethodPost,
			url:            "/users/user-test-detail",
			authUser:       "user-test-detail",
			role:           "user",
			expectedStatus: http.StatusMethodNotAllowed,
		},
		{
			name:           "PUT_InvalidJSON",
			method:         http.MethodPut,
			url:            "/users/user-test-detail",
			authUser:       "user-test-detail",
			role:           "user",
			body:           []byte("bad json"),
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "PUT_IDMismatch",
			method:         http.MethodPut,
			url:            "/users/user-test-detail",
			authUser:       "user-test-detail",
			role:           "user",
			body:           []byte(`{"user": {"id": "wrong-id"}}`),
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "PUT_DirectJSON_Success",
			method:         http.MethodPut,
			url:            "/users/user-test-detail",
			authUser:       "user-test-detail",
			role:           "user",
			body:           []byte(`{"id": "user-test-detail", "name": "updated name"}`),
			expectedStatus: http.StatusOK,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, tc.url, bytes.NewReader(tc.body))
			ctx := req.Context()
			if tc.authUser != "" {
				ctx = auth.ContextWithUser(ctx, tc.authUser)
			}
			if tc.role != "" {
				ctx = auth.ContextWithRoles(ctx, []string{tc.role})
			}
			req = req.WithContext(ctx)

			w := httptest.NewRecorder()
			handler(w, req)

			assert.Equal(t, tc.expectedStatus, w.Code)
		})
	}
}
