package middleware

import (
	"context"
	"encoding/json"
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
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	// Bypass NewPostgresVectorStore to inject mock DB
	store := &PostgresVectorStore{db: db}

	key := "test_tool"
	vec := []float32{1.0, 0.0, 0.0}
	result := map[string]interface{}{"foo": "bar"}
	ttl := 1 * time.Minute

	// Expectation for Add
	mock.ExpectExec("INSERT INTO semantic_cache_entries").
		WithArgs(key, "[1,0,0]", sqlmock.AnyArg(), sqlmock.AnyArg()). // vector as string, result as bytes, time
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = store.Add(context.Background(), key, vec, result, ttl)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgresVectorStore_Search(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	store := &PostgresVectorStore{db: db}

	key := "test_tool"
	query := []float32{1.0, 0.0, 0.0}
	resultObj := map[string]interface{}{"foo": "bar"}
	resultBytes, _ := json.Marshal(resultObj)

	// Expectation for Search
	rows := sqlmock.NewRows([]string{"result", "distance"}).
		AddRow(resultBytes, 0.1) // distance 0.1 -> similarity 0.9

	mock.ExpectQuery("SELECT result, \\(vector <=> \\$1\\) as distance").
		WithArgs("[1,0,0]", key, sqlmock.AnyArg()).
		WillReturnRows(rows)

	res, score, found := store.Search(context.Background(), key, query)
	assert.True(t, found)
	assert.Equal(t, float32(0.9), score)
	assert.Equal(t, resultObj, res)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgresVectorStore_Search_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	store := &PostgresVectorStore{db: db}

	key := "test_tool"
	query := []float32{1.0, 0.0, 0.0}

	// Remove the first expectation which was confusing the matcher logic due to sequential calls in the test (which wasn't sequential in the code, but maybe I copy pasted wrong).
	// Ah, I had two mock.ExpectQuery calls but only one Search call in the test.
	// The first one `WillReturnError(nil)` is weird for Search which calls QueryRow.
	// If QueryRow fails, it returns error. If it succeeds but no rows, it returns sql.ErrNoRows.
	// sqlmock handles QueryRow by expecting Query.
	// We just want to simulate "no rows found".

	mock.ExpectQuery("SELECT result").
		WithArgs("[1,0,0]", key, sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"result", "distance"})) // Empty rows

	res, score, found := store.Search(context.Background(), key, query)
	assert.False(t, found)
	assert.Equal(t, float32(0), score)
	assert.Nil(t, res)
	assert.NoError(t, mock.ExpectationsWereMet())
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
