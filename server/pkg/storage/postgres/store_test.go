// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package postgres

import (
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func TestPostgresStore_SaveAndGet(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	store := &Store{db: &DB{db}}

	svc := &configv1.UpstreamServiceConfig{
		Name:          proto.String("test-service"),
		Id:            proto.String("test-id"),
		SanitizedName: proto.String("test_service"),
	}

	// Expect SaveService
	mock.ExpectExec("INSERT INTO upstream_services").
		WithArgs("test-id", "test-service", sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = store.SaveService(svc)
	require.NoError(t, err)

	// Expect GetService
	// Since we mock, we can't test the actual JSON unmarshalling from the DB properly unless we provide a valid JSON string.
	// But we can test that the query is executed correctly.
	mock.ExpectQuery("SELECT config_json FROM upstream_services").
		WithArgs("test-service").
		WillReturnRows(sqlmock.NewRows([]string{"config_json"}).AddRow(`{"name":"test-service","id":"test-id"}`))

	loadedSvc, err := store.GetService("test-service")
	require.NoError(t, err)
	assert.Equal(t, "test-service", loadedSvc.GetName())
	assert.Equal(t, "test-id", loadedSvc.GetId())

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgresStore_ListServices(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	store := &Store{db: &DB{db}}

	mock.ExpectQuery("SELECT config_json FROM upstream_services").
		WillReturnRows(sqlmock.NewRows([]string{"config_json"}).
			AddRow(`{"name":"svc1","id":"id1"}`).
			AddRow(`{"name":"svc2","id":"id2"}`))

	services, err := store.ListServices()
	require.NoError(t, err)
	assert.Len(t, services, 2)
	assert.Equal(t, "svc1", services[0].GetName())
	assert.Equal(t, "svc2", services[1].GetName())

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgresStore_DeleteService(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	store := &Store{db: &DB{db}}

	mock.ExpectExec("DELETE FROM upstream_services").
		WithArgs("test-service").
		WillReturnResult(sqlmock.NewResult(0, 1))

	err = store.DeleteService("test-service")
	require.NoError(t, err)

	require.NoError(t, mock.ExpectationsWereMet())
}
