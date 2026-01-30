// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package auth

import (
	"context"
	"net/http"
	"testing"

	"github.com/gogo/protobuf/proto"
	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/storage/sqlite"
	"github.com/mcpany/core/server/pkg/util/passhash"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAuthStorageFallback(t *testing.T) {
	// Setup SQLite DB
	dbPath := t.TempDir() + "/test_auth.db"
	db, err := sqlite.NewDB(dbPath)
	require.NoError(t, err)
	defer db.Close()

	store := sqlite.NewStore(db)

	// Create Auth Manager
	am := NewManager()
	am.SetStorage(store)

	// Create user in Storage (but not in memory)
	username := "storageuser"
	password := "securepass"
	hashedPassword, err := passhash.Password(password)
	require.NoError(t, err)

	user := &configv1.User{
		Id: proto.String(username),
		Authentication: &configv1.Authentication{
			AuthMethod: &configv1.Authentication_BasicAuth{
				BasicAuth: &configv1.BasicAuth{
					Username:     proto.String(username),
					PasswordHash: proto.String(hashedPassword),
				},
			},
		},
		Roles: []string{"viewer"},
	}

	err = store.CreateUser(context.Background(), user)
	require.NoError(t, err)

	// Create Request
	req, err := http.NewRequest("GET", "/", nil)
	require.NoError(t, err)
	req.SetBasicAuth(username, password)

	// Test Authenticate (via checkBasicAuthWithUsers fallback)
	ctx := context.Background()
	ctx, err = am.Authenticate(ctx, "non_existent_service", req)
	assert.NoError(t, err)

	userID, ok := UserFromContext(ctx)
	assert.True(t, ok)
	assert.Equal(t, username, userID)

	roles, ok := RolesFromContext(ctx)
	assert.True(t, ok)
	assert.Contains(t, roles, "viewer")
}

func TestAuthStorageFallback_InvalidPassword(t *testing.T) {
	// Setup SQLite DB
	dbPath := t.TempDir() + "/test_auth_invalid.db"
	db, err := sqlite.NewDB(dbPath)
	require.NoError(t, err)
	defer db.Close()

	store := sqlite.NewStore(db)
	am := NewManager()
	am.SetStorage(store)

	username := "storageuser"
	password := "securepass"
	hashedPassword, err := passhash.Password(password)
	require.NoError(t, err)

	user := &configv1.User{
		Id: proto.String(username),
		Authentication: &configv1.Authentication{
			AuthMethod: &configv1.Authentication_BasicAuth{
				BasicAuth: &configv1.BasicAuth{
					Username:     proto.String(username),
					PasswordHash: proto.String(hashedPassword),
				},
			},
		},
	}
	err = store.CreateUser(context.Background(), user)
	require.NoError(t, err)

	req, err := http.NewRequest("GET", "/", nil)
	require.NoError(t, err)
	req.SetBasicAuth(username, "wrongpass")

	ctx := context.Background()
	_, err = am.Authenticate(ctx, "non_existent_service", req)
	assert.Error(t, err)
	// Authenticate swallows the specific error and returns a generic one if fallback fails
	assert.Contains(t, err.Error(), "unauthorized")
}
