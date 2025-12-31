// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package postgres

import (
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

func TestPostgresStore_SaveService(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	store := &Store{db: &DB{DB: db}}

	service := &configv1.UpstreamServiceConfig{
		Name: proto.String("test-service"),
		Id:   proto.String("test-id"),
	}
	opts := protojson.MarshalOptions{UseProtoNames: true}
	configJSON, err := opts.Marshal(service)
	require.NoError(t, err)

	mock.ExpectExec("INSERT INTO upstream_services").
		WithArgs("test-id", "test-service", string(configJSON)).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = store.SaveService(service)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgresStore_GetService(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	store := &Store{db: &DB{DB: db}}

	expectedService := &configv1.UpstreamServiceConfig{
		Name: proto.String("test-service"),
		Id:   proto.String("test-id"),
	}
	opts := protojson.MarshalOptions{UseProtoNames: true}
	configJSON, err := opts.Marshal(expectedService)
	require.NoError(t, err)

	rows := sqlmock.NewRows([]string{"config_json"}).
		AddRow(string(configJSON))

	mock.ExpectQuery("SELECT config_json FROM upstream_services").
		WithArgs("test-service").
		WillReturnRows(rows)

	service, err := store.GetService("test-service")
	assert.NoError(t, err)
	assert.NotNil(t, service)
	assert.Equal(t, expectedService.GetName(), service.GetName())
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgresStore_ListServices(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	store := &Store{db: &DB{DB: db}}

	expectedService := &configv1.UpstreamServiceConfig{
		Name: proto.String("test-service"),
		Id:   proto.String("test-id"),
	}
	opts := protojson.MarshalOptions{UseProtoNames: true}
	configJSON, err := opts.Marshal(expectedService)
	require.NoError(t, err)

	rows := sqlmock.NewRows([]string{"config_json"}).
		AddRow(string(configJSON))

	mock.ExpectQuery("SELECT config_json FROM upstream_services").
		WillReturnRows(rows)

	services, err := store.ListServices()
	assert.NoError(t, err)
	assert.Len(t, services, 1)
	assert.Equal(t, expectedService.GetName(), services[0].GetName())
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgresStore_DeleteService(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	store := &Store{db: &DB{DB: db}}

	mock.ExpectExec("DELETE FROM upstream_services").
		WithArgs("test-service").
		WillReturnResult(sqlmock.NewResult(0, 1))

	err = store.DeleteService("test-service")
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestNewDB(t *testing.T) {
	// Since we can't easily mock the actual connection string opening without a real DB or overwriting sql.Open,
	// we will just test that it fails with invalid driver if we were to try (but we can't easily do that here).
	// So we skip testing NewDB connectivity in unit tests, relying on integration/mock.
	// However, we can test initSchema if we pass a mock DB.
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	mock.ExpectExec("CREATE TABLE IF NOT EXISTS upstream_services").
		WillReturnResult(sqlmock.NewResult(0, 0))

	err = initSchema(db)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}
