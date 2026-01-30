package auth

import (
	"context"
	"net/http"
	"os"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/storage/sqlite"
	"github.com/mcpany/core/server/pkg/util/passhash"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func TestManager_checkBasicAuthWithUsers_StorageFallback_SQLite(t *testing.T) {
	// Create temp DB
	f, err := os.CreateTemp("", "test-auth-db-*.db")
	require.NoError(t, err)
	dbPath := f.Name()
	f.Close()
	defer os.Remove(dbPath)

	db, err := sqlite.NewDB(dbPath)
	require.NoError(t, err)
	defer db.Close()

	store := sqlite.NewStore(db)

	am := NewManager()
	am.SetStorage(store)

	ctx := context.Background()
	username := "storage-user-sqlite"
	password := "password123"
	hash, _ := passhash.Password(password)

	// Create user in storage
	user := &configv1.User{
		Id: proto.String(username),
		Authentication: &configv1.Authentication{
			AuthMethod: &configv1.Authentication_BasicAuth{
				BasicAuth: &configv1.BasicAuth{
					Username:     proto.String(username),
					PasswordHash: proto.String(hash),
				},
			},
		},
		Roles: []string{"admin"},
	}
	require.NoError(t, store.CreateUser(ctx, user))

	// Ensure user is NOT in memory
	_, found := am.GetUser(username)
	require.False(t, found)

	// Prepare request
	req, _ := http.NewRequest("GET", "/", nil)
	req.SetBasicAuth(username, password)

	// Test checkBasicAuthWithUsers
	newCtx, err := am.checkBasicAuthWithUsers(ctx, req)
	require.NoError(t, err)

	// Verify Context
	userID, ok := UserFromContext(newCtx)
	assert.True(t, ok)
	assert.Equal(t, username, userID)

	roles, ok := RolesFromContext(newCtx)
	assert.True(t, ok)
	assert.Contains(t, roles, "admin")
}
