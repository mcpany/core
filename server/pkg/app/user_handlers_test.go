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
	"github.com/stretchr/testify/require"
)

func TestHandleUpdateUserPreferences(t *testing.T) {
	// Create a new memory store
	store := memory.NewStore()
	// Create a new auth manager
	authManager := auth.NewManager()
	authManager.SetStorage(store)

	// Create a mock application
	app := &Application{
		Storage:     store,
		AuthManager: authManager,
	}

	t.Run("CreateImplicitUser", func(t *testing.T) {
		userID := "system-admin"
		prefs := map[string]string{
			"dashboard_layout": `[{"instanceId":"1","type":"metrics","size":"half"}]`,
		}
		body, _ := json.Marshal(prefs)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/user/preferences", bytes.NewBuffer(body))
		// Inject user into context
		ctx := auth.ContextWithUser(req.Context(), userID)
		req = req.WithContext(ctx)

		w := httptest.NewRecorder()

		app.HandleUpdateUserPreferences(w, req)

		require.Equal(t, http.StatusOK, w.Code)

		// Verify in storage
		user, err := store.GetUser(context.Background(), userID)
		require.NoError(t, err)
		require.NotNil(t, user)
		if user.GetPreferences() == nil {
			t.Fatal("Preferences map is nil")
		}
		require.Equal(t, prefs["dashboard_layout"], user.GetPreferences()["dashboard_layout"])
	})

	t.Run("UpdateExistingUser", func(t *testing.T) {
		userID := "existing-user"
		// Create existing user in store
		// Note: We need to build the object correctly using proto builder if possible, or manual struct
		// Using manual struct literal for simplicity if builder is annoying, but builder is safer.
		// Builder usage:
		existingUser := configv1.User_builder{
			Id: &userID,
			Preferences: map[string]string{"theme": "dark"},
		}.Build()

		err := store.CreateUser(context.Background(), existingUser)
		require.NoError(t, err)

		newPrefs := map[string]string{
			"theme": "light",
			"dashboard_layout": "[]",
		}
		body, _ := json.Marshal(newPrefs)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/user/preferences", bytes.NewBuffer(body))
		ctx := auth.ContextWithUser(req.Context(), userID)
		req = req.WithContext(ctx)
		w := httptest.NewRecorder()

		app.HandleUpdateUserPreferences(w, req)

		require.Equal(t, http.StatusOK, w.Code)

		// Verify update
		user, err := store.GetUser(context.Background(), userID)
		require.NoError(t, err)
		require.Equal(t, "light", user.GetPreferences()["theme"])
	})

	t.Run("MigrateConfigUser", func(t *testing.T) {
		// Scenario: User exists in AuthManager (from config) but not in Storage.
		// Handler should create user in Storage and COPY roles/auth.
		userID := "config-user"

		role := "admin"
		configUser := configv1.User_builder{
			Id: &userID,
			Roles: []string{role},
		}.Build()

		authManager.SetUsers([]*configv1.User{configUser})

		prefs := map[string]string{"foo": "bar"}
		body, _ := json.Marshal(prefs)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/user/preferences", bytes.NewBuffer(body))
		ctx := auth.ContextWithUser(req.Context(), userID)
		req = req.WithContext(ctx)
		w := httptest.NewRecorder()

		app.HandleUpdateUserPreferences(w, req)

		require.Equal(t, http.StatusOK, w.Code)

		// Verify created in DB with roles copied
		user, err := store.GetUser(context.Background(), userID)
		require.NoError(t, err)
		require.Equal(t, "bar", user.GetPreferences()["foo"])
		require.Equal(t, []string{"admin"}, user.GetRoles())
	})
}

func TestHandleGetUserPreferences(t *testing.T) {
	store := memory.NewStore()
	authManager := auth.NewManager()

	app := &Application{
		Storage:     store,
		AuthManager: authManager,
	}

	t.Run("GetExistingPreferences", func(t *testing.T) {
		userID := "user1"
		user := configv1.User_builder{
			Id: &userID,
			Preferences: map[string]string{"key": "val"},
		}.Build()
		require.NoError(t, store.CreateUser(context.Background(), user))

		req := httptest.NewRequest(http.MethodGet, "/api/v1/user/preferences", nil)
		ctx := auth.ContextWithUser(req.Context(), userID)
		req = req.WithContext(ctx)
		w := httptest.NewRecorder()

		app.HandleGetUserPreferences(w, req)

		require.Equal(t, http.StatusOK, w.Code)
		var res map[string]string
		require.NoError(t, json.NewDecoder(w.Body).Decode(&res))
		require.Equal(t, "val", res["key"])
	})

	t.Run("GetMissingUserReturnsEmpty", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/user/preferences", nil)
		ctx := auth.ContextWithUser(req.Context(), "missing")
		req = req.WithContext(ctx)
		w := httptest.NewRecorder()

		app.HandleGetUserPreferences(w, req)

		require.Equal(t, http.StatusOK, w.Code)
		var res map[string]string
		require.NoError(t, json.NewDecoder(w.Body).Decode(&res))
		require.Empty(t, res)
	})
}
