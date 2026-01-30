// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package integration

import (
	"context"
	"testing"

	"github.com/gogo/protobuf/proto"
	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/storage/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserRolesPersistence_SQLite(t *testing.T) {
	dbPath := t.TempDir() + "/roles.db"
	db, err := sqlite.NewDB(dbPath)
	require.NoError(t, err)
	defer db.Close()

	store := sqlite.NewStore(db)
	ctx := context.Background()

	// Create user with roles
	user := &configv1.User{
		Id:    proto.String("role_test_user"),
		Roles: []string{"admin", "editor"},
	}

	err = store.CreateUser(ctx, user)
	require.NoError(t, err)

	// Retrieve user
	retrieved, err := store.GetUser(ctx, "role_test_user")
	require.NoError(t, err)
	require.NotNil(t, retrieved)

	assert.ElementsMatch(t, []string{"admin", "editor"}, retrieved.GetRoles())

	// Update roles
	retrieved.Roles = []string{"viewer"}
	err = store.UpdateUser(ctx, retrieved)
	require.NoError(t, err)

	// Verify update
	updated, err := store.GetUser(ctx, "role_test_user")
	require.NoError(t, err)
	assert.ElementsMatch(t, []string{"viewer"}, updated.GetRoles())
}
