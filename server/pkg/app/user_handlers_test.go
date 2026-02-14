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

func TestUserPreferences(t *testing.T) {
	store := memory.NewStore()
	app := &Application{
		Storage: store,
	}

	// 1. Create a user
	user := config_v1.User_builder{
		Id:          proto.String("test-user"),
		Preferences: map[string]string{"theme": "light"},
	}.Build()
	require.NoError(t, store.CreateUser(context.Background(), user))

	// 2. Test Get Preferences
	t.Run("GetPreferences", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/user/preferences", nil)
		req = req.WithContext(auth.ContextWithUser(req.Context(), "test-user"))
		w := httptest.NewRecorder()

		app.HandleGetUserPreferences(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var prefs map[string]string
		err := json.NewDecoder(w.Body).Decode(&prefs)
		require.NoError(t, err)
		assert.Equal(t, "light", prefs["theme"])
	})

	// 3. Test Update Preferences
	t.Run("UpdatePreferences", func(t *testing.T) {
		body := map[string]string{
			"dashboard_layout": "[{}]",
			"theme":            "dark", // Override
		}
		jsonBody, _ := json.Marshal(body)
		req := httptest.NewRequest("POST", "/api/v1/user/preferences", bytes.NewBuffer(jsonBody))
		req = req.WithContext(auth.ContextWithUser(req.Context(), "test-user"))
		w := httptest.NewRecorder()

		app.HandleUpdateUserPreferences(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		// Verify persistence
		u, err := store.GetUser(context.Background(), "test-user")
		require.NoError(t, err)
		assert.Equal(t, "dark", u.GetPreferences()["theme"])
		assert.Equal(t, "[{}]", u.GetPreferences()["dashboard_layout"])
	})

	// 4. Test Implicit Admin Creation
	t.Run("ImplicitAdmin", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/user/preferences", nil)
		// Simulate global API key auth (sets user="system-admin" but might not be in DB)
		req = req.WithContext(auth.ContextWithUser(req.Context(), "system-admin"))
		w := httptest.NewRecorder()

		// Should return empty prefs (and 200 OK), not 404
		app.HandleGetUserPreferences(w, req)
		assert.Equal(t, http.StatusOK, w.Code)

		// Update
		body := map[string]string{"foo": "bar"}
		jsonBody, _ := json.Marshal(body)
		reqPost := httptest.NewRequest("POST", "/api/v1/user/preferences", bytes.NewBuffer(jsonBody))
		reqPost = reqPost.WithContext(auth.ContextWithUser(reqPost.Context(), "system-admin"))
		wPost := httptest.NewRecorder()

		app.HandleUpdateUserPreferences(wPost, reqPost)
		assert.Equal(t, http.StatusOK, wPost.Code)

		// Verify DB creation
		u, err := store.GetUser(context.Background(), "system-admin")
		require.NoError(t, err)
		assert.Equal(t, "bar", u.GetPreferences()["foo"])
	})
}
