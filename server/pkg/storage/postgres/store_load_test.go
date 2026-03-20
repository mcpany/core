// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package postgres

import (
	"context"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStore_Load(t *testing.T) {
	t.Run("Happy Path", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()
		mock.MatchExpectationsInOrder(false)

		pgDB := &DB{db}
		store := NewStore(pgDB)

		// Mock services
		mock.ExpectQuery("SELECT config_json FROM upstream_services").
			WillReturnRows(sqlmock.NewRows([]string{"config_json"}).AddRow(`{"id":"s1","name":"service1"}`))

		// Mock users
		mock.ExpectQuery("SELECT config_json FROM users").
			WillReturnRows(sqlmock.NewRows([]string{"config_json"}).AddRow(`{"id":"u1","username":"user1"}`))

		// Mock Global Settings
		mock.ExpectQuery("SELECT config_json FROM global_settings WHERE id = 1").
			WillReturnRows(sqlmock.NewRows([]string{"config_json"}).AddRow(`{}`))

		// Mock Collections
		mock.ExpectQuery("SELECT config_json FROM service_collections").
			WillReturnRows(sqlmock.NewRows([]string{"config_json"}).AddRow(`{"name":"col1"}`))

		// Mock Profiles
		mock.ExpectQuery("SELECT config_json FROM profile_definitions").
			WillReturnRows(sqlmock.NewRows([]string{"config_json"}).AddRow(`{"name":"prof1"}`))

		config, err := store.Load(context.Background())
		require.NoError(t, err)
		require.NotNil(t, config)

		assert.Len(t, config.GetUpstreamServices(), 1)
		assert.Equal(t, "s1", config.GetUpstreamServices()[0].GetId())

		assert.Len(t, config.GetUsers(), 1)
		assert.Equal(t, "u1", config.GetUsers()[0].GetId())

		assert.NotNil(t, config.GetGlobalSettings())

		assert.Len(t, config.GetCollections(), 1)
		assert.Equal(t, "col1", config.GetCollections()[0].GetName())

		assert.Len(t, config.GetGlobalSettings().GetProfileDefinitions(), 1)
		assert.Equal(t, "prof1", config.GetGlobalSettings().GetProfileDefinitions()[0].GetName())

		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Services Query Error", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()
		mock.MatchExpectationsInOrder(false)

		pgDB := &DB{db}
		store := NewStore(pgDB)

		mock.ExpectQuery("SELECT config_json FROM upstream_services").
			WillReturnError(errors.New("db error"))

		mock.ExpectQuery("SELECT config_json FROM users").
			WillReturnRows(sqlmock.NewRows([]string{"config_json"}).AddRow(`{"id":"u1"}`))
		mock.ExpectQuery("SELECT config_json FROM global_settings WHERE id = 1").
			WillReturnRows(sqlmock.NewRows([]string{"config_json"}))
		mock.ExpectQuery("SELECT config_json FROM service_collections").
			WillReturnRows(sqlmock.NewRows([]string{"config_json"}))
		mock.ExpectQuery("SELECT config_json FROM profile_definitions").
			WillReturnRows(sqlmock.NewRows([]string{"config_json"}))

		_, err = store.Load(context.Background())
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to query upstream_services")
	})

	t.Run("Users Query Error", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()
		mock.MatchExpectationsInOrder(false)

		pgDB := &DB{db}
		store := NewStore(pgDB)

		mock.ExpectQuery("SELECT config_json FROM upstream_services").
			WillReturnRows(sqlmock.NewRows([]string{"config_json"}))

		mock.ExpectQuery("SELECT config_json FROM users").
			WillReturnError(errors.New("db error"))

		mock.ExpectQuery("SELECT config_json FROM global_settings WHERE id = 1").
			WillReturnRows(sqlmock.NewRows([]string{"config_json"}))
		mock.ExpectQuery("SELECT config_json FROM service_collections").
			WillReturnRows(sqlmock.NewRows([]string{"config_json"}))
		mock.ExpectQuery("SELECT config_json FROM profile_definitions").
			WillReturnRows(sqlmock.NewRows([]string{"config_json"}))

		_, err = store.Load(context.Background())
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to query users")
	})

	t.Run("Profiles Query Error", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()
		mock.MatchExpectationsInOrder(false)

		pgDB := &DB{db}
		store := NewStore(pgDB)

		mock.ExpectQuery("SELECT config_json FROM upstream_services").
			WillReturnRows(sqlmock.NewRows([]string{"config_json"}))
		mock.ExpectQuery("SELECT config_json FROM users").
			WillReturnRows(sqlmock.NewRows([]string{"config_json"}))
		mock.ExpectQuery("SELECT config_json FROM global_settings WHERE id = 1").
			WillReturnRows(sqlmock.NewRows([]string{"config_json"}))
		mock.ExpectQuery("SELECT config_json FROM service_collections").
			WillReturnRows(sqlmock.NewRows([]string{"config_json"}))

		mock.ExpectQuery("SELECT config_json FROM profile_definitions").
			WillReturnError(errors.New("db error"))

		_, err = store.Load(context.Background())
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to query profile_definitions")
	})

	t.Run("Invalid JSON Unmarshal", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()
		mock.MatchExpectationsInOrder(false)

		pgDB := &DB{db}
		store := NewStore(pgDB)

		mock.ExpectQuery("SELECT config_json FROM upstream_services").
			WillReturnRows(sqlmock.NewRows([]string{"config_json"}).AddRow(`invalid`))

		mock.ExpectQuery("SELECT config_json FROM users").
			WillReturnRows(sqlmock.NewRows([]string{"config_json"}))
		mock.ExpectQuery("SELECT config_json FROM global_settings WHERE id = 1").
			WillReturnRows(sqlmock.NewRows([]string{"config_json"}))
		mock.ExpectQuery("SELECT config_json FROM service_collections").
			WillReturnRows(sqlmock.NewRows([]string{"config_json"}))
		mock.ExpectQuery("SELECT config_json FROM profile_definitions").
			WillReturnRows(sqlmock.NewRows([]string{"config_json"}))

		_, err = store.Load(context.Background())
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to unmarshal service config")
	})

	t.Run("Ignore Collection Error", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()
		mock.MatchExpectationsInOrder(false)

		pgDB := &DB{db}
		store := NewStore(pgDB)

		mock.ExpectQuery("SELECT config_json FROM upstream_services").
			WillReturnRows(sqlmock.NewRows([]string{"config_json"}))
		mock.ExpectQuery("SELECT config_json FROM users").
			WillReturnRows(sqlmock.NewRows([]string{"config_json"}))
		mock.ExpectQuery("SELECT config_json FROM global_settings WHERE id = 1").
			WillReturnRows(sqlmock.NewRows([]string{"config_json"}))

		mock.ExpectQuery("SELECT config_json FROM service_collections").
			WillReturnError(errors.New("db error")) // Collection errors are ignored

		mock.ExpectQuery("SELECT config_json FROM profile_definitions").
			WillReturnRows(sqlmock.NewRows([]string{"config_json"}))

		config, err := store.Load(context.Background())
		require.NoError(t, err)
		require.NotNil(t, config)
	})

	t.Run("Scan Errors", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()
		mock.MatchExpectationsInOrder(false)

		pgDB := &DB{db}
		store := NewStore(pgDB)

		mock.ExpectQuery("SELECT config_json FROM upstream_services").
			WillReturnRows(sqlmock.NewRows([]string{"config_json", "extra"}).AddRow(`{}`, "extra"))

		mock.ExpectQuery("SELECT config_json FROM users").
			WillReturnRows(sqlmock.NewRows([]string{"config_json"}))
		mock.ExpectQuery("SELECT config_json FROM global_settings WHERE id = 1").
			WillReturnRows(sqlmock.NewRows([]string{"config_json"}))
		mock.ExpectQuery("SELECT config_json FROM service_collections").
			WillReturnRows(sqlmock.NewRows([]string{"config_json"}))
		mock.ExpectQuery("SELECT config_json FROM profile_definitions").
			WillReturnRows(sqlmock.NewRows([]string{"config_json"}))

		_, err = store.Load(context.Background())
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to scan config_json")
	})

	t.Run("Scan Errors Profile", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()
		mock.MatchExpectationsInOrder(false)

		pgDB := &DB{db}
		store := NewStore(pgDB)

		mock.ExpectQuery("SELECT config_json FROM upstream_services").
			WillReturnRows(sqlmock.NewRows([]string{"config_json"}))
		mock.ExpectQuery("SELECT config_json FROM users").
			WillReturnRows(sqlmock.NewRows([]string{"config_json"}))
		mock.ExpectQuery("SELECT config_json FROM global_settings WHERE id = 1").
			WillReturnRows(sqlmock.NewRows([]string{"config_json"}))
		mock.ExpectQuery("SELECT config_json FROM service_collections").
			WillReturnRows(sqlmock.NewRows([]string{"config_json"}))

		mock.ExpectQuery("SELECT config_json FROM profile_definitions").
			WillReturnRows(sqlmock.NewRows([]string{"config_json", "extra"}).AddRow(`{}`, "extra"))

		_, err = store.Load(context.Background())
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to scan profile config_json")
	})
}
