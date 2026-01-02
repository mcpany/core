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

	t.Run("SaveService_Error", func(t *testing.T) {
		svc := &configv1.UpstreamServiceConfig{
			Name: proto.String("test-service"),
		}

		mock.ExpectExec("INSERT INTO upstream_services").
			WillReturnError(errors.New("db error"))

		err := store.SaveService(context.Background(), svc)
		if err == nil {
			t.Errorf("expected error")
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("there were unmet expectations: %s", err)
		}
	})

	t.Run("SaveService_EmptyName", func(t *testing.T) {
		svc := &configv1.UpstreamServiceConfig{}
		err := store.SaveService(context.Background(), svc)
		if err == nil {
			t.Errorf("expected error for empty name")
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

	t.Run("GetService_NotFound", func(t *testing.T) {
		mock.ExpectQuery("SELECT config_json FROM upstream_services").
			WithArgs("unknown").
			WillReturnError(sql.ErrNoRows)

		got, err := store.GetService(context.Background(), "unknown")
		if err != nil {
			t.Errorf("expected nil error for not found, got %s", err)
		}
		if got != nil {
			t.Errorf("expected nil service")
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("there were unmet expectations: %s", err)
		}
	})

	t.Run("GetService_ScanError", func(t *testing.T) {
		mock.ExpectQuery("SELECT config_json FROM upstream_services").
			WithArgs("error").
			WillReturnError(errors.New("scan error"))

		_, err := store.GetService(context.Background(), "error")
		if err == nil {
			t.Errorf("expected error")
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("there were unmet expectations: %s", err)
		}
	})

	t.Run("GetService_UnmarshalError", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"config_json"}).
			AddRow(`invalid-json`)

		mock.ExpectQuery("SELECT config_json FROM upstream_services").
			WithArgs("bad-json").
			WillReturnRows(rows)

		_, err := store.GetService(context.Background(), "bad-json")
		if err == nil {
			t.Errorf("expected error")
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

	t.Run("DeleteService_Error", func(t *testing.T) {
		mock.ExpectExec("DELETE FROM upstream_services").
			WithArgs("test-service").
			WillReturnError(errors.New("delete error"))

		err := store.DeleteService(context.Background(), "test-service")
		if err == nil {
			t.Errorf("expected error")
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

	t.Run("ListServices_QueryError", func(t *testing.T) {
		mock.ExpectQuery("SELECT config_json FROM upstream_services").
			WillReturnError(errors.New("query error"))

		_, err := store.ListServices(context.Background())
		if err == nil {
			t.Errorf("expected error")
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("there were unmet expectations: %s", err)
		}
	})

	t.Run("ListServices_ScanError", func(t *testing.T) {
		// Simulate scan error by mismatching columns/types?
		// Or using sqlmock RowError?
		// sqlmock doesn't easily simulate Scan error on valid row unless type mismatch?
		// Or we can simulate JSON unmarshal error.
		rows := sqlmock.NewRows([]string{"config_json"}).
			AddRow(`invalid-json`)

		mock.ExpectQuery("SELECT config_json FROM upstream_services").
			WillReturnRows(rows)

		_, err := store.ListServices(context.Background())
		if err == nil {
			t.Errorf("expected error")
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

func TestInitSchema_Error(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	mock.ExpectExec("CREATE TABLE IF NOT EXISTS upstream_services").
		WillReturnError(errors.New("create table error"))

	err = initSchema(db)
	if err == nil {
		t.Errorf("expected error")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unmet expectations: %s", err)
	}
}
