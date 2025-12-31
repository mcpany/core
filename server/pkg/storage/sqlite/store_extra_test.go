// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package sqlite

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"google.golang.org/protobuf/proto"
)

func TestSaveServiceValidation(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "mcpany-test-validation-*")
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

	// Test case: Empty name
	svc := &configv1.UpstreamServiceConfig{
		Name: proto.String(""),
		Id:   proto.String("test-id"),
	}
	err = store.SaveService(context.Background(), svc)
	if err == nil {
		t.Error("expected error for empty service name, got nil")
	} else if err.Error() != "service name is required" {
		t.Errorf("expected error 'service name is required', got '%v'", err)
	}
}

func TestNewDBErrors(t *testing.T) {
	// Test case: Invalid path (directory creation failure)
	// We can try to create a DB in a read-only directory or a non-existent parent that we can't create
	// But in a sandbox, permissions are tricky.
	// We can try using a path that is actually a directory.
	tmpDir, err := os.MkdirTemp("", "mcpany-test-db-errors-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a directory where the file should be
	dbPath := filepath.Join(tmpDir, "directory-as-file")
	if err := os.Mkdir(dbPath, 0755); err != nil {
		t.Fatalf("failed to create directory: %v", err)
	}

	_, err = NewDB(dbPath)
	if err == nil {
		t.Error("expected error when opening a directory as a DB file, got nil")
	}
}

func TestLoadInvalidData(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "mcpany-test-invalid-data-*")
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

	// Manually insert invalid JSON
	_, err = db.Exec("INSERT INTO upstream_services (id, name, config_json) VALUES (?, ?, ?)", "bad-id", "bad-service", "{invalid-json")
	if err != nil {
		t.Fatalf("failed to insert invalid data: %v", err)
	}

	// Try to load
	_, err = store.Load(context.Background())
	if err == nil {
		t.Error("expected error when loading invalid JSON, got nil")
	} else {
		// Error message depends on protojson implementation, but should fail
		t.Logf("Got expected error: %v", err)
	}

	// Try GetService
	_, err = store.GetService(context.Background(), "bad-service")
	if err == nil {
		t.Error("expected error when getting service with invalid JSON, got nil")
	}
}

func TestDBErrors(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "mcpany-test-db-closed-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	dbPath := filepath.Join(tmpDir, "test.db")
	db, err := NewDB(dbPath)
	if err != nil {
		t.Fatalf("failed to create db: %v", err)
	}
	store := NewStore(db)

	// Close the DB to force errors
	db.Close()

	// Test Load
	_, err = store.Load(context.Background())
	if err == nil {
		t.Error("expected error on Load with closed DB, got nil")
	}

	// Test SaveService
	svc := &configv1.UpstreamServiceConfig{
		Name: proto.String("test-service"),
	}
	err = store.SaveService(context.Background(), svc)
	if err == nil {
		t.Error("expected error on SaveService with closed DB, got nil")
	}

	// Test GetService
	_, err = store.GetService(context.Background(), "test-service")
	if err == nil {
		t.Error("expected error on GetService with closed DB, got nil")
	}

	// Test DeleteService
	err = store.DeleteService(context.Background(), "test-service")
	if err == nil {
		t.Error("expected error on DeleteService with closed DB, got nil")
	}
}
