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
	"github.com/mcpany/core/server/pkg/profile"
	"github.com/mcpany/core/server/pkg/storage/memory"
	"github.com/mcpany/core/server/pkg/tool"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHandleUserPreferences(t *testing.T) {
	store := memory.NewStore()
	app := &Application{
		AuthManager:     auth.NewManager(),
		Storage:         store,
		fs:              afero.NewMemMapFs(),
		configPaths:     []string{},
		SettingsManager: NewGlobalSettingsManager("", nil, nil),
		ToolManager:     tool.NewManager(nil),
		ProfileManager:  profile.NewManager(nil),
	}

	handler := app.handleUserPreferences(store)

	t.Run("Unauthorized", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/user/preferences", nil)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("ImplicitSystemAdmin", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/user/preferences", nil)
		ctx := auth.ContextWithUser(req.Context(), "system-admin")
		req = req.WithContext(ctx)

		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var prefs map[string]string
		err := json.Unmarshal(w.Body.Bytes(), &prefs)
		require.NoError(t, err)
		assert.Empty(t, prefs)

		// Verify user created in store
		user, err := store.GetUser(context.Background(), "system-admin")
		require.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, "system-admin", user.GetId())
	})

	t.Run("UpdatePreferences", func(t *testing.T) {
		newPrefs := map[string]string{
			"dashboard_layout": "{}",
			"theme":            "dark",
		}
		body, _ := json.Marshal(newPrefs)
		req := httptest.NewRequest("POST", "/api/v1/user/preferences", bytes.NewReader(body))
		ctx := auth.ContextWithUser(req.Context(), "system-admin")
		req = req.WithContext(ctx)

		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var prefs map[string]string
		err := json.Unmarshal(w.Body.Bytes(), &prefs)
		require.NoError(t, err)
		assert.Equal(t, "dark", prefs["theme"])

		// Verify persistence
		user, err := store.GetUser(context.Background(), "system-admin")
		require.NoError(t, err)
		assert.Equal(t, "dark", user.GetPreferences()["theme"])
	})

	t.Run("ConfigUserPromotion", func(t *testing.T) {
		// Mock a user in AuthManager that is NOT in DB
		configUser := &configv1.User{}
		configUser.SetId("config-user")
		configUser.SetRoles([]string{"viewer"})

		app.AuthManager.SetUsers([]*configv1.User{configUser})

		req := httptest.NewRequest("GET", "/api/v1/user/preferences", nil)
		ctx := auth.ContextWithUser(req.Context(), "config-user")
		req = req.WithContext(ctx)

		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		// Check if user is now in DB
		user, err := store.GetUser(context.Background(), "config-user")
		require.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, "viewer", user.GetRoles()[0])

		// Also verify AuthManager still has the user (via ReloadConfig)
		// Since ReloadConfig runs and loads from store, and store now has config-user, it should be fine.
		u, found := app.AuthManager.GetUser("config-user")
		assert.True(t, found)
		assert.Equal(t, "viewer", u.GetRoles()[0])
	})
}
