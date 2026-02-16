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
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func TestHandleGetUserPreferences(t *testing.T) {
	mockStore := new(MockStore)
	authManager := auth.NewManager()
	app := &Application{
		Storage:     mockStore,
		AuthManager: authManager,
	}

	t.Run("User found in storage", func(t *testing.T) {
		ctx := auth.ContextWithUser(context.Background(), "test-user")
		req := httptest.NewRequest("GET", "/preferences", nil).WithContext(ctx)
		w := httptest.NewRecorder()

		user := configv1.User_builder{
			Id: proto.String("test-user"),
			Preferences: map[string]string{
				"theme": "dark",
			},
		}.Build()

		mockStore.On("GetUser", mock.Anything, "test-user").Return(user, nil).Once()

		app.HandleGetUserPreferences(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var prefs map[string]string
		err := json.Unmarshal(w.Body.Bytes(), &prefs)
		require.NoError(t, err)
		assert.Equal(t, "dark", prefs["theme"])
	})

	t.Run("System Admin not found in storage (error)", func(t *testing.T) {
		ctx := auth.ContextWithUser(context.Background(), "system-admin")
		req := httptest.NewRequest("GET", "/preferences", nil).WithContext(ctx)
		w := httptest.NewRecorder()

		mockStore.On("GetUser", mock.Anything, "system-admin").Return((*configv1.User)(nil), assert.AnError).Once()

		app.HandleGetUserPreferences(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var prefs map[string]string
		err := json.Unmarshal(w.Body.Bytes(), &prefs)
		require.NoError(t, err)
		assert.Empty(t, prefs)
	})

	t.Run("System Admin not found in storage (nil, nil)", func(t *testing.T) {
		ctx := auth.ContextWithUser(context.Background(), "system-admin")
		req := httptest.NewRequest("GET", "/preferences", nil).WithContext(ctx)
		w := httptest.NewRecorder()

		// Simulating SQLite store behavior: return nil, nil
		mockStore.On("GetUser", mock.Anything, "system-admin").Return((*configv1.User)(nil), nil).Once()

		app.HandleGetUserPreferences(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var prefs map[string]string
		err := json.Unmarshal(w.Body.Bytes(), &prefs)
		require.NoError(t, err)
		assert.Empty(t, prefs)
	})

	t.Run("User not found in storage but in AuthManager", func(t *testing.T) {
		ctx := auth.ContextWithUser(context.Background(), "file-user")
		req := httptest.NewRequest("GET", "/preferences", nil).WithContext(ctx)
		w := httptest.NewRecorder()

		mockStore.On("GetUser", mock.Anything, "file-user").Return((*configv1.User)(nil), assert.AnError).Once()

		fileUser := configv1.User_builder{
			Id: proto.String("file-user"),
			Preferences: map[string]string{
				"layout": "default",
			},
		}.Build()
		authManager.SetUsers([]*configv1.User{fileUser})

		app.HandleGetUserPreferences(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var prefs map[string]string
		err := json.Unmarshal(w.Body.Bytes(), &prefs)
		require.NoError(t, err)
		assert.Equal(t, "default", prefs["layout"])
	})

	t.Run("User not found anywhere", func(t *testing.T) {
		ctx := auth.ContextWithUser(context.Background(), "unknown-user")
		req := httptest.NewRequest("GET", "/preferences", nil).WithContext(ctx)
		w := httptest.NewRecorder()

		mockStore.On("GetUser", mock.Anything, "unknown-user").Return((*configv1.User)(nil), assert.AnError).Once()

		app.HandleGetUserPreferences(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("User not found anywhere (nil, nil)", func(t *testing.T) {
		ctx := auth.ContextWithUser(context.Background(), "unknown-user")
		req := httptest.NewRequest("GET", "/preferences", nil).WithContext(ctx)
		w := httptest.NewRecorder()

		mockStore.On("GetUser", mock.Anything, "unknown-user").Return((*configv1.User)(nil), nil).Once()

		app.HandleGetUserPreferences(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}

func TestHandleUpdateUserPreferences(t *testing.T) {
	mockStore := new(MockStore)
	authManager := auth.NewManager()
	app := &Application{
		Storage:     mockStore,
		AuthManager: authManager,
	}

	t.Run("Update existing user", func(t *testing.T) {
		ctx := auth.ContextWithUser(context.Background(), "test-user")
		body := bytes.NewBufferString(`{"theme": "light"}`)
		req := httptest.NewRequest("POST", "/preferences", body).WithContext(ctx)
		w := httptest.NewRecorder()

		existingUser := configv1.User_builder{
			Id: proto.String("test-user"),
			Preferences: map[string]string{
				"theme": "dark",
				"lang":  "en",
			},
		}.Build()

		mockStore.On("GetUser", mock.Anything, "test-user").Return(existingUser, nil).Once()

		// Expect UpdateUser with merged prefs
		mockStore.On("UpdateUser", mock.Anything, mock.MatchedBy(func(u *configv1.User) bool {
			prefs := u.GetPreferences()
			return prefs["theme"] == "light" && prefs["lang"] == "en"
		})).Return(nil).Once()

		app.HandleUpdateUserPreferences(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("Create new system-admin user (error)", func(t *testing.T) {
		ctx := auth.ContextWithUser(context.Background(), "system-admin")
		body := bytes.NewBufferString(`{"theme": "blue"}`)
		req := httptest.NewRequest("POST", "/preferences", body).WithContext(ctx)
		w := httptest.NewRecorder()

		mockStore.On("GetUser", mock.Anything, "system-admin").Return((*configv1.User)(nil), assert.AnError).Once()

		// Expect CreateUser
		mockStore.On("CreateUser", mock.Anything, mock.MatchedBy(func(u *configv1.User) bool {
			return u.GetId() == "system-admin" && u.GetPreferences()["theme"] == "blue"
		})).Return(nil).Once()

		app.HandleUpdateUserPreferences(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("Create new system-admin user (nil, nil)", func(t *testing.T) {
		ctx := auth.ContextWithUser(context.Background(), "system-admin")
		body := bytes.NewBufferString(`{"theme": "red"}`)
		req := httptest.NewRequest("POST", "/preferences", body).WithContext(ctx)
		w := httptest.NewRecorder()

		// Simulate nil, nil from GetUser
		mockStore.On("GetUser", mock.Anything, "system-admin").Return((*configv1.User)(nil), nil).Once()

		// Expect CreateUser
		mockStore.On("CreateUser", mock.Anything, mock.MatchedBy(func(u *configv1.User) bool {
			return u.GetId() == "system-admin" && u.GetPreferences()["theme"] == "red"
		})).Return(nil).Once()

		app.HandleUpdateUserPreferences(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("Promote file user to DB", func(t *testing.T) {
		ctx := auth.ContextWithUser(context.Background(), "file-user")
		body := bytes.NewBufferString(`{"new": "val"}`)
		req := httptest.NewRequest("POST", "/preferences", body).WithContext(ctx)
		w := httptest.NewRecorder()

		fileUser := configv1.User_builder{
			Id: proto.String("file-user"),
			Roles: []string{"editor"},
		}.Build()
		authManager.SetUsers([]*configv1.User{fileUser})

		mockStore.On("GetUser", mock.Anything, "file-user").Return((*configv1.User)(nil), assert.AnError).Once()

		// Expect CreateUser with roles preserved
		mockStore.On("CreateUser", mock.Anything, mock.MatchedBy(func(u *configv1.User) bool {
			return u.GetId() == "file-user" && u.GetPreferences()["new"] == "val" && u.GetRoles()[0] == "editor"
		})).Return(nil).Once()

		app.HandleUpdateUserPreferences(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})
}
