// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package postgres

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"
)

// TestStore_Load tests the Load method of the PostgreSQL store.
func TestStore_Load(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	store := &Store{db: &DB{DB: db}}

	t.Run("Query Error", func(t *testing.T) {
		// Because store.Load calls loadUsers, loadSettings, etc.
		// A query error in the first call (loadUsers) will be returned.
		mock.ExpectQuery(".*").WillReturnError(context.DeadlineExceeded)

		cfg, err := store.Load(context.Background())
		require.Error(t, err)
		require.Nil(t, cfg)
		require.ErrorIs(t, err, context.DeadlineExceeded)
	})
}
