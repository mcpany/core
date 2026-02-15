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
	"github.com/mcpany/core/server/pkg/storage/memory"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func TestHandleUserPreferences(t *testing.T) {
	// Setup
	store := memory.NewStore()
	app := &Application{
		Storage:     store,
		AuthManager: auth.NewManager(),
	}
	app.AuthManager.SetStorage(store)

	// Create a test user in DB
	userID := "test-user"
	initialUser := configv1.User_builder{
		Id:    proto.String(userID),
		Roles: []string{"user"},
	}.Build()
	err := store.CreateUser(context.Background(), initialUser)
	require.NoError(t, err)

	// Test GET (Empty initially)
	t.Run("Get Preferences Empty", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/user/preferences", nil)
		// Inject Auth Context
		ctx := auth.ContextWithUser(req.Context(), userID)
		req = req.WithContext(ctx)

		w := httptest.NewRecorder()
		app.handleGetUserPreferences(w, req)

		resp := w.Result()
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var prefs map[string]string
		err := json.NewDecoder(resp.Body).Decode(&prefs)
		require.NoError(t, err)
		assert.Empty(t, prefs)
	})

	// Test POST (Update)
	t.Run("Update Preferences", func(t *testing.T) {
		prefs := map[string]string{
			"theme": "dark",
			"layout": "grid",
		}
		body, _ := json.Marshal(prefs)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/user/preferences", bytes.NewBuffer(body))
		// Inject Auth Context
		ctx := auth.ContextWithUser(req.Context(), userID)
		req = req.WithContext(ctx)

		w := httptest.NewRecorder()
		app.handleUpdateUserPreferences(w, req)

		resp := w.Result()
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		// Verify DB
		user, err := store.GetUser(context.Background(), userID)
		require.NoError(t, err)
		assert.Equal(t, "dark", user.GetPreferences()["theme"])
		assert.Equal(t, "grid", user.GetPreferences()["layout"])
	})

	// Test GET (After Update)
	t.Run("Get Preferences After Update", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/user/preferences", nil)
		// Inject Auth Context
		ctx := auth.ContextWithUser(req.Context(), userID)
		req = req.WithContext(ctx)

		w := httptest.NewRecorder()
		app.handleGetUserPreferences(w, req)

		resp := w.Result()
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var prefs map[string]string
		err := json.NewDecoder(resp.Body).Decode(&prefs)
		require.NoError(t, err)
		assert.Equal(t, "dark", prefs["theme"])
	})

	// Test Implicit System Admin Creation
	t.Run("Implicit System Admin", func(t *testing.T) {
		adminID := "system-admin"
		prefs := map[string]string{"admin_mode": "true"}
		body, _ := json.Marshal(prefs)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/user/preferences", bytes.NewBuffer(body))
		// Inject Auth Context for system-admin
		ctx := auth.ContextWithUser(req.Context(), adminID)
		req = req.WithContext(ctx)

		w := httptest.NewRecorder()
		app.handleUpdateUserPreferences(w, req)

		resp := w.Result()
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		// Verify DB creation
		user, err := store.GetUser(context.Background(), adminID)
		require.NoError(t, err)
		assert.Equal(t, "true", user.GetPreferences()["admin_mode"])
	})

	// Test Config User Promotion
	t.Run("Config User Promotion", func(t *testing.T) {
		configUserID := "config-user"
		// Add to AuthManager (simulate config load)
		configUser := configv1.User_builder{
			Id: proto.String(configUserID),
			Authentication: configv1.Authentication_builder{
				BasicAuth: configv1.BasicAuth_builder{
					Username: proto.String("test"),
				}.Build(),
			}.Build(),
		}.Build()
		app.AuthManager.SetUsers([]*configv1.User{configUser})

		prefs := map[string]string{"promoted": "yes"}
		body, _ := json.Marshal(prefs)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/user/preferences", bytes.NewBuffer(body))
		ctx := auth.ContextWithUser(req.Context(), configUserID)
		req = req.WithContext(ctx)

		w := httptest.NewRecorder()
		app.handleUpdateUserPreferences(w, req)

		resp := w.Result()
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		// Verify DB creation
		user, err := store.GetUser(context.Background(), configUserID)
		require.NoError(t, err)
		assert.Equal(t, "yes", user.GetPreferences()["promoted"])
		// Verify Auth was preserved (copied)
		assert.NotNil(t, user.GetAuthentication().GetBasicAuth())
		assert.Equal(t, "test", user.GetAuthentication().GetBasicAuth().GetUsername())
	})
}
