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

func TestPostgresStore_ServiceTemplates(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	pgDB := &DB{db}
	store := NewStore(pgDB)

	t.Run("ListServiceTemplates", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"config_json"}).
			AddRow(`{"name":"template-1","id":"id-1"}`).
			AddRow(`{"name":"template-2","id":"id-2"}`)

		mock.ExpectQuery("SELECT config_json FROM service_templates").
			WillReturnRows(rows)

		got, err := store.ListServiceTemplates(context.Background())
		require.NoError(t, err)
		assert.Len(t, got, 2)
		assert.Equal(t, "template-1", got[0].GetName())
		assert.Equal(t, "id-1", got[0].GetId())
		assert.Equal(t, "template-2", got[1].GetName())
		assert.Equal(t, "id-2", got[1].GetId())

		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("ListServiceTemplates_QueryError", func(t *testing.T) {
		mock.ExpectQuery("SELECT config_json FROM service_templates").
			WillReturnError(errors.New("query error"))

		_, err := store.ListServiceTemplates(context.Background())
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to query service_templates")

		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("ListServiceTemplates_ScanError", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"config_json"}).
			AddRow([]byte("invalid-json"))

		mock.ExpectQuery("SELECT config_json FROM service_templates").
			WillReturnRows(rows)

		_, err := store.ListServiceTemplates(context.Background())
		assert.Error(t, err)
		// Error message depends on protojson behavior, usually "failed to unmarshal"
		assert.Contains(t, err.Error(), "failed to unmarshal service template")

		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("GetServiceTemplate", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"config_json"}).
			AddRow(`{"name":"template-1","id":"id-1"}`)

		mock.ExpectQuery("SELECT config_json FROM service_templates WHERE id = \\$1").
			WithArgs("id-1").
			WillReturnRows(rows)

		got, err := store.GetServiceTemplate(context.Background(), "id-1")
		require.NoError(t, err)
		assert.NotNil(t, got)
		assert.Equal(t, "template-1", got.GetName())
		assert.Equal(t, "id-1", got.GetId())

		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("GetServiceTemplate_NotFound", func(t *testing.T) {
		mock.ExpectQuery("SELECT config_json FROM service_templates WHERE id = \\$1").
			WithArgs("id-1").
			WillReturnError(sql.ErrNoRows)

		got, err := store.GetServiceTemplate(context.Background(), "id-1")
		require.NoError(t, err)
		assert.Nil(t, got)

		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("GetServiceTemplate_QueryError", func(t *testing.T) {
		mock.ExpectQuery("SELECT config_json FROM service_templates WHERE id = \\$1").
			WithArgs("id-1").
			WillReturnError(errors.New("query error"))

		_, err := store.GetServiceTemplate(context.Background(), "id-1")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to scan config_json")

		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("GetServiceTemplate_UnmarshalError", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"config_json"}).
			AddRow(`invalid-json`)

		mock.ExpectQuery("SELECT config_json FROM service_templates WHERE id = \\$1").
			WithArgs("id-1").
			WillReturnRows(rows)

		_, err := store.GetServiceTemplate(context.Background(), "id-1")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to unmarshal service template")

		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("SaveServiceTemplate", func(t *testing.T) {
		tmpl := configv1.ServiceTemplate_builder{
			Id:   proto.String("id-1"),
			Name: proto.String("template-1"),
		}.Build()

		mock.ExpectExec("INSERT INTO service_templates").
			WithArgs("id-1", "template-1", sqlmock.AnyArg()).
			WillReturnResult(sqlmock.NewResult(1, 1))

		err := store.SaveServiceTemplate(context.Background(), tmpl)
		require.NoError(t, err)

		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("SaveServiceTemplate_MissingID", func(t *testing.T) {
		tmpl := configv1.ServiceTemplate_builder{
			Name: proto.String("template-1"),
		}.Build()

		err := store.SaveServiceTemplate(context.Background(), tmpl)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "template ID is required")
	})

	t.Run("SaveServiceTemplate_ExecError", func(t *testing.T) {
		tmpl := configv1.ServiceTemplate_builder{
			Id:   proto.String("id-1"),
			Name: proto.String("template-1"),
		}.Build()

		mock.ExpectExec("INSERT INTO service_templates").
			WithArgs("id-1", "template-1", sqlmock.AnyArg()).
			WillReturnError(errors.New("exec error"))

		err := store.SaveServiceTemplate(context.Background(), tmpl)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to save service template")

		require.NoError(t, mock.ExpectationsWereMet())
	})
}
