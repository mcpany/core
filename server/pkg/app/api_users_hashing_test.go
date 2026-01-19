// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/mcpany/core/server/pkg/auth"
	"github.com/mcpany/core/server/pkg/storage/memory"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateUser_PasswordHashing(t *testing.T) {
	app := NewApplication()
	app.fs = afero.NewMemMapFs() // Avoid nil pointer in ReloadConfig
	app.AuthManager = auth.NewManager()
	store := memory.NewStore()
	app.Storage = store
	handler := app.handleUsers(store)

	// JSON payload with plain password in password_hash field (mimicking UI behavior)
	payload := `{"user": {"id": "test-user-hash", "authentication": {"basic_auth": {"username": "test", "password_hash": "plain-password"}}}}`

	req := httptest.NewRequest(http.MethodPost, "/users", strings.NewReader(payload))
	w := httptest.NewRecorder()

	handler(w, req)

	// We don't assert w.Code because ReloadConfig might fail due to missing config files,
	// but the user creation happens before ReloadConfig in the critical path for this test.
	// Actually, CreateUser is called before ReloadConfig.

	user, err := store.GetUser(context.Background(), "test-user-hash")
	require.NoError(t, err)
	require.NotNil(t, user)

	auth := user.GetAuthentication().GetBasicAuth()
	require.NotNil(t, auth)

	hash := auth.GetPasswordHash()

	// Ensure it is NOT the plain text password
	assert.NotEqual(t, "plain-password", hash, "Password should be hashed")
	// Ensure it IS a bcrypt hash
	assert.True(t, strings.HasPrefix(hash, "$2a$"), "Password should be bcrypt hash")
}
