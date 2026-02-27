// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package testutil

import (
	"context"
	"testing"

	"github.com/mcpany/core/server/pkg/storage/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSeed(t *testing.T) {
	// Setup DB
	dbPath := t.TempDir() + "/test.db"
	db, err := sqlite.NewDB(dbPath)
	require.NoError(t, err)
	defer db.Close()

	// Run Seed
	err = Seed(context.Background(), db.DB)
	require.NoError(t, err)

	// Verify Data
	// Check User
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM users WHERE id = 'test-user'").Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 1, count)

	// Check Service
	err = db.QueryRow("SELECT COUNT(*) FROM upstream_services WHERE id = 'test-service'").Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 1, count)

	// Check Global Settings
	err = db.QueryRow("SELECT COUNT(*) FROM global_settings WHERE id = 1").Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 1, count)
}
