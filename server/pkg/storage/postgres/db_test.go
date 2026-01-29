package postgres

import (
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewDBFromSQLDB(t *testing.T) {
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

func TestNewDB_Error(t *testing.T) {
	// Without a running postgres, this should fail
	_, err := NewDB("postgres://invalid:invalid@127.0.0.1:5432/invalid?sslmode=disable")
	require.Error(t, err)
}
