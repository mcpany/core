// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package sqlite

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func TestStore_ServiceTemplates(t *testing.T) {
	// Setup
	tempDir, err := os.MkdirTemp("", "mcpany_sqlite_test_templates")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	dbPath := filepath.Join(tempDir, "test.db")
	db, err := NewDB(dbPath)
	require.NoError(t, err)
	defer db.Close()

	store := NewStore(db)
	ctx := context.Background()

	t.Run("SaveAndGet", func(t *testing.T) {
		tmpl := configv1.ServiceTemplate_builder{
			Id:   proto.String("tpl-1"),
			Name: proto.String("My Template"),
			Description: proto.String("A test template"),
		}.Build()

		err := store.SaveServiceTemplate(ctx, tmpl)
		require.NoError(t, err)

		loaded, err := store.GetServiceTemplate(ctx, "tpl-1")
		require.NoError(t, err)
		require.NotNil(t, loaded)
		assert.Equal(t, "tpl-1", loaded.GetId())
		assert.Equal(t, "My Template", loaded.GetName())
		assert.Equal(t, "A test template", loaded.GetDescription())
	})

	t.Run("Update", func(t *testing.T) {
		// Create initial
		tmpl := configv1.ServiceTemplate_builder{
			Id:   proto.String("tpl-2"),
			Name: proto.String("Original"),
		}.Build()
		err := store.SaveServiceTemplate(ctx, tmpl)
		require.NoError(t, err)

		// Update
		updated := configv1.ServiceTemplate_builder{
			Id:   proto.String("tpl-2"),
			Name: proto.String("Updated"),
		}.Build()
		err = store.SaveServiceTemplate(ctx, updated)
		require.NoError(t, err)

		loaded, err := store.GetServiceTemplate(ctx, "tpl-2")
		require.NoError(t, err)
		assert.Equal(t, "Updated", loaded.GetName())
	})

	t.Run("List", func(t *testing.T) {
		// Clear existing (from previous runs potentially, but we use fresh DB per test file run, but shared within file? No, NewDB is called once per test func TestStore_ServiceTemplates)
		// Wait, I am using subtests. They share the DB instance `store`.
		// So `tpl-1` and `tpl-2` exist.

		list, err := store.ListServiceTemplates(ctx)
		require.NoError(t, err)
		assert.Len(t, list, 2)
	})

	t.Run("Delete", func(t *testing.T) {
		err := store.DeleteServiceTemplate(ctx, "tpl-1")
		require.NoError(t, err)

		loaded, err := store.GetServiceTemplate(ctx, "tpl-1")
		require.NoError(t, err)
		assert.Nil(t, loaded)

		list, err := store.ListServiceTemplates(ctx)
		require.NoError(t, err)
		assert.Len(t, list, 1) // tpl-2 remains
	})

	t.Run("GetNotFound", func(t *testing.T) {
		loaded, err := store.GetServiceTemplate(ctx, "non-existent")
		require.NoError(t, err)
		assert.Nil(t, loaded)
	})

	t.Run("SaveMissingID", func(t *testing.T) {
		tmpl := configv1.ServiceTemplate_builder{
			Name: proto.String("No ID"),
		}.Build()
		err := store.SaveServiceTemplate(ctx, tmpl)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "template ID is required")
	})
}
