package sqlite

import (
	"context"
	"os"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func TestStore_UserRolesPersistence(t *testing.T) {
	// Create temp DB
	f, err := os.CreateTemp("", "test-user-db-*.db")
	require.NoError(t, err)
	dbPath := f.Name()
	f.Close()
	defer os.Remove(dbPath)

	db, err := NewDB(dbPath)
	require.NoError(t, err)
	defer db.Close()

	// Initialize schema (NewDB usually does migration? Check NewDB impl)
	// Assuming NewDB handles migration.

	store := NewStore(db)

	ctx := context.Background()
	username := "test-role-user"
	user := &configv1.User{
		Id: proto.String(username),
		Authentication: &configv1.Authentication{
			AuthMethod: &configv1.Authentication_BasicAuth{
				BasicAuth: &configv1.BasicAuth{
					Username:     proto.String(username),
					PasswordHash: proto.String("hash"),
				},
			},
		},
		Roles: []string{"admin", "editor"},
	}

	// Create
	err = store.CreateUser(ctx, user)
	require.NoError(t, err)

	// Get
	fetched, err := store.GetUser(ctx, username)
	require.NoError(t, err)
	require.NotNil(t, fetched)

	assert.Equal(t, username, fetched.GetId())
	assert.ElementsMatch(t, []string{"admin", "editor"}, fetched.GetRoles())

	// List
	list, err := store.ListUsers(ctx)
	require.NoError(t, err)
	found := false
	for _, u := range list {
		if u.GetId() == username {
			found = true
			assert.ElementsMatch(t, []string{"admin", "editor"}, u.GetRoles())
		}
	}
	assert.True(t, found, "User not found in list")
}
