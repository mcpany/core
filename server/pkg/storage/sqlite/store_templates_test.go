// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package sqlite

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"google.golang.org/protobuf/proto"
)

func TestServiceTemplates(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "mcpany-test-templates-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	dbPath := filepath.Join(tmpDir, "test.db")
	db, err := NewDB(dbPath)
	if err != nil {
		t.Fatalf("failed to create db: %v", err)
	}
	defer db.Close()

	store := NewStore(db)
	ctx := context.Background()

	t.Run("SaveAndLoad", func(t *testing.T) {
		tmpl := configv1.ServiceTemplate_builder{
			Id:          proto.String("tmpl-1"),
			Name:        proto.String("My Template"),
			Description: proto.String("A test template"),
			ServiceConfig: configv1.UpstreamServiceConfig_builder{
				Name: proto.String("test"),
			}.Build(),
		}.Build()

		err := store.SaveServiceTemplate(ctx, tmpl)
		if err != nil {
			t.Fatalf("failed to save template: %v", err)
		}

		loaded, err := store.GetServiceTemplate(ctx, "tmpl-1")
		if err != nil {
			t.Fatalf("failed to get template: %v", err)
		}
		if loaded == nil {
			t.Fatal("expected template to be found, got nil")
		}
		if loaded.GetName() != "My Template" {
			t.Errorf("expected name 'My Template', got %s", loaded.GetName())
		}
		if loaded.GetDescription() != "A test template" {
			t.Errorf("expected description 'A test template', got %s", loaded.GetDescription())
		}
	})

	t.Run("List", func(t *testing.T) {
		// Should have 1 from previous test
		list, err := store.ListServiceTemplates(ctx)
		if err != nil {
			t.Fatalf("failed to list templates: %v", err)
		}
		if len(list) != 1 {
			t.Errorf("expected 1 template, got %d", len(list))
		}

		// Add another
		tmpl2 := configv1.ServiceTemplate_builder{
			Id:   proto.String("tmpl-2"),
			Name: proto.String("Another Template"),
		}.Build()
		if err := store.SaveServiceTemplate(ctx, tmpl2); err != nil {
			t.Fatalf("failed to save second template: %v", err)
		}

		list, err = store.ListServiceTemplates(ctx)
		if err != nil {
			t.Fatalf("failed to list templates: %v", err)
		}
		if len(list) != 2 {
			t.Errorf("expected 2 templates, got %d", len(list))
		}
	})

	t.Run("Update", func(t *testing.T) {
		// Update tmpl-1
		tmpl := configv1.ServiceTemplate_builder{
			Id:   proto.String("tmpl-1"),
			Name: proto.String("Updated Template"),
		}.Build()

		if err := store.SaveServiceTemplate(ctx, tmpl); err != nil {
			t.Fatalf("failed to update template: %v", err)
		}

		loaded, err := store.GetServiceTemplate(ctx, "tmpl-1")
		if err != nil {
			t.Fatalf("failed to get updated template: %v", err)
		}
		if loaded.GetName() != "Updated Template" {
			t.Errorf("expected name 'Updated Template', got %s", loaded.GetName())
		}
	})

	t.Run("Delete", func(t *testing.T) {
		if err := store.DeleteServiceTemplate(ctx, "tmpl-1"); err != nil {
			t.Fatalf("failed to delete template: %v", err)
		}

		loaded, err := store.GetServiceTemplate(ctx, "tmpl-1")
		if err != nil {
			t.Fatalf("failed to check deleted template: %v", err)
		}
		if loaded != nil {
			t.Errorf("expected template to be nil (not found), got %v", loaded)
		}

		// List should show 1 remaining
		list, err := store.ListServiceTemplates(ctx)
		if err != nil {
			t.Fatalf("failed to list templates: %v", err)
		}
		if len(list) != 1 {
			t.Errorf("expected 1 template, got %d", len(list))
		}
	})

	t.Run("GetNonExistent", func(t *testing.T) {
		loaded, err := store.GetServiceTemplate(ctx, "non-existent")
		if err != nil {
			t.Fatalf("failed to check non-existent template: %v", err)
		}
		if loaded != nil {
			t.Errorf("expected nil for non-existent template")
		}
	})

	t.Run("SaveWithoutID", func(t *testing.T) {
		tmpl := configv1.ServiceTemplate_builder{
			Name: proto.String("No ID"),
		}.Build()
		err := store.SaveServiceTemplate(ctx, tmpl)
		if err == nil {
			t.Error("expected error saving template without ID, got nil")
		} else if !strings.Contains(err.Error(), "template ID is required") {
			t.Errorf("expected error containing 'template ID is required', got '%v'", err)
		}
	})
}
