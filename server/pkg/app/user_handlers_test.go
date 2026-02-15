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

	v1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/auth"
	"github.com/mcpany/core/server/pkg/storage/memory"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHandleGetUserPreferences(t *testing.T) {
	store := memory.NewStore()
	app := &Application{Storage: store}

	// Case 1: User not authenticated
	req := httptest.NewRequest("GET", "/api/v1/user/preferences", nil)
	w := httptest.NewRecorder()
	app.handleGetUserPreferences(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code)

	// Case 2: User authenticated, but not in DB (should return empty)
	ctx := auth.ContextWithUser(context.Background(), "test-user")
	req = httptest.NewRequest("GET", "/api/v1/user/preferences", nil).WithContext(ctx)
	w = httptest.NewRecorder()
	app.handleGetUserPreferences(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	var prefs map[string]string
	err := json.NewDecoder(w.Body).Decode(&prefs)
	require.NoError(t, err)
	assert.Empty(t, prefs)

	// Case 3: User authenticated and exists with preferences
	user := &v1.User{}
	user.SetId("test-user-2")
	user.SetPreferences(map[string]string{"theme": "dark", "layout": "grid"})
	err = store.CreateUser(context.Background(), user)
	require.NoError(t, err)

	ctx = auth.ContextWithUser(context.Background(), "test-user-2")
	req = httptest.NewRequest("GET", "/api/v1/user/preferences", nil).WithContext(ctx)
	w = httptest.NewRecorder()
	app.handleGetUserPreferences(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	var prefs2 map[string]string
	err = json.NewDecoder(w.Body).Decode(&prefs2)
	require.NoError(t, err)
	assert.Equal(t, "dark", prefs2["theme"])
	assert.Equal(t, "grid", prefs2["layout"])
}

func TestHandleUpdateUserPreferences(t *testing.T) {
	store := memory.NewStore()
	// Mock AuthManager needed if we want to copy roles, but for this test we can skip or mock it if Application allows
	// Application struct has AuthManager *auth.Manager
	authManager := auth.NewManager()
	app := &Application{Storage: store, AuthManager: authManager}

	// Case 1: User not authenticated
	req := httptest.NewRequest("POST", "/api/v1/user/preferences", nil)
	w := httptest.NewRecorder()
	app.handleUpdateUserPreferences(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code)

	// Case 2: Create new user preferences (implicit user creation)
	newPrefs := map[string]string{"dashboard-layout": "[]"}
	body, _ := json.Marshal(newPrefs)

	ctx := auth.ContextWithUser(context.Background(), "new-user")
	req = httptest.NewRequest("POST", "/api/v1/user/preferences", bytes.NewBuffer(body)).WithContext(ctx)
	w = httptest.NewRecorder()
	app.handleUpdateUserPreferences(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	// Verify in DB
	user, err := store.GetUser(context.Background(), "new-user")
	require.NoError(t, err)
	assert.Equal(t, "[]", user.GetPreferences()["dashboard-layout"])

	// Case 3: Update existing user preferences
	updatePrefs := map[string]string{"dashboard-layout": "[1,2,3]", "theme": "light"}
	body, _ = json.Marshal(updatePrefs)

	req = httptest.NewRequest("POST", "/api/v1/user/preferences", bytes.NewBuffer(body)).WithContext(ctx)
	w = httptest.NewRecorder()
	app.handleUpdateUserPreferences(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	// Verify in DB
	user, err = store.GetUser(context.Background(), "new-user")
	require.NoError(t, err)
	assert.Equal(t, "[1,2,3]", user.GetPreferences()["dashboard-layout"])
	assert.Equal(t, "light", user.GetPreferences()["theme"])
}
