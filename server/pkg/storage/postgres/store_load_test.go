// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package postgres

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPostgresStore_Load(t *testing.T) {
	t.Run("Load_Success", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		pgDB := &DB{db}
		store := NewStore(pgDB)

		// Expect queries in any order
		mock.MatchExpectationsInOrder(false)

		// 1. Upstream Services
		svcRows := sqlmock.NewRows([]string{"config_json"}).
			AddRow(`{"name":"svc-1","id":"id-1"}`)
		mock.ExpectQuery("SELECT config_json FROM upstream_services").
			WillReturnRows(svcRows)

		// 2. Users
		userRows := sqlmock.NewRows([]string{"config_json"}).
			AddRow(`{"id":"user-1"}`)
		mock.ExpectQuery("SELECT config_json FROM users").
			WillReturnRows(userRows)

		// 3. Global Settings
		settingsRows := sqlmock.NewRows([]string{"config_json"}).
			AddRow(`{"mcp_listen_address":":8080"}`)
		mock.ExpectQuery("SELECT config_json FROM global_settings WHERE id = 1").
			WillReturnRows(settingsRows)

		// 4. Service Collections
		collectionRows := sqlmock.NewRows([]string{"config_json"}).
			AddRow(`{"name":"col-1"}`)
		mock.ExpectQuery("SELECT config_json FROM service_collections").
			WillReturnRows(collectionRows)

		// Execute Load
		config, err := store.Load(context.Background())
		require.NoError(t, err)

		// Verify Results
		assert.Len(t, config.GetUpstreamServices(), 1)
		assert.Equal(t, "svc-1", config.GetUpstreamServices()[0].GetName())

		assert.Len(t, config.GetUsers(), 1)
		assert.Equal(t, "user-1", config.GetUsers()[0].GetId())

		assert.Equal(t, ":8080", config.GetGlobalSettings().GetMcpListenAddress())

		assert.Len(t, config.GetCollections(), 1)
		assert.Equal(t, "col-1", config.GetCollections()[0].GetName())

		require.NoError(t, mock.ExpectationsWereMet())
	})
}
