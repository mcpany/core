// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package sqlite

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func setupTestStore(t *testing.T) (*Store, *DB, string) {
	t.Helper()
	tmpDir, err := os.MkdirTemp("", "mcpany-test-load-*")
	require.NoError(t, err)
	t.Cleanup(func() { os.RemoveAll(tmpDir) })

	dbPath := filepath.Join(tmpDir, "test.db")
	db, err := NewDB(dbPath)
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })

	return NewStore(db), db, dbPath
}

func TestStore_Load_HappyPath(t *testing.T) {
	store, _, _ := setupTestStore(t)
	ctx := context.Background()

	// 1. Prepare Data
	svc := configv1.UpstreamServiceConfig_builder{
		Name: proto.String("svc-1"),
		Id:   proto.String("svc-1"),
	}.Build()
	require.NoError(t, store.SaveService(ctx, svc))

	user := configv1.User_builder{
		Id: proto.String("user-1"),
	}.Build()
	require.NoError(t, store.CreateUser(ctx, user))

	settings := configv1.GlobalSettings_builder{
		LogLevel: configv1.GlobalSettings_LOG_LEVEL_INFO.Enum(),
	}.Build()
	require.NoError(t, store.SaveGlobalSettings(ctx, settings))

	col := configv1.Collection_builder{
		Name: proto.String("col-1"),
	}.Build()
	require.NoError(t, store.SaveServiceCollection(ctx, col))

	prof := configv1.ProfileDefinition_builder{
		Name: proto.String("prof-1"),
	}.Build()
	require.NoError(t, store.SaveProfile(ctx, prof))

	// 2. Load
	config, err := store.Load(ctx)
	require.NoError(t, err)
	require.NotNil(t, config)

	// 3. Verify
	assert.Len(t, config.GetUpstreamServices(), 1)
	assert.Equal(t, "svc-1", config.GetUpstreamServices()[0].GetName())

	assert.Len(t, config.GetUsers(), 1)
	assert.Equal(t, "user-1", config.GetUsers()[0].GetId())

	assert.NotNil(t, config.GetGlobalSettings())
	assert.Equal(t, configv1.GlobalSettings_LOG_LEVEL_INFO, config.GetGlobalSettings().GetLogLevel())

	// Profile should be merged into GlobalSettings
	assert.Len(t, config.GetGlobalSettings().GetProfileDefinitions(), 1)
	assert.Equal(t, "prof-1", config.GetGlobalSettings().GetProfileDefinitions()[0].GetName())

	assert.Len(t, config.GetCollections(), 1)
	assert.Equal(t, "col-1", config.GetCollections()[0].GetName())
}

func TestStore_Load_Concurrency(t *testing.T) {
	// This test aims to trigger any race conditions by having multiple items in each table
	store, _, _ := setupTestStore(t)
	ctx := context.Background()

	for i := 0; i < 100; i++ {
		svc := configv1.UpstreamServiceConfig_builder{
			Name: proto.String(fmt.Sprintf("svc-%d", i)),
			Id:   proto.String(fmt.Sprintf("svc-%d", i)),
		}.Build()
		require.NoError(t, store.SaveService(ctx, svc))

		user := configv1.User_builder{
			Id: proto.String(fmt.Sprintf("user-%d", i)),
		}.Build()
		require.NoError(t, store.CreateUser(ctx, user))
	}

	config, err := store.Load(ctx)
	require.NoError(t, err)
	assert.Len(t, config.GetUpstreamServices(), 100)
	assert.Len(t, config.GetUsers(), 100)
}

func TestStore_Load_ErrorHandling_Services(t *testing.T) {
	store, db, _ := setupTestStore(t)
	ctx := context.Background()

	// Insert invalid JSON to trigger error
	_, err := db.ExecContext(ctx, "INSERT INTO upstream_services (id, name, config_json) VALUES (?, ?, ?)", "bad", "bad", "{invalid")
	require.NoError(t, err)

	config, err := store.Load(ctx)
	assert.Error(t, err)
	assert.Nil(t, config)
	assert.Contains(t, err.Error(), "failed to unmarshal service config")
}

func TestStore_Load_ErrorHandling_Users(t *testing.T) {
	store, db, _ := setupTestStore(t)
	ctx := context.Background()

	_, err := db.ExecContext(ctx, "INSERT INTO users (id, config_json) VALUES (?, ?)", "bad", "{invalid")
	require.NoError(t, err)

	config, err := store.Load(ctx)
	assert.Error(t, err)
	assert.Nil(t, config)
	assert.Contains(t, err.Error(), "failed to unmarshal user config")
}

func TestStore_Load_ErrorHandling_Profiles(t *testing.T) {
	store, db, _ := setupTestStore(t)
	ctx := context.Background()

	_, err := db.ExecContext(ctx, "INSERT INTO profile_definitions (name, config_json) VALUES (?, ?)", "bad", "{invalid")
	require.NoError(t, err)

	config, err := store.Load(ctx)
	assert.Error(t, err)
	assert.Nil(t, config)
	assert.Contains(t, err.Error(), "failed to unmarshal profile config")
}

func TestStore_Load_Collections_SilentFailure(t *testing.T) {
	// This tests the "Dark Matter" behavior: Collections errors are ignored.
	store, db, _ := setupTestStore(t)
	ctx := context.Background()

	// 1. Test Query Error (Table Missing)
	// We can't easily drop the table because NewDB creates it and we can't alter it easily while store is open?
	// Actually we can drop it.
	_, err := db.ExecContext(ctx, "DROP TABLE service_collections")
	require.NoError(t, err)

	config, err := store.Load(ctx)
	// Expectation: NO ERROR, simply empty collections.
	assert.NoError(t, err)
	assert.NotNil(t, config)
	assert.Empty(t, config.GetCollections())

	// 2. Test Invalid JSON
	// Re-create table
	_, err = db.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS service_collections (
		id TEXT PRIMARY KEY,
		name TEXT UNIQUE NOT NULL,
		config_json TEXT NOT NULL,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	)`)
	require.NoError(t, err)

	_, err = db.ExecContext(ctx, "INSERT INTO service_collections (id, name, config_json) VALUES (?, ?, ?)", "bad", "bad", "{invalid")
	require.NoError(t, err)

	config, err = store.Load(ctx)
	// Expectation: NO ERROR, simply empty collections (the invalid one is skipped).
	assert.NoError(t, err)
	assert.NotNil(t, config)
	assert.Empty(t, config.GetCollections())
}

func TestStore_Load_GlobalSettings_SilentFailure(t *testing.T) {
	store, db, _ := setupTestStore(t)
	ctx := context.Background()

	// Insert invalid JSON into global_settings
	// Note: global_settings has a check constraint id=1, so we must insert id=1.
	_, err := db.ExecContext(ctx, "INSERT INTO global_settings (id, config_json) VALUES (1, ?)", "{invalid")
	require.NoError(t, err)

	config, err := store.Load(ctx)
	// Expectation: NO ERROR, simply no global settings loaded (nil or default).
	assert.NoError(t, err)
	assert.NotNil(t, config)
	// If settings fail to load, it should be nil in the config?
	// Let's check store.go:
	// if settings != nil { builder.GlobalSettings = settings }
	// So it will be nil.
	assert.Nil(t, config.GetGlobalSettings())
}

func TestStore_Load_MultipleErrors(t *testing.T) {
	// Test when multiple goroutines fail.
	store, db, _ := setupTestStore(t)
	ctx := context.Background()

	_, err := db.ExecContext(ctx, "INSERT INTO upstream_services (id, name, config_json) VALUES (?, ?, ?)", "bad", "bad", "{invalid")
	require.NoError(t, err)

	_, err = db.ExecContext(ctx, "INSERT INTO users (id, config_json) VALUES (?, ?)", "bad", "{invalid")
	require.NoError(t, err)

	config, err := store.Load(ctx)
	assert.Error(t, err)
	assert.Nil(t, config)
	// We don't guarantee which error is returned, but it should be one of them.
	assert.Contains(t, err.Error(), "failed to unmarshal")
}
