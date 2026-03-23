// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package postgres

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

// TestStore_Load tests the Load method of the PostgreSQL store.
//
// Parameters:
//   - t (*testing.T): The testing context.
//
// Returns:
//   - None.
//
// Errors:
//   - None.
//
// Side Effects:
//   - Modifies testing state through assertions.
func TestStore_Load(t *testing.T) {
	t.Run("Happy Path", func(t *testing.T) {
		db, mock, err := sqlmock.New(sqlmock.MatchExpectationsInOrder(false))
		require.NoError(t, err)
		defer db.Close()

		pgDB := &DB{db}
		store := NewStore(pgDB)

		opts := protojson.MarshalOptions{UseProtoNames: true}

		// 1. Services
		svc := &configv1.UpstreamServiceConfig{
			Id:   proto.String("service-1"),
			Name: proto.String("Service One"),
		}
		svcBytes, err = opts.Marshal(svc)
		require.NoError(t, err)
		mock.ExpectQuery("SELECT config_json FROM upstream_services").
			WillReturnRows(sqlmock.NewRows([]string{"config_json"}).AddRow(svcBytes))

		// 2. Users
		user := &configv1.User{
			Id:    proto.String("user-1"),
			Roles: []string{"admin"},
		}
		userBytes, err = opts.Marshal(user)
		require.NoError(t, err)
		mock.ExpectQuery("SELECT config_json FROM users").
			WillReturnRows(sqlmock.NewRows([]string{"config_json"}).AddRow(userBytes))

		// 3. Settings
		settings := &configv1.GlobalSettings{
			InstanceId: proto.String("instance-1"),
		}
		settingsBytes, err = opts.Marshal(settings)
		require.NoError(t, err)
		mock.ExpectQuery("SELECT config_json FROM settings WHERE id = \\$1").
			WithArgs("global").
			WillReturnRows(sqlmock.NewRows([]string{"config_json"}).AddRow(settingsBytes))

		// 4. Collections
		coll := &configv1.Collection{
			Id:   proto.String("collection-1"),
			Name: proto.String("Collection One"),
		}
		collBytes, err = opts.Marshal(coll)
		require.NoError(t, err)
		mock.ExpectQuery("SELECT config_json FROM service_collections").
			WillReturnRows(sqlmock.NewRows([]string{"config_json"}).AddRow(collBytes))

		// 5. Profiles
		profile := &configv1.ProfileDefinition{
			Id:   proto.String("profile-1"),
			Name: proto.String("Profile One"),
		}
		profileBytes, err = opts.Marshal(profile)
		require.NoError(t, err)
		mock.ExpectQuery("SELECT config_json FROM profiles").
			WillReturnRows(sqlmock.NewRows([]string{"config_json"}).AddRow(profileBytes))

		cfg, err := store.Load(context.Background())
		require.NoError(t, err)
		require.NotNil(t, cfg)

		require.Len(t, cfg.GetServices(), 1)
		assert.Equal(t, "service-1", cfg.GetServices()[0].GetId())

		require.Len(t, cfg.GetUsers(), 1)
		assert.Equal(t, "user-1", cfg.GetUsers()[0].GetId())

		require.NotNil(t, cfg.GetSettings())
		assert.Equal(t, "instance-1", cfg.GetSettings().GetInstanceId())

		require.Len(t, cfg.GetCollections(), 1)
		assert.Equal(t, "collection-1", cfg.GetCollections()[0].GetId())

		require.Len(t, cfg.GetProfiles(), 1)
		assert.Equal(t, "profile-1", cfg.GetProfiles()[0].GetId())

		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Query Error", func(t *testing.T) {
		db, mock, err := sqlmock.New(sqlmock.MatchExpectationsInOrder(false))
		require.NoError(t, err)
		defer db.Close()

		pgDB := &DB{db}
		store := NewStore(pgDB)

		// Make the first query fail
		mock.ExpectQuery("SELECT config_json FROM upstream_services").
			WillReturnError(errors.New("db error"))

		// The others might execute or not depending on timing, but let's allow them to fail or succeed
		mock.ExpectQuery("SELECT config_json FROM users").
			WillReturnRows(sqlmock.NewRows([]string{"config_json"}))
		mock.ExpectQuery("SELECT config_json FROM settings WHERE id = \\$1").
			WithArgs("global").
			WillReturnRows(sqlmock.NewRows([]string{"config_json"}))
		mock.ExpectQuery("SELECT config_json FROM service_collections").
			WillReturnRows(sqlmock.NewRows([]string{"config_json"}))
		mock.ExpectQuery("SELECT config_json FROM profiles").
			WillReturnRows(sqlmock.NewRows([]string{"config_json"}))

		cfg, err := store.Load(context.Background())
		require.Error(t, err)
		require.Nil(t, cfg)
		assert.Contains(t, err.Error(), "failed to query upstream_services")
	})

	t.Run("Scan Error - Invalid JSON", func(t *testing.T) {
		db, mock, err := sqlmock.New(sqlmock.MatchExpectationsInOrder(false))
		require.NoError(t, err)
		defer db.Close()

		pgDB := &DB{db}
		store := NewStore(pgDB)

		// Return invalid JSON for services
		mock.ExpectQuery("SELECT config_json FROM upstream_services").
			WillReturnRows(sqlmock.NewRows([]string{"config_json"}).AddRow([]byte("invalid json")))

		// Empty for others
		mock.ExpectQuery("SELECT config_json FROM users").
			WillReturnRows(sqlmock.NewRows([]string{"config_json"}))
		mock.ExpectQuery("SELECT config_json FROM settings WHERE id = \\$1").
			WithArgs("global").
			WillReturnRows(sqlmock.NewRows([]string{"config_json"}))
		mock.ExpectQuery("SELECT config_json FROM service_collections").
			WillReturnRows(sqlmock.NewRows([]string{"config_json"}))
		mock.ExpectQuery("SELECT config_json FROM profiles").
			WillReturnRows(sqlmock.NewRows([]string{"config_json"}))

		cfg, err := store.Load(context.Background())
		require.Error(t, err)
		require.Nil(t, cfg)
		assert.Contains(t, err.Error(), "failed to unmarshal service config")
	})

	t.Run("Settings Not Found", func(t *testing.T) {
		db, mock, err := sqlmock.New(sqlmock.MatchExpectationsInOrder(false))
		require.NoError(t, err)
		defer db.Close()

		pgDB := &DB{db}
		store := NewStore(pgDB)

		// Empty rows
		mock.ExpectQuery("SELECT config_json FROM upstream_services").
			WillReturnRows(sqlmock.NewRows([]string{"config_json"}))
		mock.ExpectQuery("SELECT config_json FROM users").
			WillReturnRows(sqlmock.NewRows([]string{"config_json"}))

		// Return ErrNoRows for settings
		mock.ExpectQuery("SELECT config_json FROM settings WHERE id = \\$1").
			WithArgs("global").
			WillReturnError(sql.ErrNoRows)

		mock.ExpectQuery("SELECT config_json FROM service_collections").
			WillReturnRows(sqlmock.NewRows([]string{"config_json"}))
		mock.ExpectQuery("SELECT config_json FROM profiles").
			WillReturnRows(sqlmock.NewRows([]string{"config_json"}))

		cfg, err := store.Load(context.Background())
		// sql.ErrNoRows is treated as no settings, not a failure
		require.NoError(t, err)
		require.NotNil(t, cfg)
		require.Nil(t, cfg.GetSettings())

		require.NoError(t, mock.ExpectationsWereMet())
	})
}
