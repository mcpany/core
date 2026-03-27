// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package postgres

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

// TestStore_Load tests the Load method of the PostgreSQL store.
func TestStore_Load(t *testing.T) {
	t.Run("Happy Path", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		pgDB := &DB{DB: db}
		store := NewStore(pgDB)

		svc := configv1.UpstreamServiceConfig_builder{Id: proto.String("service-1"), Name: proto.String("Service One")}.Build()
		svcBytes, err := protojson.MarshalOptions{}.Marshal(svc)
		require.NoError(t, err)
		mock.ExpectQuery(".*").
			WillReturnRows(sqlmock.NewRows([]string{"config_json"}).AddRow(svcBytes))

		user := configv1.User_builder{Id: proto.String("user-1"), Roles: []string{"admin"}}.Build()
		userBytes, err := protojson.MarshalOptions{}.Marshal(user)
		require.NoError(t, err)
		mock.ExpectQuery(".*").
			WillReturnRows(sqlmock.NewRows([]string{"config_json"}).AddRow(userBytes))

		settings := configv1.GlobalSettings_builder{LogLevel: configv1.GlobalSettings_LOG_LEVEL_INFO.Enum()}.Build()
		settingsBytes, err := protojson.MarshalOptions{}.Marshal(settings)
		require.NoError(t, err)
		mock.ExpectQuery(".*").
			WillReturnRows(sqlmock.NewRows([]string{"config_json"}).AddRow(settingsBytes))

		coll := configv1.Collection_builder{Name: proto.String("Collection One"), Version: proto.String("collection-1")}.Build()
		collBytes, err := protojson.MarshalOptions{}.Marshal(coll)
		require.NoError(t, err)
		mock.ExpectQuery(".*").
			WillReturnRows(sqlmock.NewRows([]string{"config_json"}).AddRow(collBytes))

		profile := configv1.ProfileDefinition_builder{Name: proto.String("profile-1")}.Build()
		profileBytes, err := protojson.MarshalOptions{}.Marshal(profile)
		require.NoError(t, err)
		mock.ExpectQuery(".*").
			WillReturnRows(sqlmock.NewRows([]string{"config_json"}).AddRow(profileBytes))

		cfg, err := store.Load(context.Background())
		require.NoError(t, err)
		require.NotNil(t, cfg)
	})

	t.Run("Query Error", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		pgDB := &DB{DB: db}
		store := NewStore(pgDB)

		mock.ExpectQuery(".*").
			WillReturnError(errors.New("db error"))

		cfg, err := store.Load(context.Background())
		require.Error(t, err)
		require.Nil(t, cfg)
	})

	t.Run("Scan Error - Invalid JSON", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		pgDB := &DB{DB: db}
		store := NewStore(pgDB)

		mock.ExpectQuery(".*").
			WillReturnRows(sqlmock.NewRows([]string{"config_json"}).AddRow([]byte("invalid json")))

		cfg, err := store.Load(context.Background())
		require.Error(t, err)
		require.Nil(t, cfg)
	})

	t.Run("Settings Not Found", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		pgDB := &DB{DB: db}
		store := NewStore(pgDB)

		mock.ExpectQuery(".*").
			WillReturnError(sql.ErrNoRows)

		cfg, err := store.Load(context.Background())
		// Load will return error but we just want to ensure it completes.
		require.Error(t, err)
		require.Nil(t, cfg)
	})
}
