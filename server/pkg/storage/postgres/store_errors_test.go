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
	"google.golang.org/protobuf/proto"
)

func TestPostgresStore_Load_Errors(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	pgDB := &DB{db}
	store := NewStore(pgDB)

	t.Run("QueryError", func(t *testing.T) {
		mock.MatchExpectationsInOrder(false)

		// Fail upstream_services
		mock.ExpectQuery("SELECT config_json FROM upstream_services").
			WillReturnError(errors.New("query error"))

		// Others might run, so allow them (return empty)
		mock.ExpectQuery("SELECT config_json FROM users").
			WillReturnRows(sqlmock.NewRows([]string{"config_json"}))
		mock.ExpectQuery("SELECT config_json FROM global_settings").
			WillReturnError(sql.ErrNoRows)
		mock.ExpectQuery("SELECT config_json FROM service_collections").
			WillReturnRows(sqlmock.NewRows([]string{"config_json"}))

		_, err := store.Load(context.Background())
		assert.Error(t, err)
		// We can't guarantee "query error" is the one returned if context cancellation causes others to error differently.
		// But usually the first error is returned by errgroup or context error.
		// If errgroup returns the first non-nil error from the goroutines, it should be "query error".
		// However, if another goroutine context is canceled, it might return context.Canceled?
		// errgroup usually returns the error returned by the function.
		assert.Contains(t, err.Error(), "query error")
	})

	t.Run("UnmarshalError", func(t *testing.T) {
		mock.MatchExpectationsInOrder(false)

		rows := sqlmock.NewRows([]string{"config_json"}).
			AddRow([]byte("invalid-json"))

		mock.ExpectQuery("SELECT config_json FROM upstream_services").
			WillReturnRows(rows)

		mock.ExpectQuery("SELECT config_json FROM users").
			WillReturnRows(sqlmock.NewRows([]string{"config_json"}))
		mock.ExpectQuery("SELECT config_json FROM global_settings").
			WillReturnError(sql.ErrNoRows)
		mock.ExpectQuery("SELECT config_json FROM service_collections").
			WillReturnRows(sqlmock.NewRows([]string{"config_json"}))

		_, err := store.Load(context.Background())
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to unmarshal service config")
	})

	t.Run("RowsIterationError", func(t *testing.T) {
		mock.MatchExpectationsInOrder(false)

		rows := sqlmock.NewRows([]string{"config_json"}).
			AddRow([]byte("{}")).
			RowError(0, errors.New("row error"))

		mock.ExpectQuery("SELECT config_json FROM upstream_services").
			WillReturnRows(rows)

		mock.ExpectQuery("SELECT config_json FROM users").
			WillReturnRows(sqlmock.NewRows([]string{"config_json"}))
		mock.ExpectQuery("SELECT config_json FROM global_settings").
			WillReturnError(sql.ErrNoRows)
		mock.ExpectQuery("SELECT config_json FROM service_collections").
			WillReturnRows(sqlmock.NewRows([]string{"config_json"}))

		_, err := store.Load(context.Background())
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "row error")
	})
}

func TestPostgresStore_GetService_Errors(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	pgDB := &DB{db}
	store := NewStore(pgDB)

	t.Run("ScanError_NotNoRows", func(t *testing.T) {
		mock.ExpectQuery("SELECT config_json FROM upstream_services").
			WithArgs("svc").
			WillReturnError(errors.New("db connection error"))

		_, err := store.GetService(context.Background(), "svc")
		assert.Error(t, err)
		assert.NotEqual(t, err, errors.New("db connection error")) // Wrapped
		assert.Contains(t, err.Error(), "failed to scan config_json")
	})

	t.Run("UnmarshalError", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"config_json"}).
			AddRow([]byte("invalid"))

		mock.ExpectQuery("SELECT config_json FROM upstream_services").
			WithArgs("svc").
			WillReturnRows(rows)

		_, err := store.GetService(context.Background(), "svc")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to unmarshal service config")
	})
}

func TestPostgresStore_GetGlobalSettings_Errors(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	pgDB := &DB{db}
	store := NewStore(pgDB)

	t.Run("UnmarshalError", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"config_json"}).
			AddRow([]byte("invalid"))

		mock.ExpectQuery("SELECT config_json FROM global_settings").
			WillReturnRows(rows)

		_, err := store.GetGlobalSettings(context.Background())
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to unmarshal global settings")
	})
}

func TestPostgresStore_SaveGlobalSettings_Errors(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	pgDB := &DB{db}
	store := NewStore(pgDB)

	t.Run("ExecError", func(t *testing.T) {
		mock.ExpectExec("INSERT INTO global_settings").
			WillReturnError(errors.New("exec error"))

		err := store.SaveGlobalSettings(context.Background(), configv1.GlobalSettings_builder{}.Build())
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to save global settings")
	})
}

func TestPostgresStore_User_Errors(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	pgDB := &DB{db}
	store := NewStore(pgDB)

	t.Run("ListUsers_QueryError", func(t *testing.T) {
		mock.ExpectQuery("SELECT config_json FROM users").
			WillReturnError(errors.New("query error"))
		_, err := store.ListUsers(context.Background())
		assert.Error(t, err)
	})

	t.Run("ListUsers_UnmarshalError", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"config_json"}).AddRow([]byte("invalid"))
		mock.ExpectQuery("SELECT config_json FROM users").WillReturnRows(rows)
		_, err := store.ListUsers(context.Background())
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to unmarshal user config")
	})

	t.Run("GetUser_UnmarshalError", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"config_json"}).AddRow([]byte("invalid"))
		mock.ExpectQuery("SELECT config_json FROM users").WillReturnRows(rows)
		_, err := store.GetUser(context.Background(), "u1")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to unmarshal user config")
	})

	t.Run("CreateUser_ExecError", func(t *testing.T) {
		mock.ExpectExec("INSERT INTO users").WillReturnError(errors.New("exec error"))
		err := store.CreateUser(context.Background(), configv1.User_builder{Id: proto.String("u1")}.Build())
		assert.Error(t, err)
	})

	t.Run("UpdateUser_ExecError", func(t *testing.T) {
		mock.ExpectExec("UPDATE users").WillReturnError(errors.New("exec error"))
		err := store.UpdateUser(context.Background(), configv1.User_builder{Id: proto.String("u1")}.Build())
		assert.Error(t, err)
	})

	t.Run("UpdateUser_RowsAffectedError", func(t *testing.T) {
		mock.ExpectExec("UPDATE users").WillReturnResult(sqlmock.NewErrorResult(errors.New("rows affected error")))
		err := store.UpdateUser(context.Background(), configv1.User_builder{Id: proto.String("u1")}.Build())
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to get rows affected")
	})

	t.Run("DeleteUser_ExecError", func(t *testing.T) {
		mock.ExpectExec("DELETE FROM users").WillReturnError(errors.New("exec error"))
		err := store.DeleteUser(context.Background(), "u1")
		assert.Error(t, err)
	})
}

func TestPostgresStore_Secret_Errors(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	pgDB := &DB{db}
	store := NewStore(pgDB)

	t.Run("ListSecrets_QueryError", func(t *testing.T) {
		mock.ExpectQuery("SELECT config_json FROM secrets").WillReturnError(errors.New("err"))
		_, err := store.ListSecrets(context.Background())
		assert.Error(t, err)
	})

	t.Run("ListSecrets_UnmarshalError", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"config_json"}).AddRow([]byte("invalid"))
		mock.ExpectQuery("SELECT config_json FROM secrets").WillReturnRows(rows)
		_, err := store.ListSecrets(context.Background())
		assert.Error(t, err)
	})

	t.Run("GetSecret_UnmarshalError", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"config_json"}).AddRow([]byte("invalid"))
		mock.ExpectQuery("SELECT config_json FROM secrets").WillReturnRows(rows)
		_, err := store.GetSecret(context.Background(), "s1")
		assert.Error(t, err)
	})

	t.Run("SaveSecret_ExecError", func(t *testing.T) {
		mock.ExpectExec("INSERT INTO secrets").WillReturnError(errors.New("err"))
		err := store.SaveSecret(context.Background(), configv1.Secret_builder{Id: proto.String("s1")}.Build())
		assert.Error(t, err)
	})

	t.Run("DeleteSecret_ExecError", func(t *testing.T) {
		mock.ExpectExec("DELETE FROM secrets").WillReturnError(errors.New("err"))
		err := store.DeleteSecret(context.Background(), "s1")
		assert.Error(t, err)
	})
}

func TestPostgresStore_Profile_Errors(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	pgDB := &DB{db}
	store := NewStore(pgDB)

	t.Run("ListProfiles_QueryError", func(t *testing.T) {
		mock.ExpectQuery("SELECT config_json FROM profile_definitions").WillReturnError(errors.New("err"))
		_, err := store.ListProfiles(context.Background())
		assert.Error(t, err)
	})

	t.Run("ListProfiles_UnmarshalError", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"config_json"}).AddRow([]byte("invalid"))
		mock.ExpectQuery("SELECT config_json FROM profile_definitions").WillReturnRows(rows)
		_, err := store.ListProfiles(context.Background())
		assert.Error(t, err)
	})

	t.Run("GetProfile_UnmarshalError", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"config_json"}).AddRow([]byte("invalid"))
		mock.ExpectQuery("SELECT config_json FROM profile_definitions").WillReturnRows(rows)
		_, err := store.GetProfile(context.Background(), "p1")
		assert.Error(t, err)
	})

	t.Run("SaveProfile_ExecError", func(t *testing.T) {
		mock.ExpectExec("INSERT INTO profile_definitions").WillReturnError(errors.New("err"))
		err := store.SaveProfile(context.Background(), configv1.ProfileDefinition_builder{Name: proto.String("p1")}.Build())
		assert.Error(t, err)
	})

	t.Run("DeleteProfile_ExecError", func(t *testing.T) {
		mock.ExpectExec("DELETE FROM profile_definitions").WillReturnError(errors.New("err"))
		err := store.DeleteProfile(context.Background(), "p1")
		assert.Error(t, err)
	})
}

func TestPostgresStore_Collection_Errors(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	pgDB := &DB{db}
	store := NewStore(pgDB)

	t.Run("ListCollections_QueryError", func(t *testing.T) {
		mock.ExpectQuery("SELECT config_json FROM service_collections").WillReturnError(errors.New("err"))
		_, err := store.ListServiceCollections(context.Background())
		assert.Error(t, err)
	})

	t.Run("ListCollections_UnmarshalError", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"config_json"}).AddRow([]byte("invalid"))
		mock.ExpectQuery("SELECT config_json FROM service_collections").WillReturnRows(rows)
		_, err := store.ListServiceCollections(context.Background())
		assert.Error(t, err)
	})

	t.Run("GetCollection_UnmarshalError", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"config_json"}).AddRow([]byte("invalid"))
		mock.ExpectQuery("SELECT config_json FROM service_collections").WillReturnRows(rows)
		_, err := store.GetServiceCollection(context.Background(), "c1")
		assert.Error(t, err)
	})

	t.Run("SaveCollection_ExecError", func(t *testing.T) {
		mock.ExpectExec("INSERT INTO service_collections").WillReturnError(errors.New("err"))
		err := store.SaveServiceCollection(context.Background(), configv1.Collection_builder{Name: proto.String("c1")}.Build())
		assert.Error(t, err)
	})

	t.Run("DeleteCollection_ExecError", func(t *testing.T) {
		mock.ExpectExec("DELETE FROM service_collections").WillReturnError(errors.New("err"))
		err := store.DeleteServiceCollection(context.Background(), "c1")
		assert.Error(t, err)
	})
}

func TestPostgresStore_Token_Errors(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	pgDB := &DB{db}
	store := NewStore(pgDB)

	t.Run("GetToken_UnmarshalError", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"config_json"}).AddRow([]byte("invalid"))
		mock.ExpectQuery("SELECT config_json FROM user_tokens").WillReturnRows(rows)
		_, err := store.GetToken(context.Background(), "u1", "s1")
		assert.Error(t, err)
	})

	t.Run("SaveToken_ExecError", func(t *testing.T) {
		mock.ExpectExec("INSERT INTO user_tokens").WillReturnError(errors.New("err"))
		err := store.SaveToken(context.Background(), configv1.UserToken_builder{UserId: proto.String("u1"), ServiceId: proto.String("s1")}.Build())
		assert.Error(t, err)
	})

	t.Run("DeleteToken_ExecError", func(t *testing.T) {
		mock.ExpectExec("DELETE FROM user_tokens").WillReturnError(errors.New("err"))
		err := store.DeleteToken(context.Background(), "u1", "s1")
		assert.Error(t, err)
	})
}

func TestHasConfigSources(t *testing.T) {
	store := NewStore(nil)
	assert.True(t, store.HasConfigSources())
}

func TestNewDBFromSQLDB_PingError(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
	require.NoError(t, err)
	defer db.Close()

	mock.ExpectPing().WillReturnError(errors.New("ping error"))

	pgDB, err := NewDBFromSQLDB(db)
	assert.Error(t, err)
	assert.Nil(t, pgDB)
	assert.Contains(t, err.Error(), "failed to ping db")
}

func TestNewDBFromSQLDB_SchemaError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	mock.ExpectPing()
	mock.ExpectExec("CREATE TABLE IF NOT EXISTS upstream_services").WillReturnError(errors.New("schema error"))

	pgDB, err := NewDBFromSQLDB(db)
	assert.Error(t, err)
	assert.Nil(t, pgDB)
	assert.Contains(t, err.Error(), "failed to init schema")
}
