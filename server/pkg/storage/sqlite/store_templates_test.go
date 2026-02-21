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
	// Setup logic replicated from store_test.go to ensure isolation
	tmpDir, err := os.MkdirTemp("", "mcpany-test-templates-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	dbPath := filepath.Join(tmpDir, "test_templates.db")
	db, err := NewDB(dbPath)
	if err != nil {
		t.Fatalf("failed to create db: %v", err)
	}
	defer db.Close()

	store := NewStore(db)
	ctx := context.Background()

	t.Run("SaveAndGet", func(t *testing.T) {
		tmpl := configv1.ServiceTemplate_builder{
			Id:          proto.String("tmpl-1"),
			Name:        proto.String("My Template"),
			Description: proto.String("A test template"),
			ServiceConfig: configv1.UpstreamServiceConfig_builder{
				Name: proto.String("my-service"),
			}.Build(),
		}.Build()

		// Save
		if err := store.SaveServiceTemplate(ctx, tmpl); err != nil {
			t.Fatalf("failed to save template: %v", err)
		}

		// Get
		got, err := store.GetServiceTemplate(ctx, "tmpl-1")
		if err != nil {
			t.Fatalf("failed to get template: %v", err)
		}
		if got == nil {
			t.Fatal("expected template, got nil")
		}
		if got.GetName() != "My Template" {
			t.Errorf("expected name 'My Template', got '%s'", got.GetName())
		}
		if got.GetDescription() != "A test template" {
			t.Errorf("expected description 'A test template', got '%s'", got.GetDescription())
		}
	})

	t.Run("Update", func(t *testing.T) {
		tmpl := configv1.ServiceTemplate_builder{
			Id:   proto.String("tmpl-update"),
			Name: proto.String("Original Name"),
		}.Build()

		if err := store.SaveServiceTemplate(ctx, tmpl); err != nil {
			t.Fatalf("failed to save template: %v", err)
		}

		// Update
		tmplUpdated := configv1.ServiceTemplate_builder{
			Id:   proto.String("tmpl-update"),
			Name: proto.String("Updated Name"),
		}.Build()

		if err := store.SaveServiceTemplate(ctx, tmplUpdated); err != nil {
			t.Fatalf("failed to update template: %v", err)
		}

		got, err := store.GetServiceTemplate(ctx, "tmpl-update")
		if err != nil {
			t.Fatalf("failed to get updated template: %v", err)
		}
		if got.GetName() != "Updated Name" {
			t.Errorf("expected updated name 'Updated Name', got '%s'", got.GetName())
		}
	})

	t.Run("List", func(t *testing.T) {
		// Clear existing (though separate subtests run in sequence on same DB in this file's setup)
		// Since we reuse the DB, we might have templates from previous runs.
		// Let's count current templates first.
		initialList, err := store.ListServiceTemplates(ctx)
		if err != nil {
			t.Fatalf("failed to list initial templates: %v", err)
		}
		initialCount := len(initialList)

		// Add two new templates
		t1 := configv1.ServiceTemplate_builder{Id: proto.String("list-1"), Name: proto.String("List 1")}.Build()
		t2 := configv1.ServiceTemplate_builder{Id: proto.String("list-2"), Name: proto.String("List 2")}.Build()

		if err := store.SaveServiceTemplate(ctx, t1); err != nil {
			t.Fatal(err)
		}
		if err := store.SaveServiceTemplate(ctx, t2); err != nil {
			t.Fatal(err)
		}

		list, err := store.ListServiceTemplates(ctx)
		if err != nil {
			t.Fatalf("failed to list templates: %v", err)
		}

		if len(list) != initialCount+2 {
			t.Errorf("expected %d templates, got %d", initialCount+2, len(list))
		}

		// Verify existence
		found1 := false
		found2 := false
		for _, t := range list {
			if t.GetId() == "list-1" {
				found1 = true
			}
			if t.GetId() == "list-2" {
				found2 = true
			}
		}
		if !found1 || !found2 {
			t.Error("failed to find added templates in list")
		}
	})

	t.Run("Delete", func(t *testing.T) {
		tmpl := configv1.ServiceTemplate_builder{Id: proto.String("del-1"), Name: proto.String("Delete Me")}.Build()
		if err := store.SaveServiceTemplate(ctx, tmpl); err != nil {
			t.Fatal(err)
		}

		// Verify exists
		got, err := store.GetServiceTemplate(ctx, "del-1")
		if err != nil || got == nil {
			t.Fatal("failed to verify template exists before delete")
		}

		// Delete
		if err := store.DeleteServiceTemplate(ctx, "del-1"); err != nil {
			t.Fatalf("failed to delete template: %v", err)
		}

		// Verify gone
		got, err = store.GetServiceTemplate(ctx, "del-1")
		if err != nil {
			t.Fatalf("failed to check template after delete: %v", err)
		}
		if got != nil {
			t.Errorf("expected nil after delete, got %v", got)
		}
	})

	t.Run("EdgeCases", func(t *testing.T) {
		// Empty ID
		emptyID := configv1.ServiceTemplate_builder{Name: proto.String("No ID")}.Build()
		err := store.SaveServiceTemplate(ctx, emptyID)
		if err == nil {
			t.Error("expected error for empty ID, got nil")
		} else if !strings.Contains(err.Error(), "template ID is required") {
			t.Errorf("expected error containing 'template ID is required', got '%v'", err)
		}

		// Get Non-Existent
		got, err := store.GetServiceTemplate(ctx, "non-existent")
		if err != nil {
			t.Errorf("expected no error for non-existent template, got %v", err)
		}
		if got != nil {
			t.Errorf("expected nil for non-existent template, got %v", got)
		}

		// Delete Non-Existent (should succeed/no-op)
		if err := store.DeleteServiceTemplate(ctx, "non-existent"); err != nil {
			t.Errorf("expected no error for deleting non-existent template, got %v", err)
		}
	})
}
