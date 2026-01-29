// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"context"
	"os"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/storage/sqlite"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSeeding_OfficialCollections(t *testing.T) {
	// Setup real SQLite DB
	dbPath := "test_seed.db"
	// Ensure cleanup
	_ = os.Remove(dbPath)
	defer os.Remove(dbPath)

	db, err := sqlite.NewDB(dbPath)
	require.NoError(t, err)
	defer db.Close()

	store := sqlite.NewStore(db)

	app := NewApplication()
	app.Storage = store
	app.fs = afero.NewMemMapFs()

	// Initialize (this creates tables usually? NewDB does migrate)
	err = app.initializeDatabase(context.Background(), store)
	require.NoError(t, err)

	// Run seeding explicitly
	err = app.initializeOfficialCollections(context.Background(), store)
	require.NoError(t, err)

	// Verify
	collections, err := store.ListServiceCollections(context.Background())
	require.NoError(t, err)

	// Should have at least 2
	assert.GreaterOrEqual(t, len(collections), 2)

	var dataStack *configv1.Collection
	for _, c := range collections {
		if c.GetName() == "Data Engineering Stack" {
			dataStack = c
			break
		}
	}

	require.NotNil(t, dataStack, "Data Engineering Stack should exist")
	assert.Equal(t, "MCP Any Team", dataStack.GetAuthor())
	require.NotEmpty(t, dataStack.GetServices())

	svc := dataStack.GetServices()[0]
	assert.Equal(t, "sqlite-db", svc.GetId())

	// Verify Command
	cmdSvc := svc.GetCommandLineService()
	assert.Equal(t, "npx -y @modelcontextprotocol/server-sqlite", cmdSvc.GetCommand())
}
