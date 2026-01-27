// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"context"
	"os"
	"testing"

	"github.com/mcpany/core/server/pkg/storage/sqlite"
	"github.com/mcpany/core/server/pkg/util/passhash"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReproInsecureDefaultPassword(t *testing.T) {
	// Unset the env var to trigger default behavior
	os.Unsetenv("MCPANY_ADMIN_INIT_PASSWORD")
	os.Setenv("MCPANY_ADMIN_INIT_USERNAME", "admin_repro") // Use a unique user to be safe

	// Create in-memory SQLite store
	db, err := sqlite.NewDB(":memory:")
	require.NoError(t, err)
	defer db.Close()

	store := sqlite.NewStore(db)

	app := &Application{}

	// Call initializeAdminUser
	// Note: We are testing the private method via this test in the same package
	err = app.initializeAdminUser(context.Background(), store)
	require.NoError(t, err)

	// Verify user created
	user, err := store.GetUser(context.Background(), "admin_repro")
	require.NoError(t, err)
	require.NotNil(t, user)

	// Verify password is NOT "password"
	auth := user.GetAuthentication().GetBasicAuth()
	require.NotNil(t, auth)

	hash := auth.GetPasswordHash()
	assert.False(t, passhash.CheckPassword("password", hash), "Default password should NOT be 'password'")
}
