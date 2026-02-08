package sqlite

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func TestStore_Close(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "mcpany-test-store-close-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	dbPath := filepath.Join(tmpDir, "test.db")
	db, err := NewDB(dbPath)
	require.NoError(t, err)

	store := NewStore(db)

	// Test Close
	err = store.Close()
	assert.NoError(t, err)

	// Verify it's closed by trying to query
	_, err = store.ListServices(context.Background())
	assert.Error(t, err)
}

func TestStore_GetGlobalSettings_Empty(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "mcpany-test-store-gs-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	dbPath := filepath.Join(tmpDir, "test.db")
	db, err := NewDB(dbPath)
	require.NoError(t, err)
	defer db.Close()

	store := NewStore(db)

	gs, err := store.GetGlobalSettings(context.Background())
	assert.NoError(t, err)
	assert.Nil(t, gs)
}

func TestStore_Validations(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "mcpany-test-store-val-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	dbPath := filepath.Join(tmpDir, "test.db")
	db, err := NewDB(dbPath)
	require.NoError(t, err)
	defer db.Close()

	store := NewStore(db)
	ctx := context.Background()

	t.Run("SaveToken Missing IDs", func(t *testing.T) {
		err := store.SaveToken(ctx, configv1.UserToken_builder{UserId: proto.String("u1")}.Build())
		assert.ErrorContains(t, err, "user ID and service ID are required")

		err = store.SaveToken(ctx, configv1.UserToken_builder{ServiceId: proto.String("s1")}.Build())
		assert.ErrorContains(t, err, "user ID and service ID are required")
	})

	t.Run("SaveCredential Missing ID", func(t *testing.T) {
		err := store.SaveCredential(ctx, configv1.Credential_builder{}.Build())
		assert.ErrorContains(t, err, "credential ID is required")
	})

	t.Run("SaveProfile Missing Name", func(t *testing.T) {
		err := store.SaveProfile(ctx, configv1.ProfileDefinition_builder{}.Build())
		assert.ErrorContains(t, err, "profile name is required")
	})

	t.Run("SaveServiceCollection Missing Name", func(t *testing.T) {
		err := store.SaveServiceCollection(ctx, configv1.Collection_builder{}.Build())
		assert.ErrorContains(t, err, "collection name is required")
	})

	t.Run("SaveSecret Missing ID", func(t *testing.T) {
		err := store.SaveSecret(ctx, configv1.Secret_builder{}.Build())
		assert.ErrorContains(t, err, "secret id is required")
	})

	t.Run("CreateUser Missing ID", func(t *testing.T) {
		err := store.CreateUser(ctx, configv1.User_builder{}.Build())
		assert.ErrorContains(t, err, "user ID is required")
	})

	t.Run("UpdateUser Missing ID", func(t *testing.T) {
		err := store.UpdateUser(ctx, configv1.User_builder{}.Build())
		assert.ErrorContains(t, err, "user ID is required")
	})

	t.Run("UpdateUser Not Found", func(t *testing.T) {
		err := store.UpdateUser(ctx, configv1.User_builder{Id: proto.String("non-existent")}.Build())
		assert.ErrorContains(t, err, "user not found")
	})
}

func TestStore_HasConfigSources(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "mcpany-test-store-config-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	dbPath := filepath.Join(tmpDir, "test.db")
	db, err := NewDB(dbPath)
	require.NoError(t, err)
	defer db.Close()

	store := NewStore(db)
	assert.True(t, store.HasConfigSources())
}
