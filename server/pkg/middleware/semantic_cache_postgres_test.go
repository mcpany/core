// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewPostgresVectorStoreWithDB(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		db, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
		require.NoError(t, err)
		defer db.Close()

		mock.ExpectPing()
		mock.ExpectExec("CREATE EXTENSION IF NOT EXISTS vector").WillReturnResult(sqlmock.NewResult(0, 0))
		// Use a loose regex to match the table creation and indices
		mock.ExpectExec("CREATE TABLE IF NOT EXISTS semantic_cache_entries").WillReturnResult(sqlmock.NewResult(0, 0))

		store, err := NewPostgresVectorStoreWithDB(db)
		assert.NoError(t, err)
		assert.NotNil(t, store)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("PingFailure", func(t *testing.T) {
		db, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
		require.NoError(t, err)
		defer db.Close()

		mock.ExpectPing().WillReturnError(assert.AnError)

		store, err := NewPostgresVectorStoreWithDB(db)
		assert.Error(t, err)
		assert.Nil(t, store)
		assert.Contains(t, err.Error(), "failed to ping postgres")
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("ExtensionFailure", func(t *testing.T) {
		db, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
		require.NoError(t, err)
		defer db.Close()

		mock.ExpectPing()
		mock.ExpectExec("CREATE EXTENSION IF NOT EXISTS vector").WillReturnError(assert.AnError)

		store, err := NewPostgresVectorStoreWithDB(db)
		assert.Error(t, err)
		assert.Nil(t, store)
		assert.Contains(t, err.Error(), "failed to create vector extension")
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("SchemaFailure", func(t *testing.T) {
		db, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
		require.NoError(t, err)
		defer db.Close()

		mock.ExpectPing()
		mock.ExpectExec("CREATE EXTENSION IF NOT EXISTS vector").WillReturnResult(sqlmock.NewResult(0, 0))
		mock.ExpectExec("CREATE TABLE IF NOT EXISTS semantic_cache_entries").WillReturnError(assert.AnError)

		store, err := NewPostgresVectorStoreWithDB(db)
		assert.Error(t, err)
		assert.Nil(t, store)
		assert.Contains(t, err.Error(), "failed to create semantic_cache_entries table")
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestPostgresVectorStore_Add(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		store := &PostgresVectorStore{db: db}

		key := "test_tool"
		vec := []float32{1.0, 0.0, 0.0}
		result := map[string]interface{}{"foo": "bar"}
		ttl := 1 * time.Minute

		mock.ExpectExec("INSERT INTO semantic_cache_entries").
			WithArgs(key, "[1,0,0]", sqlmock.AnyArg(), sqlmock.AnyArg()).
			WillReturnResult(sqlmock.NewResult(1, 1))

		err = store.Add(context.Background(), key, vec, result, ttl)
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("ExecError", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		store := &PostgresVectorStore{db: db}

		key := "test_tool"
		vec := []float32{1.0, 0.0, 0.0}
		result := map[string]interface{}{"foo": "bar"}
		ttl := 1 * time.Minute

		mock.ExpectExec("INSERT INTO semantic_cache_entries").
			WillReturnError(errors.New("db error"))

		err = store.Add(context.Background(), key, vec, result, ttl)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to insert entry")
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("MarshalVectorError", func(t *testing.T) {
		// Impossible to fail json.Marshal for []float32
	})

	t.Run("MarshalResultError", func(t *testing.T) {
		db, _, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		store := &PostgresVectorStore{db: db}

		key := "test_tool"
		vec := []float32{1.0, 0.0, 0.0}
		result := make(chan int) // Cannot marshal channel
		ttl := 1 * time.Minute

		err = store.Add(context.Background(), key, vec, result, ttl)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to marshal result")
	})
}

func TestPostgresVectorStore_Search(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		store := &PostgresVectorStore{db: db}

		key := "test_tool"
		query := []float32{1.0, 0.0, 0.0}
		resultObj := map[string]interface{}{"foo": "bar"}
		resultBytes, _ := json.Marshal(resultObj)

		rows := sqlmock.NewRows([]string{"result", "distance"}).
			AddRow(resultBytes, 0.1)

		mock.ExpectQuery("SELECT result, \\(vector <=> \\$1\\) as distance").
			WithArgs("[1,0,0]", key, sqlmock.AnyArg()).
			WillReturnRows(rows)

		res, score, found := store.Search(context.Background(), key, query)
		assert.True(t, found)
		assert.Equal(t, float32(0.9), score)
		assert.Equal(t, resultObj, res)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("NotFound", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		store := &PostgresVectorStore{db: db}

		key := "test_tool"
		query := []float32{1.0, 0.0, 0.0}

		mock.ExpectQuery("SELECT result").
			WithArgs("[1,0,0]", key, sqlmock.AnyArg()).
			WillReturnRows(sqlmock.NewRows([]string{"result", "distance"}))

		res, score, found := store.Search(context.Background(), key, query)
		assert.False(t, found)
		assert.Equal(t, float32(0), score)
		assert.Nil(t, res)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("QueryError", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		store := &PostgresVectorStore{db: db}

		key := "test_tool"
		query := []float32{1.0, 0.0, 0.0}

		mock.ExpectQuery("SELECT result").
			WillReturnError(errors.New("db error"))

		res, score, found := store.Search(context.Background(), key, query)
		assert.False(t, found)
		assert.Equal(t, float32(0), score)
		assert.Nil(t, res)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("UnmarshalError", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		store := &PostgresVectorStore{db: db}

		key := "test_tool"
		query := []float32{1.0, 0.0, 0.0}
		resultBytes := []byte("invalid json")

		rows := sqlmock.NewRows([]string{"result", "distance"}).
			AddRow(resultBytes, 0.1)

		mock.ExpectQuery("SELECT result").
			WithArgs("[1,0,0]", key, sqlmock.AnyArg()).
			WillReturnRows(rows)

		res, score, found := store.Search(context.Background(), key, query)
		assert.False(t, found)
		assert.Equal(t, float32(0), score)
		assert.Nil(t, res)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestPostgresVectorStore_Prune(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	store := &PostgresVectorStore{db: db}

	mock.ExpectExec("DELETE FROM semantic_cache_entries").
		WithArgs(sqlmock.AnyArg(), "test_key").
		WillReturnResult(sqlmock.NewResult(0, 5))

	store.Prune(context.Background(), "test_key")
	assert.NoError(t, mock.ExpectationsWereMet())
}
