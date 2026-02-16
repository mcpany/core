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

	config_v1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/auth"
	"github.com/mcpany/core/server/pkg/storage/memory"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHandleUserPreferences(t *testing.T) {
	// Setup
	store := memory.NewStore()
	authManager := auth.NewManager()
	// Need to set users in AuthManager to mock an authenticated user
	userID := "test-user"
	testUser := config_v1.User_builder{
		Id:    &userID,
		Roles: []string{"user"},
	}.Build()

	authManager.SetUsers([]*config_v1.User{testUser})

	app := &Application{
		Storage:     store,
		AuthManager: authManager,
	}

	// Test Case 1: Get Preferences (Empty)
	t.Run("Get Empty Preferences", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/user/preferences", nil)
		ctx := auth.ContextWithUser(req.Context(), "test-user")
		req = req.WithContext(ctx)
		w := httptest.NewRecorder()

		app.HandleGetUserPreferences(w, req)

		require.Equal(t, http.StatusOK, w.Code)
		var prefs map[string]string
		err := json.NewDecoder(w.Body).Decode(&prefs)
		require.NoError(t, err)
		assert.Empty(t, prefs)
	})

	// Test Case 2: Update Preferences
	t.Run("Update Preferences", func(t *testing.T) {
		prefs := map[string]string{
			"theme":  "dark",
			"layout": "grid",
		}
		body, _ := json.Marshal(prefs)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/user/preferences", bytes.NewBuffer(body))
		ctx := auth.ContextWithUser(req.Context(), "test-user")
		req = req.WithContext(ctx)
		w := httptest.NewRecorder()

		app.HandleUpdateUserPreferences(w, req)

		require.Equal(t, http.StatusOK, w.Code)
		var respPrefs map[string]string
		err := json.NewDecoder(w.Body).Decode(&respPrefs)
		require.NoError(t, err)
		assert.Equal(t, "dark", respPrefs["theme"])
		assert.Equal(t, "grid", respPrefs["layout"])

		// Verify it persists in storage
		user, err := store.GetUser(context.Background(), "test-user")
		require.NoError(t, err)
		assert.Equal(t, "dark", user.GetPreferences()["theme"])
	})

	// Test Case 3: Get Preferences (Persisted)
	t.Run("Get Persisted Preferences", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/user/preferences", nil)
		ctx := auth.ContextWithUser(req.Context(), "test-user")
		req = req.WithContext(ctx)
		w := httptest.NewRecorder()

		app.HandleGetUserPreferences(w, req)

		require.Equal(t, http.StatusOK, w.Code)
		var prefs map[string]string
		err := json.NewDecoder(w.Body).Decode(&prefs)
		require.NoError(t, err)
		assert.Equal(t, "dark", prefs["theme"])
	})
}
