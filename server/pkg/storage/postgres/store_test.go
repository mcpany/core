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
)

// stringPtr returns a pointer to the passed string.
func stringPtr(s string) *string {
	return &s
}

func TestPostgresStore(t *testing.T) {
	// Create a new mock database
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	// Wrap sql.DB in our DB struct
	pgDB := &DB{db}
	store := NewStore(pgDB)

	t.Run("SaveService", func(t *testing.T) {
		svc := &configv1.UpstreamServiceConfig{
			Name: stringPtr("test-service"),
			Id:   stringPtr("test-id"),
		}

		mock.ExpectExec("INSERT INTO upstream_services").
			WithArgs("test-id", "test-service", sqlmock.AnyArg()).
			WillReturnResult(sqlmock.NewResult(1, 1))

		err := store.SaveService(context.Background(), svc)
		require.NoError(t, err)

		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("SaveService_Error", func(t *testing.T) {
		svc := &configv1.UpstreamServiceConfig{
			Name: stringPtr("test-service"),
			Id:   stringPtr("test-id"),
		}

		mock.ExpectExec("INSERT INTO upstream_services").
			WithArgs("test-id", "test-service", sqlmock.AnyArg()).
			WillReturnError(errors.New("db error"))

		err := store.SaveService(context.Background(), svc)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to save service")

		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("GetService", func(t *testing.T) {
		svc := &configv1.UpstreamServiceConfig{
			Name: stringPtr("test-service"),
			Id:   stringPtr("test-id"),
		}

		rows := sqlmock.NewRows([]string{"config_json"}).
			AddRow(`{"name":"test-service","id":"test-id"}`)

		mock.ExpectQuery("SELECT config_json FROM upstream_services").
			WithArgs("test-service").
			WillReturnRows(rows)

		got, err := store.GetService(context.Background(), "test-service")
		require.NoError(t, err)
		assert.Equal(t, svc.GetName(), got.GetName())

		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("GetService_NotFound", func(t *testing.T) {
		mock.ExpectQuery("SELECT config_json FROM upstream_services").
			WithArgs("unknown").
			WillReturnError(sql.ErrNoRows)

		got, err := store.GetService(context.Background(), "unknown")
		require.NoError(t, err)
		assert.Nil(t, got)

		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("ListServices", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"config_json"}).
			AddRow(`{"name":"test-service-1","id":"id-1"}`).
			AddRow(`{"name":"test-service-2","id":"id-2"}`)

		mock.ExpectQuery("SELECT config_json FROM upstream_services").
			WillReturnRows(rows)

		got, err := store.ListServices(context.Background())
		require.NoError(t, err)
		assert.Len(t, got, 2)

		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("DeleteService", func(t *testing.T) {
		mock.ExpectExec("DELETE FROM upstream_services").
			WithArgs("test-service").
			WillReturnResult(sqlmock.NewResult(1, 1))

		err := store.DeleteService(context.Background(), "test-service")
		require.NoError(t, err)

		require.NoError(t, mock.ExpectationsWereMet())
	})

	// Global Settings Tests
	t.Run("GetGlobalSettings", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"config_json"}).
			AddRow(`{"mcp_listen_address":":8080"}`)

		mock.ExpectQuery("SELECT config_json FROM global_settings").
			WillReturnRows(rows)

		got, err := store.GetGlobalSettings(context.Background())
		require.NoError(t, err)
		assert.Equal(t, ":8080", got.GetMcpListenAddress())

		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("GetGlobalSettings_NotFound", func(t *testing.T) {
		mock.ExpectQuery("SELECT config_json FROM global_settings").
			WillReturnError(sql.ErrNoRows)

		got, err := store.GetGlobalSettings(context.Background())
		require.NoError(t, err)
		assert.Nil(t, got)

		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("SaveGlobalSettings", func(t *testing.T) {
		settings := &configv1.GlobalSettings{
			McpListenAddress: stringPtr(":8080"),
		}

		mock.ExpectExec("INSERT INTO global_settings").
			WithArgs(sqlmock.AnyArg()).
			WillReturnResult(sqlmock.NewResult(1, 1))

		err := store.SaveGlobalSettings(context.Background(), settings)
		require.NoError(t, err)

		require.NoError(t, mock.ExpectationsWereMet())
	})

	// User Tests
	t.Run("CreateUser", func(t *testing.T) {
		user := &configv1.User{
			Id:    stringPtr("user-1"),
			Roles: []string{"admin"},
		}

		mock.ExpectExec("INSERT INTO users").
			WithArgs("user-1", sqlmock.AnyArg()).
			WillReturnResult(sqlmock.NewResult(1, 1))

		err := store.CreateUser(context.Background(), user)
		require.NoError(t, err)

		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("CreateUser_NoID", func(t *testing.T) {
		user := &configv1.User{
			Roles: []string{"admin"},
		}
		err := store.CreateUser(context.Background(), user)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "user ID is required")
	})

	t.Run("GetUser", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"config_json"}).
			AddRow(`{"id":"user-1","roles":["admin"]}`)

		mock.ExpectQuery("SELECT config_json FROM users").
			WithArgs("user-1").
			WillReturnRows(rows)

		got, err := store.GetUser(context.Background(), "user-1")
		require.NoError(t, err)
		assert.Equal(t, "user-1", got.GetId())

		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("GetUser_NotFound", func(t *testing.T) {
		mock.ExpectQuery("SELECT config_json FROM users").
			WithArgs("unknown").
			WillReturnError(sql.ErrNoRows)

		got, err := store.GetUser(context.Background(), "unknown")
		require.NoError(t, err)
		assert.Nil(t, got)

		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("ListUsers", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"config_json"}).
			AddRow(`{"id":"user-1","roles":["admin"]}`).
			AddRow(`{"id":"user-2","roles":["viewer"]}`)

		mock.ExpectQuery("SELECT config_json FROM users").
			WillReturnRows(rows)

		got, err := store.ListUsers(context.Background())
		require.NoError(t, err)
		assert.Len(t, got, 2)
		assert.Equal(t, "user-1", got[0].GetId())

		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("UpdateUser", func(t *testing.T) {
		user := &configv1.User{
			Id:    stringPtr("user-1"),
			Roles: []string{"editor"},
		}

		mock.ExpectExec("UPDATE users").
			WithArgs("user-1", sqlmock.AnyArg()).
			WillReturnResult(sqlmock.NewResult(1, 1))

		err := store.UpdateUser(context.Background(), user)
		require.NoError(t, err)

		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("UpdateUser_NotFound", func(t *testing.T) {
		user := &configv1.User{
			Id:    stringPtr("user-1"),
			Roles: []string{"editor"},
		}

		mock.ExpectExec("UPDATE users").
			WithArgs("user-1", sqlmock.AnyArg()).
			WillReturnResult(sqlmock.NewResult(0, 0))

		err := store.UpdateUser(context.Background(), user)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "user not found")

		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("DeleteUser", func(t *testing.T) {
		mock.ExpectExec("DELETE FROM users").
			WithArgs("user-1").
			WillReturnResult(sqlmock.NewResult(1, 1))

		err := store.DeleteUser(context.Background(), "user-1")
		require.NoError(t, err)

		require.NoError(t, mock.ExpectationsWereMet())
	})

	// Secrets Tests
	t.Run("ListSecrets", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"config_json"}).
			AddRow(`{"id":"sec-1","name":"secret1"}`).
			AddRow(`{"id":"sec-2","name":"secret2"}`)

		mock.ExpectQuery("SELECT config_json FROM secrets").
			WillReturnRows(rows)

		got, err := store.ListSecrets(context.Background())
		require.NoError(t, err)
		assert.Len(t, got, 2)

		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("GetSecret", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"config_json"}).
			AddRow(`{"id":"sec-1","name":"secret1"}`)

		mock.ExpectQuery("SELECT config_json FROM secrets").
			WithArgs("sec-1").
			WillReturnRows(rows)

		got, err := store.GetSecret(context.Background(), "sec-1")
		require.NoError(t, err)
		assert.Equal(t, "secret1", got.GetName())

		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("SaveSecret", func(t *testing.T) {
		secret := &configv1.Secret{
			Id:   stringPtr("sec-1"),
			Name: stringPtr("secret1"),
			Key:  stringPtr("key1"),
		}

		mock.ExpectExec("INSERT INTO secrets").
			WithArgs("sec-1", "secret1", "key1", sqlmock.AnyArg()).
			WillReturnResult(sqlmock.NewResult(1, 1))

		err := store.SaveSecret(context.Background(), secret)
		require.NoError(t, err)

		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("SaveSecret_NoID", func(t *testing.T) {
		secret := &configv1.Secret{
			Name: stringPtr("secret1"),
		}
		err := store.SaveSecret(context.Background(), secret)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "secret id is required")
	})

	t.Run("DeleteSecret", func(t *testing.T) {
		mock.ExpectExec("DELETE FROM secrets").
			WithArgs("sec-1").
			WillReturnResult(sqlmock.NewResult(1, 1))

		err := store.DeleteSecret(context.Background(), "sec-1")
		require.NoError(t, err)

		require.NoError(t, mock.ExpectationsWereMet())
	})

	// Profiles Tests
	t.Run("ListProfiles", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"config_json"}).
			AddRow(`{"name":"profile1"}`).
			AddRow(`{"name":"profile2"}`)

		mock.ExpectQuery("SELECT config_json FROM profile_definitions").
			WillReturnRows(rows)

		got, err := store.ListProfiles(context.Background())
		require.NoError(t, err)
		assert.Len(t, got, 2)

		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("GetProfile", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"config_json"}).
			AddRow(`{"name":"profile1"}`)

		mock.ExpectQuery("SELECT config_json FROM profile_definitions").
			WithArgs("profile1").
			WillReturnRows(rows)

		got, err := store.GetProfile(context.Background(), "profile1")
		require.NoError(t, err)
		assert.Equal(t, "profile1", got.GetName())

		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("SaveProfile", func(t *testing.T) {
		profile := &configv1.ProfileDefinition{
			Name: stringPtr("profile1"),
		}

		mock.ExpectExec("INSERT INTO profile_definitions").
			WithArgs("profile1", "profile1", sqlmock.AnyArg()).
			WillReturnResult(sqlmock.NewResult(1, 1))

		err := store.SaveProfile(context.Background(), profile)
		require.NoError(t, err)

		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("DeleteProfile", func(t *testing.T) {
		mock.ExpectExec("DELETE FROM profile_definitions").
			WithArgs("profile1").
			WillReturnResult(sqlmock.NewResult(1, 1))

		err := store.DeleteProfile(context.Background(), "profile1")
		require.NoError(t, err)

		require.NoError(t, mock.ExpectationsWereMet())
	})

	// Service Collections Tests
	t.Run("ListServiceCollections", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"config_json"}).
			AddRow(`{"name":"col1"}`).
			AddRow(`{"name":"col2"}`)

		mock.ExpectQuery("SELECT config_json FROM service_collections").
			WillReturnRows(rows)

		got, err := store.ListServiceCollections(context.Background())
		require.NoError(t, err)
		assert.Len(t, got, 2)

		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("GetServiceCollection", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"config_json"}).
			AddRow(`{"name":"col1"}`)

		mock.ExpectQuery("SELECT config_json FROM service_collections").
			WithArgs("col1").
			WillReturnRows(rows)

		got, err := store.GetServiceCollection(context.Background(), "col1")
		require.NoError(t, err)
		assert.Equal(t, "col1", got.GetName())

		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("SaveServiceCollection", func(t *testing.T) {
		col := &configv1.UpstreamServiceCollectionShare{
			Name: stringPtr("col1"),
		}

		mock.ExpectExec("INSERT INTO service_collections").
			WithArgs("col1", "col1", sqlmock.AnyArg()).
			WillReturnResult(sqlmock.NewResult(1, 1))

		err := store.SaveServiceCollection(context.Background(), col)
		require.NoError(t, err)

		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("DeleteServiceCollection", func(t *testing.T) {
		mock.ExpectExec("DELETE FROM service_collections").
			WithArgs("col1").
			WillReturnResult(sqlmock.NewResult(1, 1))

		err := store.DeleteServiceCollection(context.Background(), "col1")
		require.NoError(t, err)

		require.NoError(t, mock.ExpectationsWereMet())
	})
}

// initSchema test
func TestInitSchema(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	mock.ExpectExec("CREATE TABLE IF NOT EXISTS upstream_services").
		WillReturnResult(sqlmock.NewResult(0, 0))

	err = initSchema(context.Background(), db)
	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}
