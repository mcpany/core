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
)

func TestHandleGetUserPreferences(t *testing.T) {
	// Setup
	app := &Application{
		Storage: memory.NewStore(),
	}

	// Create a user in storage
	user := &configv1.User{}
	user.SetId("test-user")
	prefs := map[string]string{"theme": "dark"}
	user.SetPreferences(prefs)
	require.NoError(t, app.Storage.CreateUser(context.Background(), user))

	// Request
	req := httptest.NewRequest(http.MethodGet, "/api/v1/user/preferences", nil)
	ctx := auth.ContextWithUser(req.Context(), "test-user")
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	// Execute
	app.HandleGetUserPreferences(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, "dark", resp["theme"])
}

func TestHandleGetUserPreferences_NotFound(t *testing.T) {
	// Setup
	app := &Application{
		Storage: memory.NewStore(),
	}

	// Request for non-existent user
	req := httptest.NewRequest(http.MethodGet, "/api/v1/user/preferences", nil)
	ctx := auth.ContextWithUser(req.Context(), "system-admin")
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	// Execute
	app.HandleGetUserPreferences(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Empty(t, resp)
}

func TestHandleUpdateUserPreferences_UpdateExisting(t *testing.T) {
	// Setup
	app := &Application{
		Storage: memory.NewStore(),
	}

	// Create a user
	user := &configv1.User{}
	user.SetId("test-user")
	user.SetPreferences(map[string]string{"theme": "light"})
	require.NoError(t, app.Storage.CreateUser(context.Background(), user))

	// Request
	body := map[string]string{"theme": "dark", "layout": "grid"}
	bodyBytes, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/user/preferences", bytes.NewReader(bodyBytes))
	ctx := auth.ContextWithUser(req.Context(), "test-user")
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	// Execute
	app.HandleUpdateUserPreferences(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, "dark", resp["theme"])
	assert.Equal(t, "grid", resp["layout"])

	// Verify storage
	updatedUser, err := app.Storage.GetUser(context.Background(), "test-user")
	require.NoError(t, err)
	assert.Equal(t, "dark", updatedUser.GetPreferences()["theme"])
}

func TestHandleUpdateUserPreferences_CreateNew(t *testing.T) {
	// Setup
	app := &Application{
		Storage: memory.NewStore(),
	}

	// Request for new user
	body := map[string]string{"theme": "dark"}
	bodyBytes, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/user/preferences", bytes.NewReader(bodyBytes))
	ctx := auth.ContextWithUser(req.Context(), "new-user")
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	// Execute
	app.HandleUpdateUserPreferences(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, "dark", resp["theme"])

	// Verify storage
	newUser, err := app.Storage.GetUser(context.Background(), "new-user")
	require.NoError(t, err)
	assert.Equal(t, "dark", newUser.GetPreferences()["theme"])
}
