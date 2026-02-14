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
)

func TestHandleUserPreferences(t *testing.T) {
	app := NewApplication()
	app.fs = afero.NewMemMapFs()
	store := memory.NewStore()
	app.Storage = store

	t.Run("Get Preferences - User Exists", func(t *testing.T) {
		// Create a user with preferences
		userID := "user-prefs-1"
		prefs := map[string]string{
			"theme": "dark",
			"lang":  "en",
		}
		user := configv1.User_builder{
			Id:          &userID,
			Preferences: prefs,
		}.Build()
		require.NoError(t, store.CreateUser(context.Background(), user))

		req := httptest.NewRequest(http.MethodGet, "/user/preferences", nil)
		ctx := auth.ContextWithUser(req.Context(), userID)
		req = req.WithContext(ctx)

		w := httptest.NewRecorder()
		app.handleGetUserPreferences(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var respPrefs map[string]string
		err := json.Unmarshal(w.Body.Bytes(), &respPrefs)
		require.NoError(t, err)
		assert.Equal(t, "dark", respPrefs["theme"])
		assert.Equal(t, "en", respPrefs["lang"])
	})

	t.Run("Get Preferences - User Not Found (Empty)", func(t *testing.T) {
		userID := "user-implicit"
		// User does NOT exist in store

		req := httptest.NewRequest(http.MethodGet, "/user/preferences", nil)
		ctx := auth.ContextWithUser(req.Context(), userID)
		req = req.WithContext(ctx)

		w := httptest.NewRecorder()
		app.handleGetUserPreferences(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var respPrefs map[string]string
		err := json.Unmarshal(w.Body.Bytes(), &respPrefs)
		require.NoError(t, err)
		assert.Empty(t, respPrefs)
	})

	t.Run("Update Preferences - User Exists", func(t *testing.T) {
		userID := "user-prefs-update"
		user := configv1.User_builder{
			Id:          &userID,
			Preferences: map[string]string{"theme": "light"},
		}.Build()
		require.NoError(t, store.CreateUser(context.Background(), user))

		newPrefs := map[string]string{
			"theme": "dark",
			"layout": "grid",
		}
		body, _ := json.Marshal(newPrefs)
		req := httptest.NewRequest(http.MethodPost, "/user/preferences", bytes.NewReader(body))
		ctx := auth.ContextWithUser(req.Context(), userID)
		req = req.WithContext(ctx)

		w := httptest.NewRecorder()
		app.handleUpdateUserPreferences(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		// Verify update
		u, err := store.GetUser(context.Background(), userID)
		require.NoError(t, err)
		assert.Equal(t, "dark", u.GetPreferences()["theme"])
		assert.Equal(t, "grid", u.GetPreferences()["layout"])
	})

	t.Run("Update Preferences - User Not Found (Create)", func(t *testing.T) {
		userID := "user-prefs-create"
		// User does NOT exist

		newPrefs := map[string]string{
			"dashboard": "custom",
		}
		body, _ := json.Marshal(newPrefs)
		req := httptest.NewRequest(http.MethodPost, "/user/preferences", bytes.NewReader(body))
		ctx := auth.ContextWithUser(req.Context(), userID)
		req = req.WithContext(ctx)

		w := httptest.NewRecorder()
		app.handleUpdateUserPreferences(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		// Verify creation
		u, err := store.GetUser(context.Background(), userID)
		require.NoError(t, err)
		assert.NotNil(t, u)
		assert.Equal(t, userID, u.GetId())
		assert.Equal(t, "custom", u.GetPreferences()["dashboard"])
	})

	t.Run("Unauthorized", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/user/preferences", nil)
		// No auth context

		w := httptest.NewRecorder()
		app.handleGetUserPreferences(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}
