package postgres

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStore_Load(t *testing.T) {

	t.Run("Load_Success", func(t *testing.T) {
	    db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	    require.NoError(t, err)
	    defer db.Close()

        mock.ExpectExec("CREATE TABLE IF NOT EXISTS upstream_services ( id TEXT PRIMARY KEY, name TEXT UNIQUE NOT NULL, config_json TEXT NOT NULL, created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP, updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP ); CREATE TABLE IF NOT EXISTS users ( id TEXT PRIMARY KEY, config_json TEXT NOT NULL, created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP, updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP ); CREATE TABLE IF NOT EXISTS global_settings ( id INTEGER PRIMARY KEY CHECK (id = 1), config_json TEXT NOT NULL, updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP ); CREATE TABLE IF NOT EXISTS secrets ( id TEXT PRIMARY KEY, name TEXT NOT NULL, key TEXT NOT NULL, config_json TEXT NOT NULL, created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP, updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP ); CREATE TABLE IF NOT EXISTS profile_definitions ( id TEXT PRIMARY KEY, name TEXT UNIQUE NOT NULL, config_json TEXT NOT NULL, created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP, updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP ); CREATE TABLE IF NOT EXISTS service_collections ( id TEXT PRIMARY KEY, name TEXT UNIQUE NOT NULL, config_json TEXT NOT NULL, created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP, updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP ); CREATE TABLE IF NOT EXISTS user_tokens ( user_id TEXT NOT NULL, service_id TEXT NOT NULL, config_json TEXT NOT NULL, updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP, PRIMARY KEY (user_id, service_id) ); CREATE TABLE IF NOT EXISTS service_templates ( id TEXT PRIMARY KEY, name TEXT NOT NULL, config_json TEXT NOT NULL, created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP, updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP ); CREATE TABLE IF NOT EXISTS logs ( id TEXT PRIMARY KEY, timestamp TIMESTAMPTZ NOT NULL, level TEXT NOT NULL, source TEXT, message TEXT NOT NULL, metadata_json TEXT, created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP ); CREATE INDEX IF NOT EXISTS idx_logs_timestamp ON logs(timestamp);").WillReturnResult(sqlmock.NewResult(0, 0))

	    pgDB, err := NewDBFromSQLDB(db)
	    require.NoError(t, err)
	    store := NewStore(pgDB)

        mock.MatchExpectationsInOrder(false)
		mock.ExpectQuery("SELECT config_json FROM upstream_services").WillReturnRows(sqlmock.NewRows([]string{"config_json"}).AddRow(`{"name":"service1"}`).AddRow(`{"name":"service2"}`))
		mock.ExpectQuery("SELECT config_json FROM users").WillReturnRows(sqlmock.NewRows([]string{"config_json"}).AddRow(`{"id":"user1"}`).AddRow(`{"id":"user2"}`))
		mock.ExpectQuery("SELECT config_json FROM global_settings WHERE id = 1").WillReturnRows(sqlmock.NewRows([]string{"config_json"}).AddRow(`{}`))
		mock.ExpectQuery("SELECT config_json FROM service_collections").WillReturnRows(sqlmock.NewRows([]string{"config_json"}).AddRow(`{"name":"collection1"}`))
		mock.ExpectQuery("SELECT config_json FROM profile_definitions").WillReturnRows(sqlmock.NewRows([]string{"config_json"}).AddRow(`{"name":"profile1"}`))

		ctx := context.Background()
		config, err := store.Load(ctx)
		require.NoError(t, err)
		require.NotNil(t, config)

		assert.Len(t, config.GetUpstreamServices(), 2)
		assert.Len(t, config.GetUsers(), 2)
		assert.NotNil(t, config.GetGlobalSettings())
		assert.Len(t, config.GetCollections(), 1)
	})

	t.Run("Load_ErrorInServices", func(t *testing.T) {
	    db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	    require.NoError(t, err)
	    defer db.Close()
        mock.ExpectExec("CREATE TABLE IF NOT EXISTS upstream_services ( id TEXT PRIMARY KEY, name TEXT UNIQUE NOT NULL, config_json TEXT NOT NULL, created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP, updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP ); CREATE TABLE IF NOT EXISTS users ( id TEXT PRIMARY KEY, config_json TEXT NOT NULL, created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP, updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP ); CREATE TABLE IF NOT EXISTS global_settings ( id INTEGER PRIMARY KEY CHECK (id = 1), config_json TEXT NOT NULL, updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP ); CREATE TABLE IF NOT EXISTS secrets ( id TEXT PRIMARY KEY, name TEXT NOT NULL, key TEXT NOT NULL, config_json TEXT NOT NULL, created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP, updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP ); CREATE TABLE IF NOT EXISTS profile_definitions ( id TEXT PRIMARY KEY, name TEXT UNIQUE NOT NULL, config_json TEXT NOT NULL, created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP, updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP ); CREATE TABLE IF NOT EXISTS service_collections ( id TEXT PRIMARY KEY, name TEXT UNIQUE NOT NULL, config_json TEXT NOT NULL, created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP, updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP ); CREATE TABLE IF NOT EXISTS user_tokens ( user_id TEXT NOT NULL, service_id TEXT NOT NULL, config_json TEXT NOT NULL, updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP, PRIMARY KEY (user_id, service_id) ); CREATE TABLE IF NOT EXISTS service_templates ( id TEXT PRIMARY KEY, name TEXT NOT NULL, config_json TEXT NOT NULL, created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP, updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP ); CREATE TABLE IF NOT EXISTS logs ( id TEXT PRIMARY KEY, timestamp TIMESTAMPTZ NOT NULL, level TEXT NOT NULL, source TEXT, message TEXT NOT NULL, metadata_json TEXT, created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP ); CREATE INDEX IF NOT EXISTS idx_logs_timestamp ON logs(timestamp);").WillReturnResult(sqlmock.NewResult(0, 0))
	    pgDB, err := NewDBFromSQLDB(db)
	    require.NoError(t, err)
	    store := NewStore(pgDB)

        mock.MatchExpectationsInOrder(false)
		mock.ExpectQuery("SELECT config_json FROM upstream_services").WillReturnError(errors.New("db error"))
		mock.ExpectQuery("SELECT config_json FROM users").WillReturnRows(sqlmock.NewRows([]string{"config_json"}).AddRow(`{"id":"user1"}`))
		mock.ExpectQuery("SELECT config_json FROM global_settings WHERE id = 1").WillReturnRows(sqlmock.NewRows([]string{"config_json"}).AddRow(`{}`))
		mock.ExpectQuery("SELECT config_json FROM service_collections").WillReturnRows(sqlmock.NewRows([]string{"config_json"}).AddRow(`{"name":"collection1"}`))
		mock.ExpectQuery("SELECT config_json FROM profile_definitions").WillReturnRows(sqlmock.NewRows([]string{"config_json"}).AddRow(`{"name":"profile1"}`))

		ctx := context.Background()
		config, err := store.Load(ctx)
		require.Error(t, err)
		assert.Nil(t, config)
		assert.Contains(t, err.Error(), "failed to query upstream_services: db error")
	})

	t.Run("Load_SettingsNoRows", func(t *testing.T) {
	    db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	    require.NoError(t, err)
	    defer db.Close()
        mock.ExpectExec("CREATE TABLE IF NOT EXISTS upstream_services ( id TEXT PRIMARY KEY, name TEXT UNIQUE NOT NULL, config_json TEXT NOT NULL, created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP, updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP ); CREATE TABLE IF NOT EXISTS users ( id TEXT PRIMARY KEY, config_json TEXT NOT NULL, created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP, updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP ); CREATE TABLE IF NOT EXISTS global_settings ( id INTEGER PRIMARY KEY CHECK (id = 1), config_json TEXT NOT NULL, updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP ); CREATE TABLE IF NOT EXISTS secrets ( id TEXT PRIMARY KEY, name TEXT NOT NULL, key TEXT NOT NULL, config_json TEXT NOT NULL, created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP, updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP ); CREATE TABLE IF NOT EXISTS profile_definitions ( id TEXT PRIMARY KEY, name TEXT UNIQUE NOT NULL, config_json TEXT NOT NULL, created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP, updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP ); CREATE TABLE IF NOT EXISTS service_collections ( id TEXT PRIMARY KEY, name TEXT UNIQUE NOT NULL, config_json TEXT NOT NULL, created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP, updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP ); CREATE TABLE IF NOT EXISTS user_tokens ( user_id TEXT NOT NULL, service_id TEXT NOT NULL, config_json TEXT NOT NULL, updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP, PRIMARY KEY (user_id, service_id) ); CREATE TABLE IF NOT EXISTS service_templates ( id TEXT PRIMARY KEY, name TEXT NOT NULL, config_json TEXT NOT NULL, created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP, updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP ); CREATE TABLE IF NOT EXISTS logs ( id TEXT PRIMARY KEY, timestamp TIMESTAMPTZ NOT NULL, level TEXT NOT NULL, source TEXT, message TEXT NOT NULL, metadata_json TEXT, created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP ); CREATE INDEX IF NOT EXISTS idx_logs_timestamp ON logs(timestamp);").WillReturnResult(sqlmock.NewResult(0, 0))
	    pgDB, err := NewDBFromSQLDB(db)
	    require.NoError(t, err)
	    store := NewStore(pgDB)

        mock.MatchExpectationsInOrder(false)
		mock.ExpectQuery("SELECT config_json FROM upstream_services").WillReturnRows(sqlmock.NewRows([]string{"config_json"}).AddRow(`{"name":"service1"}`))
		mock.ExpectQuery("SELECT config_json FROM users").WillReturnRows(sqlmock.NewRows([]string{"config_json"}).AddRow(`{"id":"user1"}`))
		mock.ExpectQuery("SELECT config_json FROM global_settings WHERE id = 1").WillReturnError(sql.ErrNoRows)
		mock.ExpectQuery("SELECT config_json FROM service_collections").WillReturnRows(sqlmock.NewRows([]string{"config_json"}).AddRow(`{"name":"collection1"}`))
		mock.ExpectQuery("SELECT config_json FROM profile_definitions").WillReturnRows(sqlmock.NewRows([]string{"config_json"})) // Empty profiles to test settings = nil

		ctx := context.Background()
		config, err := store.Load(ctx)
		require.NoError(t, err)
		require.NotNil(t, config)
		assert.Nil(t, config.GetGlobalSettings())
	})

	t.Run("Load_InvalidJSON", func(t *testing.T) {
	    db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	    require.NoError(t, err)
	    defer db.Close()
        mock.ExpectExec("CREATE TABLE IF NOT EXISTS upstream_services ( id TEXT PRIMARY KEY, name TEXT UNIQUE NOT NULL, config_json TEXT NOT NULL, created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP, updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP ); CREATE TABLE IF NOT EXISTS users ( id TEXT PRIMARY KEY, config_json TEXT NOT NULL, created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP, updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP ); CREATE TABLE IF NOT EXISTS global_settings ( id INTEGER PRIMARY KEY CHECK (id = 1), config_json TEXT NOT NULL, updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP ); CREATE TABLE IF NOT EXISTS secrets ( id TEXT PRIMARY KEY, name TEXT NOT NULL, key TEXT NOT NULL, config_json TEXT NOT NULL, created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP, updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP ); CREATE TABLE IF NOT EXISTS profile_definitions ( id TEXT PRIMARY KEY, name TEXT UNIQUE NOT NULL, config_json TEXT NOT NULL, created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP, updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP ); CREATE TABLE IF NOT EXISTS service_collections ( id TEXT PRIMARY KEY, name TEXT UNIQUE NOT NULL, config_json TEXT NOT NULL, created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP, updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP ); CREATE TABLE IF NOT EXISTS user_tokens ( user_id TEXT NOT NULL, service_id TEXT NOT NULL, config_json TEXT NOT NULL, updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP, PRIMARY KEY (user_id, service_id) ); CREATE TABLE IF NOT EXISTS service_templates ( id TEXT PRIMARY KEY, name TEXT NOT NULL, config_json TEXT NOT NULL, created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP, updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP ); CREATE TABLE IF NOT EXISTS logs ( id TEXT PRIMARY KEY, timestamp TIMESTAMPTZ NOT NULL, level TEXT NOT NULL, source TEXT, message TEXT NOT NULL, metadata_json TEXT, created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP ); CREATE INDEX IF NOT EXISTS idx_logs_timestamp ON logs(timestamp);").WillReturnResult(sqlmock.NewResult(0, 0))
	    pgDB, err := NewDBFromSQLDB(db)
	    require.NoError(t, err)
	    store := NewStore(pgDB)

        mock.MatchExpectationsInOrder(false)
		mock.ExpectQuery("SELECT config_json FROM upstream_services").WillReturnRows(sqlmock.NewRows([]string{"config_json"}).AddRow(`{"name":"service1"`))
		mock.ExpectQuery("SELECT config_json FROM users").WillReturnRows(sqlmock.NewRows([]string{"config_json"}).AddRow(`{"id":"user1"}`))
		mock.ExpectQuery("SELECT config_json FROM global_settings WHERE id = 1").WillReturnRows(sqlmock.NewRows([]string{"config_json"}).AddRow(`{}`))
		mock.ExpectQuery("SELECT config_json FROM service_collections").WillReturnRows(sqlmock.NewRows([]string{"config_json"}).AddRow(`{"name":"collection1"}`))
		mock.ExpectQuery("SELECT config_json FROM profile_definitions").WillReturnRows(sqlmock.NewRows([]string{"config_json"}).AddRow(`{"name":"profile1"}`))

		ctx := context.Background()
		config, err := store.Load(ctx)
		require.Error(t, err)
		assert.Nil(t, config)
		assert.Contains(t, err.Error(), "failed to unmarshal service config")
	})

	t.Run("Load_ErrorInUsers", func(t *testing.T) {
	    db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	    require.NoError(t, err)
	    defer db.Close()
        mock.ExpectExec("CREATE TABLE IF NOT EXISTS upstream_services ( id TEXT PRIMARY KEY, name TEXT UNIQUE NOT NULL, config_json TEXT NOT NULL, created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP, updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP ); CREATE TABLE IF NOT EXISTS users ( id TEXT PRIMARY KEY, config_json TEXT NOT NULL, created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP, updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP ); CREATE TABLE IF NOT EXISTS global_settings ( id INTEGER PRIMARY KEY CHECK (id = 1), config_json TEXT NOT NULL, updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP ); CREATE TABLE IF NOT EXISTS secrets ( id TEXT PRIMARY KEY, name TEXT NOT NULL, key TEXT NOT NULL, config_json TEXT NOT NULL, created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP, updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP ); CREATE TABLE IF NOT EXISTS profile_definitions ( id TEXT PRIMARY KEY, name TEXT UNIQUE NOT NULL, config_json TEXT NOT NULL, created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP, updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP ); CREATE TABLE IF NOT EXISTS service_collections ( id TEXT PRIMARY KEY, name TEXT UNIQUE NOT NULL, config_json TEXT NOT NULL, created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP, updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP ); CREATE TABLE IF NOT EXISTS user_tokens ( user_id TEXT NOT NULL, service_id TEXT NOT NULL, config_json TEXT NOT NULL, updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP, PRIMARY KEY (user_id, service_id) ); CREATE TABLE IF NOT EXISTS service_templates ( id TEXT PRIMARY KEY, name TEXT NOT NULL, config_json TEXT NOT NULL, created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP, updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP ); CREATE TABLE IF NOT EXISTS logs ( id TEXT PRIMARY KEY, timestamp TIMESTAMPTZ NOT NULL, level TEXT NOT NULL, source TEXT, message TEXT NOT NULL, metadata_json TEXT, created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP ); CREATE INDEX IF NOT EXISTS idx_logs_timestamp ON logs(timestamp);").WillReturnResult(sqlmock.NewResult(0, 0))
	    pgDB, err := NewDBFromSQLDB(db)
	    require.NoError(t, err)
	    store := NewStore(pgDB)

        mock.MatchExpectationsInOrder(false)
		mock.ExpectQuery("SELECT config_json FROM upstream_services").WillReturnRows(sqlmock.NewRows([]string{"config_json"}).AddRow(`{"name":"service1"}`))
		mock.ExpectQuery("SELECT config_json FROM users").WillReturnError(errors.New("db error users"))
		mock.ExpectQuery("SELECT config_json FROM global_settings WHERE id = 1").WillReturnRows(sqlmock.NewRows([]string{"config_json"}).AddRow(`{}`))
		mock.ExpectQuery("SELECT config_json FROM service_collections").WillReturnRows(sqlmock.NewRows([]string{"config_json"}).AddRow(`{"name":"collection1"}`))
		mock.ExpectQuery("SELECT config_json FROM profile_definitions").WillReturnRows(sqlmock.NewRows([]string{"config_json"}).AddRow(`{"name":"profile1"}`))

		ctx := context.Background()
		config, err := store.Load(ctx)
		require.Error(t, err)
		assert.Nil(t, config)
		assert.Contains(t, err.Error(), "failed to query users: db error users")
	})

	t.Run("Load_ErrorInProfiles", func(t *testing.T) {
	    db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	    require.NoError(t, err)
	    defer db.Close()
        mock.ExpectExec("CREATE TABLE IF NOT EXISTS upstream_services ( id TEXT PRIMARY KEY, name TEXT UNIQUE NOT NULL, config_json TEXT NOT NULL, created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP, updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP ); CREATE TABLE IF NOT EXISTS users ( id TEXT PRIMARY KEY, config_json TEXT NOT NULL, created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP, updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP ); CREATE TABLE IF NOT EXISTS global_settings ( id INTEGER PRIMARY KEY CHECK (id = 1), config_json TEXT NOT NULL, updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP ); CREATE TABLE IF NOT EXISTS secrets ( id TEXT PRIMARY KEY, name TEXT NOT NULL, key TEXT NOT NULL, config_json TEXT NOT NULL, created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP, updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP ); CREATE TABLE IF NOT EXISTS profile_definitions ( id TEXT PRIMARY KEY, name TEXT UNIQUE NOT NULL, config_json TEXT NOT NULL, created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP, updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP ); CREATE TABLE IF NOT EXISTS service_collections ( id TEXT PRIMARY KEY, name TEXT UNIQUE NOT NULL, config_json TEXT NOT NULL, created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP, updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP ); CREATE TABLE IF NOT EXISTS user_tokens ( user_id TEXT NOT NULL, service_id TEXT NOT NULL, config_json TEXT NOT NULL, updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP, PRIMARY KEY (user_id, service_id) ); CREATE TABLE IF NOT EXISTS service_templates ( id TEXT PRIMARY KEY, name TEXT NOT NULL, config_json TEXT NOT NULL, created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP, updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP ); CREATE TABLE IF NOT EXISTS logs ( id TEXT PRIMARY KEY, timestamp TIMESTAMPTZ NOT NULL, level TEXT NOT NULL, source TEXT, message TEXT NOT NULL, metadata_json TEXT, created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP ); CREATE INDEX IF NOT EXISTS idx_logs_timestamp ON logs(timestamp);").WillReturnResult(sqlmock.NewResult(0, 0))
	    pgDB, err := NewDBFromSQLDB(db)
	    require.NoError(t, err)
	    store := NewStore(pgDB)

        mock.MatchExpectationsInOrder(false)
		mock.ExpectQuery("SELECT config_json FROM upstream_services").WillReturnRows(sqlmock.NewRows([]string{"config_json"}).AddRow(`{"name":"service1"}`))
		mock.ExpectQuery("SELECT config_json FROM users").WillReturnRows(sqlmock.NewRows([]string{"config_json"}).AddRow(`{"id":"user1"}`))
		mock.ExpectQuery("SELECT config_json FROM global_settings WHERE id = 1").WillReturnRows(sqlmock.NewRows([]string{"config_json"}).AddRow(`{}`))
		mock.ExpectQuery("SELECT config_json FROM service_collections").WillReturnRows(sqlmock.NewRows([]string{"config_json"}).AddRow(`{"name":"collection1"}`))
		mock.ExpectQuery("SELECT config_json FROM profile_definitions").WillReturnError(errors.New("db error profiles"))

		ctx := context.Background()
		config, err := store.Load(ctx)
		require.Error(t, err)
		assert.Nil(t, config)
		assert.Contains(t, err.Error(), "failed to query profile_definitions: db error profiles")
	})
}

func TestStore_Load_ErrorInCollections(t *testing.T) {
	    db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	    require.NoError(t, err)
	    defer db.Close()
        mock.ExpectExec("CREATE TABLE IF NOT EXISTS upstream_services ( id TEXT PRIMARY KEY, name TEXT UNIQUE NOT NULL, config_json TEXT NOT NULL, created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP, updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP ); CREATE TABLE IF NOT EXISTS users ( id TEXT PRIMARY KEY, config_json TEXT NOT NULL, created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP, updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP ); CREATE TABLE IF NOT EXISTS global_settings ( id INTEGER PRIMARY KEY CHECK (id = 1), config_json TEXT NOT NULL, updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP ); CREATE TABLE IF NOT EXISTS secrets ( id TEXT PRIMARY KEY, name TEXT NOT NULL, key TEXT NOT NULL, config_json TEXT NOT NULL, created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP, updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP ); CREATE TABLE IF NOT EXISTS profile_definitions ( id TEXT PRIMARY KEY, name TEXT UNIQUE NOT NULL, config_json TEXT NOT NULL, created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP, updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP ); CREATE TABLE IF NOT EXISTS service_collections ( id TEXT PRIMARY KEY, name TEXT UNIQUE NOT NULL, config_json TEXT NOT NULL, created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP, updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP ); CREATE TABLE IF NOT EXISTS user_tokens ( user_id TEXT NOT NULL, service_id TEXT NOT NULL, config_json TEXT NOT NULL, updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP, PRIMARY KEY (user_id, service_id) ); CREATE TABLE IF NOT EXISTS service_templates ( id TEXT PRIMARY KEY, name TEXT NOT NULL, config_json TEXT NOT NULL, created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP, updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP ); CREATE TABLE IF NOT EXISTS logs ( id TEXT PRIMARY KEY, timestamp TIMESTAMPTZ NOT NULL, level TEXT NOT NULL, source TEXT, message TEXT NOT NULL, metadata_json TEXT, created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP ); CREATE INDEX IF NOT EXISTS idx_logs_timestamp ON logs(timestamp);").WillReturnResult(sqlmock.NewResult(0, 0))
	    pgDB, err := NewDBFromSQLDB(db)
	    require.NoError(t, err)
	    store := NewStore(pgDB)

        mock.MatchExpectationsInOrder(false)
		mock.ExpectQuery("SELECT config_json FROM upstream_services").WillReturnRows(sqlmock.NewRows([]string{"config_json"}).AddRow(`{"name":"service1"}`))
		mock.ExpectQuery("SELECT config_json FROM users").WillReturnRows(sqlmock.NewRows([]string{"config_json"}).AddRow(`{"id":"user1"}`))
		mock.ExpectQuery("SELECT config_json FROM global_settings WHERE id = 1").WillReturnRows(sqlmock.NewRows([]string{"config_json"}).AddRow(`{}`))
		mock.ExpectQuery("SELECT config_json FROM service_collections").WillReturnError(errors.New("db error collections"))
		mock.ExpectQuery("SELECT config_json FROM profile_definitions").WillReturnRows(sqlmock.NewRows([]string{"config_json"}).AddRow(`{"name":"profile1"}`))

		ctx := context.Background()
		config, err := store.Load(ctx)
		require.NoError(t, err)
		require.NotNil(t, config)
}

func TestStore_Load_GlobalSettingsScanError(t *testing.T) {
	    db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	    require.NoError(t, err)
	    defer db.Close()
        mock.ExpectExec("CREATE TABLE IF NOT EXISTS upstream_services ( id TEXT PRIMARY KEY, name TEXT UNIQUE NOT NULL, config_json TEXT NOT NULL, created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP, updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP ); CREATE TABLE IF NOT EXISTS users ( id TEXT PRIMARY KEY, config_json TEXT NOT NULL, created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP, updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP ); CREATE TABLE IF NOT EXISTS global_settings ( id INTEGER PRIMARY KEY CHECK (id = 1), config_json TEXT NOT NULL, updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP ); CREATE TABLE IF NOT EXISTS secrets ( id TEXT PRIMARY KEY, name TEXT NOT NULL, key TEXT NOT NULL, config_json TEXT NOT NULL, created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP, updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP ); CREATE TABLE IF NOT EXISTS profile_definitions ( id TEXT PRIMARY KEY, name TEXT UNIQUE NOT NULL, config_json TEXT NOT NULL, created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP, updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP ); CREATE TABLE IF NOT EXISTS service_collections ( id TEXT PRIMARY KEY, name TEXT UNIQUE NOT NULL, config_json TEXT NOT NULL, created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP, updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP ); CREATE TABLE IF NOT EXISTS user_tokens ( user_id TEXT NOT NULL, service_id TEXT NOT NULL, config_json TEXT NOT NULL, updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP, PRIMARY KEY (user_id, service_id) ); CREATE TABLE IF NOT EXISTS service_templates ( id TEXT PRIMARY KEY, name TEXT NOT NULL, config_json TEXT NOT NULL, created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP, updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP ); CREATE TABLE IF NOT EXISTS logs ( id TEXT PRIMARY KEY, timestamp TIMESTAMPTZ NOT NULL, level TEXT NOT NULL, source TEXT, message TEXT NOT NULL, metadata_json TEXT, created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP ); CREATE INDEX IF NOT EXISTS idx_logs_timestamp ON logs(timestamp);").WillReturnResult(sqlmock.NewResult(0, 0))
	    pgDB, err := NewDBFromSQLDB(db)
	    require.NoError(t, err)
	    store := NewStore(pgDB)

        mock.MatchExpectationsInOrder(false)
		mock.ExpectQuery("SELECT config_json FROM upstream_services").WillReturnRows(sqlmock.NewRows([]string{"config_json"}).AddRow(`{"name":"service1"}`))
		mock.ExpectQuery("SELECT config_json FROM users").WillReturnRows(sqlmock.NewRows([]string{"config_json"}).AddRow(`{"id":"user1"}`))
		mock.ExpectQuery("SELECT config_json FROM global_settings WHERE id = 1").WillReturnRows(sqlmock.NewRows([]string{"config_json", "extra"}).AddRow(`{}`, "bad")) // Forces scan error
		mock.ExpectQuery("SELECT config_json FROM service_collections").WillReturnRows(sqlmock.NewRows([]string{"config_json"}).AddRow(`{"name":"collection1"}`))
		mock.ExpectQuery("SELECT config_json FROM profile_definitions").WillReturnRows(sqlmock.NewRows([]string{"config_json"}).AddRow(`{"name":"profile1"}`))

		ctx := context.Background()
		config, err := store.Load(ctx)
		require.NoError(t, err)
		require.NotNil(t, config)
		assert.NotNil(t, config.GetGlobalSettings())
}

func TestStore_Load_ScanErrors(t *testing.T) {
	    db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	    require.NoError(t, err)
	    defer db.Close()
        mock.ExpectExec("CREATE TABLE IF NOT EXISTS upstream_services ( id TEXT PRIMARY KEY, name TEXT UNIQUE NOT NULL, config_json TEXT NOT NULL, created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP, updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP ); CREATE TABLE IF NOT EXISTS users ( id TEXT PRIMARY KEY, config_json TEXT NOT NULL, created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP, updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP ); CREATE TABLE IF NOT EXISTS global_settings ( id INTEGER PRIMARY KEY CHECK (id = 1), config_json TEXT NOT NULL, updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP ); CREATE TABLE IF NOT EXISTS secrets ( id TEXT PRIMARY KEY, name TEXT NOT NULL, key TEXT NOT NULL, config_json TEXT NOT NULL, created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP, updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP ); CREATE TABLE IF NOT EXISTS profile_definitions ( id TEXT PRIMARY KEY, name TEXT UNIQUE NOT NULL, config_json TEXT NOT NULL, created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP, updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP ); CREATE TABLE IF NOT EXISTS service_collections ( id TEXT PRIMARY KEY, name TEXT UNIQUE NOT NULL, config_json TEXT NOT NULL, created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP, updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP ); CREATE TABLE IF NOT EXISTS user_tokens ( user_id TEXT NOT NULL, service_id TEXT NOT NULL, config_json TEXT NOT NULL, updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP, PRIMARY KEY (user_id, service_id) ); CREATE TABLE IF NOT EXISTS service_templates ( id TEXT PRIMARY KEY, name TEXT NOT NULL, config_json TEXT NOT NULL, created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP, updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP ); CREATE TABLE IF NOT EXISTS logs ( id TEXT PRIMARY KEY, timestamp TIMESTAMPTZ NOT NULL, level TEXT NOT NULL, source TEXT, message TEXT NOT NULL, metadata_json TEXT, created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP ); CREATE INDEX IF NOT EXISTS idx_logs_timestamp ON logs(timestamp);").WillReturnResult(sqlmock.NewResult(0, 0))
	    pgDB, err := NewDBFromSQLDB(db)
	    require.NoError(t, err)
	    store := NewStore(pgDB)

        mock.MatchExpectationsInOrder(false)
		mock.ExpectQuery("SELECT config_json FROM upstream_services").WillReturnRows(sqlmock.NewRows([]string{"config_json", "extra"}).AddRow(`{"name":"service1"}`, "bad")) // Force scan error
		mock.ExpectQuery("SELECT config_json FROM users").WillReturnRows(sqlmock.NewRows([]string{"config_json"}).AddRow(`{"id":"user1"}`))
		mock.ExpectQuery("SELECT config_json FROM global_settings WHERE id = 1").WillReturnRows(sqlmock.NewRows([]string{"config_json"}).AddRow(`{}`))
		mock.ExpectQuery("SELECT config_json FROM service_collections").WillReturnRows(sqlmock.NewRows([]string{"config_json"}).AddRow(`{"name":"collection1"}`))
		mock.ExpectQuery("SELECT config_json FROM profile_definitions").WillReturnRows(sqlmock.NewRows([]string{"config_json"}).AddRow(`{"name":"profile1"}`))

		ctx := context.Background()
		config, err := store.Load(ctx)
		require.Error(t, err)
		assert.Nil(t, config)
		assert.Contains(t, err.Error(), "failed to scan config_json")
}

func TestStore_Load_ErrorRows(t *testing.T) {
	    db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	    require.NoError(t, err)
	    defer db.Close()
        mock.ExpectExec("CREATE TABLE IF NOT EXISTS upstream_services ( id TEXT PRIMARY KEY, name TEXT UNIQUE NOT NULL, config_json TEXT NOT NULL, created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP, updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP ); CREATE TABLE IF NOT EXISTS users ( id TEXT PRIMARY KEY, config_json TEXT NOT NULL, created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP, updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP ); CREATE TABLE IF NOT EXISTS global_settings ( id INTEGER PRIMARY KEY CHECK (id = 1), config_json TEXT NOT NULL, updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP ); CREATE TABLE IF NOT EXISTS secrets ( id TEXT PRIMARY KEY, name TEXT NOT NULL, key TEXT NOT NULL, config_json TEXT NOT NULL, created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP, updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP ); CREATE TABLE IF NOT EXISTS profile_definitions ( id TEXT PRIMARY KEY, name TEXT UNIQUE NOT NULL, config_json TEXT NOT NULL, created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP, updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP ); CREATE TABLE IF NOT EXISTS service_collections ( id TEXT PRIMARY KEY, name TEXT UNIQUE NOT NULL, config_json TEXT NOT NULL, created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP, updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP ); CREATE TABLE IF NOT EXISTS user_tokens ( user_id TEXT NOT NULL, service_id TEXT NOT NULL, config_json TEXT NOT NULL, updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP, PRIMARY KEY (user_id, service_id) ); CREATE TABLE IF NOT EXISTS service_templates ( id TEXT PRIMARY KEY, name TEXT NOT NULL, config_json TEXT NOT NULL, created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP, updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP ); CREATE TABLE IF NOT EXISTS logs ( id TEXT PRIMARY KEY, timestamp TIMESTAMPTZ NOT NULL, level TEXT NOT NULL, source TEXT, message TEXT NOT NULL, metadata_json TEXT, created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP ); CREATE INDEX IF NOT EXISTS idx_logs_timestamp ON logs(timestamp);").WillReturnResult(sqlmock.NewResult(0, 0))
	    pgDB, err := NewDBFromSQLDB(db)
	    require.NoError(t, err)
	    store := NewStore(pgDB)

        mock.MatchExpectationsInOrder(false)
		mock.ExpectQuery("SELECT config_json FROM upstream_services").WillReturnRows(sqlmock.NewRows([]string{"config_json"}).AddRow(`{"name":"service1"}`).RowError(0, errors.New("row error services")))
		mock.ExpectQuery("SELECT config_json FROM users").WillReturnRows(sqlmock.NewRows([]string{"config_json"}).AddRow(`{"id":"user1"}`))
		mock.ExpectQuery("SELECT config_json FROM global_settings WHERE id = 1").WillReturnRows(sqlmock.NewRows([]string{"config_json"}).AddRow(`{}`))
		mock.ExpectQuery("SELECT config_json FROM service_collections").WillReturnRows(sqlmock.NewRows([]string{"config_json"}).AddRow(`{"name":"collection1"}`))
		mock.ExpectQuery("SELECT config_json FROM profile_definitions").WillReturnRows(sqlmock.NewRows([]string{"config_json"}).AddRow(`{"name":"profile1"}`))

		ctx := context.Background()
		config, err := store.Load(ctx)
		require.Error(t, err)
		assert.Nil(t, config)
		assert.Contains(t, err.Error(), "row error services")
}

func TestStore_Load_ErrorRowsUsers(t *testing.T) {
	    db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	    require.NoError(t, err)
	    defer db.Close()
        mock.ExpectExec("CREATE TABLE IF NOT EXISTS upstream_services ( id TEXT PRIMARY KEY, name TEXT UNIQUE NOT NULL, config_json TEXT NOT NULL, created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP, updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP ); CREATE TABLE IF NOT EXISTS users ( id TEXT PRIMARY KEY, config_json TEXT NOT NULL, created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP, updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP ); CREATE TABLE IF NOT EXISTS global_settings ( id INTEGER PRIMARY KEY CHECK (id = 1), config_json TEXT NOT NULL, updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP ); CREATE TABLE IF NOT EXISTS secrets ( id TEXT PRIMARY KEY, name TEXT NOT NULL, key TEXT NOT NULL, config_json TEXT NOT NULL, created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP, updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP ); CREATE TABLE IF NOT EXISTS profile_definitions ( id TEXT PRIMARY KEY, name TEXT UNIQUE NOT NULL, config_json TEXT NOT NULL, created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP, updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP ); CREATE TABLE IF NOT EXISTS service_collections ( id TEXT PRIMARY KEY, name TEXT UNIQUE NOT NULL, config_json TEXT NOT NULL, created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP, updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP ); CREATE TABLE IF NOT EXISTS user_tokens ( user_id TEXT NOT NULL, service_id TEXT NOT NULL, config_json TEXT NOT NULL, updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP, PRIMARY KEY (user_id, service_id) ); CREATE TABLE IF NOT EXISTS service_templates ( id TEXT PRIMARY KEY, name TEXT NOT NULL, config_json TEXT NOT NULL, created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP, updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP ); CREATE TABLE IF NOT EXISTS logs ( id TEXT PRIMARY KEY, timestamp TIMESTAMPTZ NOT NULL, level TEXT NOT NULL, source TEXT, message TEXT NOT NULL, metadata_json TEXT, created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP ); CREATE INDEX IF NOT EXISTS idx_logs_timestamp ON logs(timestamp);").WillReturnResult(sqlmock.NewResult(0, 0))
	    pgDB, err := NewDBFromSQLDB(db)
	    require.NoError(t, err)
	    store := NewStore(pgDB)

        mock.MatchExpectationsInOrder(false)
		mock.ExpectQuery("SELECT config_json FROM upstream_services").WillReturnRows(sqlmock.NewRows([]string{"config_json"}).AddRow(`{"name":"service1"}`))
		mock.ExpectQuery("SELECT config_json FROM users").WillReturnRows(sqlmock.NewRows([]string{"config_json"}).AddRow(`{"id":"user1"}`).RowError(0, errors.New("row error users")))
		mock.ExpectQuery("SELECT config_json FROM global_settings WHERE id = 1").WillReturnRows(sqlmock.NewRows([]string{"config_json"}).AddRow(`{}`))
		mock.ExpectQuery("SELECT config_json FROM service_collections").WillReturnRows(sqlmock.NewRows([]string{"config_json"}).AddRow(`{"name":"collection1"}`))
		mock.ExpectQuery("SELECT config_json FROM profile_definitions").WillReturnRows(sqlmock.NewRows([]string{"config_json"}).AddRow(`{"name":"profile1"}`))

		ctx := context.Background()
		config, err := store.Load(ctx)
		require.Error(t, err)
		assert.Nil(t, config)
		assert.Contains(t, err.Error(), "row error users")
}

func TestStore_Load_ErrorRowsCollections(t *testing.T) {
	    db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	    require.NoError(t, err)
	    defer db.Close()
        mock.ExpectExec("CREATE TABLE IF NOT EXISTS upstream_services ( id TEXT PRIMARY KEY, name TEXT UNIQUE NOT NULL, config_json TEXT NOT NULL, created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP, updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP ); CREATE TABLE IF NOT EXISTS users ( id TEXT PRIMARY KEY, config_json TEXT NOT NULL, created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP, updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP ); CREATE TABLE IF NOT EXISTS global_settings ( id INTEGER PRIMARY KEY CHECK (id = 1), config_json TEXT NOT NULL, updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP ); CREATE TABLE IF NOT EXISTS secrets ( id TEXT PRIMARY KEY, name TEXT NOT NULL, key TEXT NOT NULL, config_json TEXT NOT NULL, created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP, updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP ); CREATE TABLE IF NOT EXISTS profile_definitions ( id TEXT PRIMARY KEY, name TEXT UNIQUE NOT NULL, config_json TEXT NOT NULL, created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP, updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP ); CREATE TABLE IF NOT EXISTS service_collections ( id TEXT PRIMARY KEY, name TEXT UNIQUE NOT NULL, config_json TEXT NOT NULL, created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP, updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP ); CREATE TABLE IF NOT EXISTS user_tokens ( user_id TEXT NOT NULL, service_id TEXT NOT NULL, config_json TEXT NOT NULL, updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP, PRIMARY KEY (user_id, service_id) ); CREATE TABLE IF NOT EXISTS service_templates ( id TEXT PRIMARY KEY, name TEXT NOT NULL, config_json TEXT NOT NULL, created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP, updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP ); CREATE TABLE IF NOT EXISTS logs ( id TEXT PRIMARY KEY, timestamp TIMESTAMPTZ NOT NULL, level TEXT NOT NULL, source TEXT, message TEXT NOT NULL, metadata_json TEXT, created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP ); CREATE INDEX IF NOT EXISTS idx_logs_timestamp ON logs(timestamp);").WillReturnResult(sqlmock.NewResult(0, 0))
	    pgDB, err := NewDBFromSQLDB(db)
	    require.NoError(t, err)
	    store := NewStore(pgDB)

        mock.MatchExpectationsInOrder(false)
		mock.ExpectQuery("SELECT config_json FROM upstream_services").WillReturnRows(sqlmock.NewRows([]string{"config_json"}).AddRow(`{"name":"service1"}`))
		mock.ExpectQuery("SELECT config_json FROM users").WillReturnRows(sqlmock.NewRows([]string{"config_json"}).AddRow(`{"id":"user1"}`))
		mock.ExpectQuery("SELECT config_json FROM global_settings WHERE id = 1").WillReturnRows(sqlmock.NewRows([]string{"config_json"}).AddRow(`{}`))
		mock.ExpectQuery("SELECT config_json FROM service_collections").WillReturnRows(sqlmock.NewRows([]string{"config_json"}).AddRow(`{"name":"collection1"}`).RowError(0, errors.New("row error collections")))
		mock.ExpectQuery("SELECT config_json FROM profile_definitions").WillReturnRows(sqlmock.NewRows([]string{"config_json"}).AddRow(`{"name":"profile1"}`))

		ctx := context.Background()
		config, err := store.Load(ctx)
		require.Error(t, err)
		assert.Nil(t, config)
		assert.Contains(t, err.Error(), "row error collections")
}

func TestStore_Load_ErrorRowsProfiles(t *testing.T) {
	    db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	    require.NoError(t, err)
	    defer db.Close()
        mock.ExpectExec("CREATE TABLE IF NOT EXISTS upstream_services ( id TEXT PRIMARY KEY, name TEXT UNIQUE NOT NULL, config_json TEXT NOT NULL, created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP, updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP ); CREATE TABLE IF NOT EXISTS users ( id TEXT PRIMARY KEY, config_json TEXT NOT NULL, created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP, updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP ); CREATE TABLE IF NOT EXISTS global_settings ( id INTEGER PRIMARY KEY CHECK (id = 1), config_json TEXT NOT NULL, updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP ); CREATE TABLE IF NOT EXISTS secrets ( id TEXT PRIMARY KEY, name TEXT NOT NULL, key TEXT NOT NULL, config_json TEXT NOT NULL, created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP, updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP ); CREATE TABLE IF NOT EXISTS profile_definitions ( id TEXT PRIMARY KEY, name TEXT UNIQUE NOT NULL, config_json TEXT NOT NULL, created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP, updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP ); CREATE TABLE IF NOT EXISTS service_collections ( id TEXT PRIMARY KEY, name TEXT UNIQUE NOT NULL, config_json TEXT NOT NULL, created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP, updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP ); CREATE TABLE IF NOT EXISTS user_tokens ( user_id TEXT NOT NULL, service_id TEXT NOT NULL, config_json TEXT NOT NULL, updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP, PRIMARY KEY (user_id, service_id) ); CREATE TABLE IF NOT EXISTS service_templates ( id TEXT PRIMARY KEY, name TEXT NOT NULL, config_json TEXT NOT NULL, created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP, updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP ); CREATE TABLE IF NOT EXISTS logs ( id TEXT PRIMARY KEY, timestamp TIMESTAMPTZ NOT NULL, level TEXT NOT NULL, source TEXT, message TEXT NOT NULL, metadata_json TEXT, created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP ); CREATE INDEX IF NOT EXISTS idx_logs_timestamp ON logs(timestamp);").WillReturnResult(sqlmock.NewResult(0, 0))
	    pgDB, err := NewDBFromSQLDB(db)
	    require.NoError(t, err)
	    store := NewStore(pgDB)

        mock.MatchExpectationsInOrder(false)
		mock.ExpectQuery("SELECT config_json FROM upstream_services").WillReturnRows(sqlmock.NewRows([]string{"config_json"}).AddRow(`{"name":"service1"}`))
		mock.ExpectQuery("SELECT config_json FROM users").WillReturnRows(sqlmock.NewRows([]string{"config_json"}).AddRow(`{"id":"user1"}`))
		mock.ExpectQuery("SELECT config_json FROM global_settings WHERE id = 1").WillReturnRows(sqlmock.NewRows([]string{"config_json"}).AddRow(`{}`))
		mock.ExpectQuery("SELECT config_json FROM service_collections").WillReturnRows(sqlmock.NewRows([]string{"config_json"}).AddRow(`{"name":"collection1"}`))
		mock.ExpectQuery("SELECT config_json FROM profile_definitions").WillReturnRows(sqlmock.NewRows([]string{"config_json"}).AddRow(`{"name":"profile1"}`).RowError(0, errors.New("row error profiles")))

		ctx := context.Background()
		config, err := store.Load(ctx)
		require.Error(t, err)
		assert.Nil(t, config)
		assert.Contains(t, err.Error(), "row error profiles")
}

func TestStore_Load_ScanErrorUsers(t *testing.T) {
	    db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	    require.NoError(t, err)
	    defer db.Close()
        mock.ExpectExec("CREATE TABLE IF NOT EXISTS upstream_services ( id TEXT PRIMARY KEY, name TEXT UNIQUE NOT NULL, config_json TEXT NOT NULL, created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP, updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP ); CREATE TABLE IF NOT EXISTS users ( id TEXT PRIMARY KEY, config_json TEXT NOT NULL, created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP, updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP ); CREATE TABLE IF NOT EXISTS global_settings ( id INTEGER PRIMARY KEY CHECK (id = 1), config_json TEXT NOT NULL, updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP ); CREATE TABLE IF NOT EXISTS secrets ( id TEXT PRIMARY KEY, name TEXT NOT NULL, key TEXT NOT NULL, config_json TEXT NOT NULL, created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP, updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP ); CREATE TABLE IF NOT EXISTS profile_definitions ( id TEXT PRIMARY KEY, name TEXT UNIQUE NOT NULL, config_json TEXT NOT NULL, created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP, updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP ); CREATE TABLE IF NOT EXISTS service_collections ( id TEXT PRIMARY KEY, name TEXT UNIQUE NOT NULL, config_json TEXT NOT NULL, created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP, updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP ); CREATE TABLE IF NOT EXISTS user_tokens ( user_id TEXT NOT NULL, service_id TEXT NOT NULL, config_json TEXT NOT NULL, updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP, PRIMARY KEY (user_id, service_id) ); CREATE TABLE IF NOT EXISTS service_templates ( id TEXT PRIMARY KEY, name TEXT NOT NULL, config_json TEXT NOT NULL, created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP, updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP ); CREATE TABLE IF NOT EXISTS logs ( id TEXT PRIMARY KEY, timestamp TIMESTAMPTZ NOT NULL, level TEXT NOT NULL, source TEXT, message TEXT NOT NULL, metadata_json TEXT, created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP ); CREATE INDEX IF NOT EXISTS idx_logs_timestamp ON logs(timestamp);").WillReturnResult(sqlmock.NewResult(0, 0))
	    pgDB, err := NewDBFromSQLDB(db)
	    require.NoError(t, err)
	    store := NewStore(pgDB)

        mock.MatchExpectationsInOrder(false)
		mock.ExpectQuery("SELECT config_json FROM upstream_services").WillReturnRows(sqlmock.NewRows([]string{"config_json"}).AddRow(`{"name":"service1"}`))
		mock.ExpectQuery("SELECT config_json FROM users").WillReturnRows(sqlmock.NewRows([]string{"config_json", "extra"}).AddRow(`{"id":"user1"}`, "bad")) // Force scan error
		mock.ExpectQuery("SELECT config_json FROM global_settings WHERE id = 1").WillReturnRows(sqlmock.NewRows([]string{"config_json"}).AddRow(`{}`))
		mock.ExpectQuery("SELECT config_json FROM service_collections").WillReturnRows(sqlmock.NewRows([]string{"config_json"}).AddRow(`{"name":"collection1"}`))
		mock.ExpectQuery("SELECT config_json FROM profile_definitions").WillReturnRows(sqlmock.NewRows([]string{"config_json"}).AddRow(`{"name":"profile1"}`))

		ctx := context.Background()
		config, err := store.Load(ctx)
		require.Error(t, err)
		assert.Nil(t, config)
		assert.Contains(t, err.Error(), "scan user config_json")
}

func TestStore_Load_ScanErrorProfiles(t *testing.T) {
	    db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	    require.NoError(t, err)
	    defer db.Close()
        mock.ExpectExec("CREATE TABLE IF NOT EXISTS upstream_services ( id TEXT PRIMARY KEY, name TEXT UNIQUE NOT NULL, config_json TEXT NOT NULL, created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP, updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP ); CREATE TABLE IF NOT EXISTS users ( id TEXT PRIMARY KEY, config_json TEXT NOT NULL, created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP, updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP ); CREATE TABLE IF NOT EXISTS global_settings ( id INTEGER PRIMARY KEY CHECK (id = 1), config_json TEXT NOT NULL, updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP ); CREATE TABLE IF NOT EXISTS secrets ( id TEXT PRIMARY KEY, name TEXT NOT NULL, key TEXT NOT NULL, config_json TEXT NOT NULL, created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP, updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP ); CREATE TABLE IF NOT EXISTS profile_definitions ( id TEXT PRIMARY KEY, name TEXT UNIQUE NOT NULL, config_json TEXT NOT NULL, created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP, updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP ); CREATE TABLE IF NOT EXISTS service_collections ( id TEXT PRIMARY KEY, name TEXT UNIQUE NOT NULL, config_json TEXT NOT NULL, created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP, updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP ); CREATE TABLE IF NOT EXISTS user_tokens ( user_id TEXT NOT NULL, service_id TEXT NOT NULL, config_json TEXT NOT NULL, updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP, PRIMARY KEY (user_id, service_id) ); CREATE TABLE IF NOT EXISTS service_templates ( id TEXT PRIMARY KEY, name TEXT NOT NULL, config_json TEXT NOT NULL, created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP, updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP ); CREATE TABLE IF NOT EXISTS logs ( id TEXT PRIMARY KEY, timestamp TIMESTAMPTZ NOT NULL, level TEXT NOT NULL, source TEXT, message TEXT NOT NULL, metadata_json TEXT, created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP ); CREATE INDEX IF NOT EXISTS idx_logs_timestamp ON logs(timestamp);").WillReturnResult(sqlmock.NewResult(0, 0))
	    pgDB, err := NewDBFromSQLDB(db)
	    require.NoError(t, err)
	    store := NewStore(pgDB)

        mock.MatchExpectationsInOrder(false)
		mock.ExpectQuery("SELECT config_json FROM upstream_services").WillReturnRows(sqlmock.NewRows([]string{"config_json"}).AddRow(`{"name":"service1"}`))
		mock.ExpectQuery("SELECT config_json FROM users").WillReturnRows(sqlmock.NewRows([]string{"config_json"}).AddRow(`{"id":"user1"}`))
		mock.ExpectQuery("SELECT config_json FROM global_settings WHERE id = 1").WillReturnRows(sqlmock.NewRows([]string{"config_json"}).AddRow(`{}`))
		mock.ExpectQuery("SELECT config_json FROM service_collections").WillReturnRows(sqlmock.NewRows([]string{"config_json"}).AddRow(`{"name":"collection1"}`))
		mock.ExpectQuery("SELECT config_json FROM profile_definitions").WillReturnRows(sqlmock.NewRows([]string{"config_json", "extra"}).AddRow(`{"name":"profile1"}`, "bad"))

		ctx := context.Background()
		config, err := store.Load(ctx)
		require.Error(t, err)
		assert.Nil(t, config)
		assert.Contains(t, err.Error(), "scan profile config_json")
}

func TestStore_Load_ScanErrorCollections(t *testing.T) {
	    db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	    require.NoError(t, err)
	    defer db.Close()
        mock.ExpectExec("CREATE TABLE IF NOT EXISTS upstream_services ( id TEXT PRIMARY KEY, name TEXT UNIQUE NOT NULL, config_json TEXT NOT NULL, created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP, updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP ); CREATE TABLE IF NOT EXISTS users ( id TEXT PRIMARY KEY, config_json TEXT NOT NULL, created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP, updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP ); CREATE TABLE IF NOT EXISTS global_settings ( id INTEGER PRIMARY KEY CHECK (id = 1), config_json TEXT NOT NULL, updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP ); CREATE TABLE IF NOT EXISTS secrets ( id TEXT PRIMARY KEY, name TEXT NOT NULL, key TEXT NOT NULL, config_json TEXT NOT NULL, created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP, updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP ); CREATE TABLE IF NOT EXISTS profile_definitions ( id TEXT PRIMARY KEY, name TEXT UNIQUE NOT NULL, config_json TEXT NOT NULL, created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP, updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP ); CREATE TABLE IF NOT EXISTS service_collections ( id TEXT PRIMARY KEY, name TEXT UNIQUE NOT NULL, config_json TEXT NOT NULL, created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP, updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP ); CREATE TABLE IF NOT EXISTS user_tokens ( user_id TEXT NOT NULL, service_id TEXT NOT NULL, config_json TEXT NOT NULL, updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP, PRIMARY KEY (user_id, service_id) ); CREATE TABLE IF NOT EXISTS service_templates ( id TEXT PRIMARY KEY, name TEXT NOT NULL, config_json TEXT NOT NULL, created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP, updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP ); CREATE TABLE IF NOT EXISTS logs ( id TEXT PRIMARY KEY, timestamp TIMESTAMPTZ NOT NULL, level TEXT NOT NULL, source TEXT, message TEXT NOT NULL, metadata_json TEXT, created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP ); CREATE INDEX IF NOT EXISTS idx_logs_timestamp ON logs(timestamp);").WillReturnResult(sqlmock.NewResult(0, 0))
	    pgDB, err := NewDBFromSQLDB(db)
	    require.NoError(t, err)
	    store := NewStore(pgDB)

        mock.MatchExpectationsInOrder(false)
		mock.ExpectQuery("SELECT config_json FROM upstream_services").WillReturnRows(sqlmock.NewRows([]string{"config_json"}).AddRow(`{"name":"service1"}`))
		mock.ExpectQuery("SELECT config_json FROM users").WillReturnRows(sqlmock.NewRows([]string{"config_json"}).AddRow(`{"id":"user1"}`))
		mock.ExpectQuery("SELECT config_json FROM global_settings WHERE id = 1").WillReturnRows(sqlmock.NewRows([]string{"config_json"}).AddRow(`{}`))
		mock.ExpectQuery("SELECT config_json FROM service_collections").WillReturnRows(sqlmock.NewRows([]string{"config_json", "extra"}).AddRow(`{"name":"collection1"}`, "bad"))
		mock.ExpectQuery("SELECT config_json FROM profile_definitions").WillReturnRows(sqlmock.NewRows([]string{"config_json"}).AddRow(`{"name":"profile1"}`))

		ctx := context.Background()
		config, err := store.Load(ctx)
		require.NoError(t, err)
		require.NotNil(t, config)
		assert.Len(t, config.GetCollections(), 0) // Should skip silently
}

func TestStore_Load_UnmarshalErrorProfiles(t *testing.T) {
	    db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	    require.NoError(t, err)
	    defer db.Close()
        mock.ExpectExec("CREATE TABLE IF NOT EXISTS upstream_services ( id TEXT PRIMARY KEY, name TEXT UNIQUE NOT NULL, config_json TEXT NOT NULL, created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP, updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP ); CREATE TABLE IF NOT EXISTS users ( id TEXT PRIMARY KEY, config_json TEXT NOT NULL, created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP, updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP ); CREATE TABLE IF NOT EXISTS global_settings ( id INTEGER PRIMARY KEY CHECK (id = 1), config_json TEXT NOT NULL, updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP ); CREATE TABLE IF NOT EXISTS secrets ( id TEXT PRIMARY KEY, name TEXT NOT NULL, key TEXT NOT NULL, config_json TEXT NOT NULL, created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP, updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP ); CREATE TABLE IF NOT EXISTS profile_definitions ( id TEXT PRIMARY KEY, name TEXT UNIQUE NOT NULL, config_json TEXT NOT NULL, created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP, updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP ); CREATE TABLE IF NOT EXISTS service_collections ( id TEXT PRIMARY KEY, name TEXT UNIQUE NOT NULL, config_json TEXT NOT NULL, created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP, updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP ); CREATE TABLE IF NOT EXISTS user_tokens ( user_id TEXT NOT NULL, service_id TEXT NOT NULL, config_json TEXT NOT NULL, updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP, PRIMARY KEY (user_id, service_id) ); CREATE TABLE IF NOT EXISTS service_templates ( id TEXT PRIMARY KEY, name TEXT NOT NULL, config_json TEXT NOT NULL, created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP, updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP ); CREATE TABLE IF NOT EXISTS logs ( id TEXT PRIMARY KEY, timestamp TIMESTAMPTZ NOT NULL, level TEXT NOT NULL, source TEXT, message TEXT NOT NULL, metadata_json TEXT, created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP ); CREATE INDEX IF NOT EXISTS idx_logs_timestamp ON logs(timestamp);").WillReturnResult(sqlmock.NewResult(0, 0))
	    pgDB, err := NewDBFromSQLDB(db)
	    require.NoError(t, err)
	    store := NewStore(pgDB)

        mock.MatchExpectationsInOrder(false)
		mock.ExpectQuery("SELECT config_json FROM upstream_services").WillReturnRows(sqlmock.NewRows([]string{"config_json"}).AddRow(`{"name":"service1"}`))
		mock.ExpectQuery("SELECT config_json FROM users").WillReturnRows(sqlmock.NewRows([]string{"config_json"}).AddRow(`{"id":"user1"}`))
		mock.ExpectQuery("SELECT config_json FROM global_settings WHERE id = 1").WillReturnRows(sqlmock.NewRows([]string{"config_json"}).AddRow(`{}`))
		mock.ExpectQuery("SELECT config_json FROM service_collections").WillReturnRows(sqlmock.NewRows([]string{"config_json"}).AddRow(`{"name":"collection1"}`))
		mock.ExpectQuery("SELECT config_json FROM profile_definitions").WillReturnRows(sqlmock.NewRows([]string{"config_json"}).AddRow(`{"name":"profile1"`)) // invalid JSON

		ctx := context.Background()
		config, err := store.Load(ctx)
		require.Error(t, err)
		assert.Nil(t, config)
		assert.Contains(t, err.Error(), "unmarshal profile config")
}

func TestStore_Load_UnmarshalErrorUsers(t *testing.T) {
	    db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	    require.NoError(t, err)
	    defer db.Close()
        mock.ExpectExec("CREATE TABLE IF NOT EXISTS upstream_services ( id TEXT PRIMARY KEY, name TEXT UNIQUE NOT NULL, config_json TEXT NOT NULL, created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP, updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP ); CREATE TABLE IF NOT EXISTS users ( id TEXT PRIMARY KEY, config_json TEXT NOT NULL, created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP, updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP ); CREATE TABLE IF NOT EXISTS global_settings ( id INTEGER PRIMARY KEY CHECK (id = 1), config_json TEXT NOT NULL, updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP ); CREATE TABLE IF NOT EXISTS secrets ( id TEXT PRIMARY KEY, name TEXT NOT NULL, key TEXT NOT NULL, config_json TEXT NOT NULL, created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP, updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP ); CREATE TABLE IF NOT EXISTS profile_definitions ( id TEXT PRIMARY KEY, name TEXT UNIQUE NOT NULL, config_json TEXT NOT NULL, created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP, updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP ); CREATE TABLE IF NOT EXISTS service_collections ( id TEXT PRIMARY KEY, name TEXT UNIQUE NOT NULL, config_json TEXT NOT NULL, created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP, updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP ); CREATE TABLE IF NOT EXISTS user_tokens ( user_id TEXT NOT NULL, service_id TEXT NOT NULL, config_json TEXT NOT NULL, updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP, PRIMARY KEY (user_id, service_id) ); CREATE TABLE IF NOT EXISTS service_templates ( id TEXT PRIMARY KEY, name TEXT NOT NULL, config_json TEXT NOT NULL, created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP, updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP ); CREATE TABLE IF NOT EXISTS logs ( id TEXT PRIMARY KEY, timestamp TIMESTAMPTZ NOT NULL, level TEXT NOT NULL, source TEXT, message TEXT NOT NULL, metadata_json TEXT, created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP ); CREATE INDEX IF NOT EXISTS idx_logs_timestamp ON logs(timestamp);").WillReturnResult(sqlmock.NewResult(0, 0))
	    pgDB, err := NewDBFromSQLDB(db)
	    require.NoError(t, err)
	    store := NewStore(pgDB)

        mock.MatchExpectationsInOrder(false)
		mock.ExpectQuery("SELECT config_json FROM upstream_services").WillReturnRows(sqlmock.NewRows([]string{"config_json"}).AddRow(`{"name":"service1"}`))
		mock.ExpectQuery("SELECT config_json FROM users").WillReturnRows(sqlmock.NewRows([]string{"config_json"}).AddRow(`{"id":"user1"`))
		mock.ExpectQuery("SELECT config_json FROM global_settings WHERE id = 1").WillReturnRows(sqlmock.NewRows([]string{"config_json"}).AddRow(`{}`))
		mock.ExpectQuery("SELECT config_json FROM service_collections").WillReturnRows(sqlmock.NewRows([]string{"config_json"}).AddRow(`{"name":"collection1"}`))
		mock.ExpectQuery("SELECT config_json FROM profile_definitions").WillReturnRows(sqlmock.NewRows([]string{"config_json"}).AddRow(`{"name":"profile1"}`))

		ctx := context.Background()
		config, err := store.Load(ctx)
		require.Error(t, err)
		assert.Nil(t, config)
		assert.Contains(t, err.Error(), "unmarshal user config")
}
