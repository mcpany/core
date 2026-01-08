// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package postgres

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	configv1 "github.com/mcpany/core/proto/config/v1"
	"google.golang.org/protobuf/proto"
)

func TestPostgresStore(t *testing.T) {
	// Create a new mock database
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	// Wrap sql.DB in our DB struct
	pgDB := &DB{db}
	store := NewStore(pgDB)

	t.Run("SaveService", func(t *testing.T) {
		svc := &configv1.UpstreamServiceConfig{
			Name: proto.String("test-service"),
			Id:   proto.String("test-id"),
		}

		mock.ExpectExec("INSERT INTO upstream_services").
			WithArgs("test-id", "test-service", sqlmock.AnyArg()).
			WillReturnResult(sqlmock.NewResult(1, 1))

		err := store.SaveService(context.Background(), svc)
		if err != nil {
			t.Errorf("error was not expected while updating stats: %s", err)
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("there were unmet expectations: %s", err)
		}
	})

	t.Run("ListUsers", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"config_json"}).
			AddRow(`{"id":"user-1","roles":["admin"]}`).
			AddRow(`{"id":"user-2","roles":["viewer"]}`)

		mock.ExpectQuery("SELECT config_json FROM users").
			WillReturnRows(rows)

		got, err := store.ListUsers(context.Background())
		if err != nil {
			t.Errorf("error was not expected while listing users: %s", err)
		}
		if len(got) != 2 {
			t.Errorf("expected 2 users, got %d", len(got))
		}
		if got[0].GetId() != "user-1" {
			t.Errorf("expected user-1, got %s", got[0].GetId())
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("there were unmet expectations: %s", err)
		}
	})

	t.Run("GetService", func(t *testing.T) {
		svc := &configv1.UpstreamServiceConfig{
			Name: proto.String("test-service"),
			Id:   proto.String("test-id"),
		}
		// Assuming json marshaling works as expected
		// We mock the return row
		rows := sqlmock.NewRows([]string{"config_json"}).
			AddRow(`{"name":"test-service","id":"test-id"}`)

		mock.ExpectQuery("SELECT config_json FROM upstream_services").
			WithArgs("test-service").
			WillReturnRows(rows)

		got, err := store.GetService(context.Background(), "test-service")
		if err != nil {
			t.Errorf("error was not expected while getting service: %s", err)
		}
		if got.GetName() != svc.GetName() {
			t.Errorf("expected name %s, got %s", svc.GetName(), got.GetName())
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("there were unmet expectations: %s", err)
		}
	})

	t.Run("DeleteService", func(t *testing.T) {
		mock.ExpectExec("DELETE FROM upstream_services").
			WithArgs("test-service").
			WillReturnResult(sqlmock.NewResult(1, 1))

		err := store.DeleteService(context.Background(), "test-service")
		if err != nil {
			t.Errorf("error was not expected while deleting service: %s", err)
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("there were unmet expectations: %s", err)
		}
	})

	t.Run("ListServices", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"config_json"}).
			AddRow(`{"name":"test-service-1","id":"id-1"}`).
			AddRow(`{"name":"test-service-2","id":"id-2"}`)

		mock.ExpectQuery("SELECT config_json FROM upstream_services").
			WillReturnRows(rows)

		got, err := store.ListServices(context.Background())
		if err != nil {
			t.Errorf("error was not expected while listing services: %s", err)
		}
		if len(got) != 2 {
			t.Errorf("expected 2 services, got %d", len(got))
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("there were unmet expectations: %s", err)
		}
	})
}

// initSchema test
func TestInitSchema(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	mock.ExpectExec("CREATE TABLE IF NOT EXISTS upstream_services").
		WillReturnResult(sqlmock.NewResult(0, 0))

	err = initSchema(db)
	if err != nil {
		t.Errorf("error was not expected while init schema: %s", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unmet expectations: %s", err)
	}
}

// Connection failure test (simulated)
func TestNewDB_PingFailure(t *testing.T) {
	// Not easy to mock NewDB internal sql.Open with sqlmock since it opens a real driver or we need to register a mock driver.
	// sqlmock registers itself as a driver if we use sqlmock.New(), but NewDB calls sql.Open("postgres", ...).
	// To test NewDB properly with sqlmock, we would need to dependency inject the opener or use "sqlmock" as driver name.
	// Since NewDB hardcodes "postgres", we skip this unit test or refactor NewDB.
	// For now, skipping NewDB test that requires real connection or driver injection.
}
