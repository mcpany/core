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
	"google.golang.org/protobuf/proto"
)

func TestUserPreferencesHandlers(t *testing.T) {
	// Setup
	store := memory.NewStore()
	app := &Application{
		Storage: store,
	}

	userID := "test-user"
	ctx := auth.ContextWithUser(context.Background(), userID)

	// Create initial user
	user := config_v1.User_builder{
		Id: proto.String(userID),
	}.Build()
	require.NoError(t, store.CreateUser(ctx, user))

	// Test Update Preferences
	t.Run("Update Preferences", func(t *testing.T) {
		prefs := map[string]string{
			"dashboard-layout": "some-json",
			"theme":            "dark",
		}
		body, _ := json.Marshal(prefs)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/user/preferences", bytes.NewBuffer(body))
		req = req.WithContext(ctx)
		w := httptest.NewRecorder()

		app.handleUpdateUserPreferences(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		// Verify storage
		updatedUser, err := store.GetUser(ctx, userID)
		require.NoError(t, err)
		assert.Equal(t, prefs, updatedUser.GetPreferences())
	})

	// Test Get Preferences
	t.Run("Get Preferences", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/user/preferences", nil)
		req = req.WithContext(ctx)
		w := httptest.NewRecorder()

		app.handleGetUserPreferences(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp map[string]string
		err := json.NewDecoder(w.Body).Decode(&resp)
		require.NoError(t, err)
		assert.Equal(t, "some-json", resp["dashboard-layout"])
		assert.Equal(t, "dark", resp["theme"])
	})

	// Test Implicit Creation
	t.Run("Implicit User Creation", func(t *testing.T) {
		newUserID := "implicit-user"
		newCtx := auth.ContextWithUser(context.Background(), newUserID)

		prefs := map[string]string{
			"new-pref": "value",
		}
		body, _ := json.Marshal(prefs)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/user/preferences", bytes.NewBuffer(body))
		req = req.WithContext(newCtx)
		w := httptest.NewRecorder()

		app.handleUpdateUserPreferences(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		// Verify storage
		newUser, err := store.GetUser(newCtx, newUserID)
		require.NoError(t, err)
		assert.Equal(t, prefs, newUser.GetPreferences())
	})
}
