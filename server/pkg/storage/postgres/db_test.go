// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package postgres

import (
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewDBFromSQLDB ...
// Summary: TestNewDBFromSQLDB
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	// Expect Ping
	mock.ExpectPing()

	// Expect Schema Init
	mock.ExpectExec("CREATE TABLE IF NOT EXISTS upstream_services").
		WillReturnResult(sqlmock.NewResult(0, 0))

	pgDB, err := NewDBFromSQLDB(db)
	require.NoError(t, err)
	assert.NotNil(t, pgDB)

	require.NoError(t, mock.ExpectationsWereMet())
}

// TestNewDB_Error ...
// Summary: TestNewDB_Error
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	// Without a running postgres, this should fail
	_, err := NewDB("postgres://invalid:invalid@127.0.0.1:5432/invalid?sslmode=disable")
	require.Error(t, err)
}
