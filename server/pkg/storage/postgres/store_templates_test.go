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
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func TestServiceTemplate(t *testing.T) {
	// Create a new mock database
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	// Wrap sql.DB in our DB struct
	pgDB := &DB{db}
	store := NewStore(pgDB)

	t.Run("SaveServiceTemplate", func(t *testing.T) {
		tmpl := configv1.ServiceTemplate_builder{
			Name: proto.String("template-1"),
			Id:   proto.String("tmpl-1"),
		}.Build()

		mock.ExpectExec("INSERT INTO service_templates").
			WithArgs("tmpl-1", "template-1", sqlmock.AnyArg()).
			WillReturnResult(sqlmock.NewResult(1, 1))

		err := store.SaveServiceTemplate(context.Background(), tmpl)
		require.NoError(t, err)

		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("SaveServiceTemplate_Error", func(t *testing.T) {
		tmpl := configv1.ServiceTemplate_builder{
			Name: proto.String("template-1"),
			Id:   proto.String("tmpl-1"),
		}.Build()

		mock.ExpectExec("INSERT INTO service_templates").
			WithArgs("tmpl-1", "template-1", sqlmock.AnyArg()).
			WillReturnError(errors.New("db error"))

		err := store.SaveServiceTemplate(context.Background(), tmpl)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to save service template")

		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("SaveServiceTemplate_MissingID", func(t *testing.T) {
		tmpl := configv1.ServiceTemplate_builder{
			Name: proto.String("template-1"),
		}.Build()

		err := store.SaveServiceTemplate(context.Background(), tmpl)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "template ID is required")
	})

	t.Run("GetServiceTemplate", func(t *testing.T) {
		tmpl := configv1.ServiceTemplate_builder{
			Name: proto.String("template-1"),
			Id:   proto.String("tmpl-1"),
		}.Build()

		rows := sqlmock.NewRows([]string{"config_json"}).
			AddRow(`{"name":"template-1","id":"tmpl-1"}`)

		mock.ExpectQuery("SELECT config_json FROM service_templates").
			WithArgs("tmpl-1").
			WillReturnRows(rows)

		got, err := store.GetServiceTemplate(context.Background(), "tmpl-1")
		require.NoError(t, err)
		assert.Equal(t, tmpl.GetName(), got.GetName())

		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("GetServiceTemplate_NotFound", func(t *testing.T) {
		mock.ExpectQuery("SELECT config_json FROM service_templates").
			WithArgs("unknown").
			WillReturnError(sql.ErrNoRows)

		got, err := store.GetServiceTemplate(context.Background(), "unknown")
		require.NoError(t, err)
		assert.Nil(t, got)

		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("GetServiceTemplate_DBError", func(t *testing.T) {
		mock.ExpectQuery("SELECT config_json FROM service_templates").
			WithArgs("tmpl-1").
			WillReturnError(errors.New("db connection error"))

		got, err := store.GetServiceTemplate(context.Background(), "tmpl-1")
		require.Error(t, err)
		assert.Nil(t, got)
		assert.Contains(t, err.Error(), "failed to scan config_json")

		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("GetServiceTemplate_UnmarshalError", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"config_json"}).
			AddRow(`invalid-json`)

		mock.ExpectQuery("SELECT config_json FROM service_templates").
			WithArgs("tmpl-1").
			WillReturnRows(rows)

		got, err := store.GetServiceTemplate(context.Background(), "tmpl-1")
		require.Error(t, err)
		assert.Nil(t, got)
		assert.Contains(t, err.Error(), "failed to unmarshal service template")

		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("ListServiceTemplates", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"config_json"}).
			AddRow(`{"name":"template-1","id":"tmpl-1"}`).
			AddRow(`{"name":"template-2","id":"tmpl-2"}`)

		mock.ExpectQuery("SELECT config_json FROM service_templates").
			WillReturnRows(rows)

		got, err := store.ListServiceTemplates(context.Background())
		require.NoError(t, err)
		assert.Len(t, got, 2)
		assert.Equal(t, "template-1", got[0].GetName())
		assert.Equal(t, "template-2", got[1].GetName())

		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("ListServiceTemplates_Empty", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"config_json"})

		mock.ExpectQuery("SELECT config_json FROM service_templates").
			WillReturnRows(rows)

		got, err := store.ListServiceTemplates(context.Background())
		require.NoError(t, err)
		assert.Len(t, got, 0)

		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("ListServiceTemplates_DBError", func(t *testing.T) {
		mock.ExpectQuery("SELECT config_json FROM service_templates").
			WillReturnError(errors.New("db error"))

		got, err := store.ListServiceTemplates(context.Background())
		require.Error(t, err)
		assert.Nil(t, got)
		assert.Contains(t, err.Error(), "failed to query service_templates")

		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("ListServiceTemplates_ScanError", func(t *testing.T) {
		// Simulate a scan error by defining columns that don't match or other sqlmock behavior
		// But simpler is to have Unmarshal fail or simulate row error.
		// Let's simulate Unmarshal error.
		rows := sqlmock.NewRows([]string{"config_json"}).
			AddRow(`invalid-json`)

		mock.ExpectQuery("SELECT config_json FROM service_templates").
			WillReturnRows(rows)

		got, err := store.ListServiceTemplates(context.Background())
		require.Error(t, err)
		assert.Nil(t, got)
		assert.Contains(t, err.Error(), "failed to unmarshal service template")

		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("ListServiceTemplates_RowError", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"config_json"}).
			AddRow(`{}`).
			RowError(0, errors.New("row error"))

		mock.ExpectQuery("SELECT config_json FROM service_templates").
			WillReturnRows(rows)

		got, err := store.ListServiceTemplates(context.Background())
		require.Error(t, err)
		assert.Nil(t, got)
		// sqlmock RowError might result in Scan error or Next returning false but Err being set.
		// The loop: for rows.Next() { scan } check Err().
		// If RowError happens on first row, Next() might return false, then Err() returns error.
		assert.Contains(t, err.Error(), "error iterating rows")

		require.NoError(t, mock.ExpectationsWereMet())
	})
}
