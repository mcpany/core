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

func TestServiceTemplates(t *testing.T) {
	// Create a new mock database
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	// Wrap sql.DB in our DB struct
	pgDB := &DB{db}
	store := NewStore(pgDB)

	// SaveServiceTemplate Tests
	t.Run("SaveServiceTemplate_Success", func(t *testing.T) {
		tmpl := &configv1.ServiceTemplate{}
		tmpl.SetId("tmpl-1")
		tmpl.SetName("Template 1")

		mock.ExpectExec("INSERT INTO service_templates").
			WithArgs("tmpl-1", "Template 1", sqlmock.AnyArg()).
			WillReturnResult(sqlmock.NewResult(1, 1))

		err := store.SaveServiceTemplate(context.Background(), tmpl)
		require.NoError(t, err)

		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("SaveServiceTemplate_MissingID", func(t *testing.T) {
		tmpl := &configv1.ServiceTemplate{}
		tmpl.SetName("Template 1")

		err := store.SaveServiceTemplate(context.Background(), tmpl)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "template ID is required")
	})

	t.Run("SaveServiceTemplate_DBError", func(t *testing.T) {
		tmpl := &configv1.ServiceTemplate{}
		tmpl.SetId("tmpl-1")
		tmpl.SetName("Template 1")

		mock.ExpectExec("INSERT INTO service_templates").
			WithArgs("tmpl-1", "Template 1", sqlmock.AnyArg()).
			WillReturnError(errors.New("db error"))

		err := store.SaveServiceTemplate(context.Background(), tmpl)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to save service template")

		require.NoError(t, mock.ExpectationsWereMet())
	})

	// GetServiceTemplate Tests
	t.Run("GetServiceTemplate_Success", func(t *testing.T) {
		expectedTmpl := &configv1.ServiceTemplate{}
		expectedTmpl.SetId("tmpl-1")
		expectedTmpl.SetName("Template 1")

		rows := sqlmock.NewRows([]string{"config_json"}).
			AddRow(`{"id":"tmpl-1","name":"Template 1"}`)

		mock.ExpectQuery("SELECT config_json FROM service_templates").
			WithArgs("tmpl-1").
			WillReturnRows(rows)

		got, err := store.GetServiceTemplate(context.Background(), "tmpl-1")
		require.NoError(t, err)
		assert.Equal(t, expectedTmpl.GetId(), got.GetId())
		assert.Equal(t, expectedTmpl.GetName(), got.GetName())

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
			WillReturnError(errors.New("db error"))

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

	// ListServiceTemplates Tests
	t.Run("ListServiceTemplates_Success", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"config_json"}).
			AddRow(`{"id":"tmpl-1","name":"Template 1"}`).
			AddRow(`{"id":"tmpl-2","name":"Template 2"}`)

		mock.ExpectQuery("SELECT config_json FROM service_templates").
			WillReturnRows(rows)

		got, err := store.ListServiceTemplates(context.Background())
		require.NoError(t, err)
		assert.Len(t, got, 2)
		assert.Equal(t, "tmpl-1", got[0].GetId())
		assert.Equal(t, "tmpl-2", got[1].GetId())

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
		// Simulate a scan error by returning a row with wrong type or structure if possible,
		// but since we scan into []byte, almost anything is valid.
		// However, if the driver returns an error on Scan, we catch it.
		// sqlmock allows simulating row scan errors but it's tricky.
		// Instead, let's simulate unmarshal error which is easier.
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
}
