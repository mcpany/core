// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package sql

import (
	"database/sql"
	"os"
	"path/filepath"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"google.golang.org/protobuf/proto"
	_ "modernc.org/sqlite" // Ensure sqlite driver is available for database/sql
)

func TestSaveServiceValidation(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "mcpany-test-validation-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	dbPath := filepath.Join(tmpDir, "test.db")
	store, err := NewStore("sqlite", dbPath)
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}
	defer store.Close()

	// Test case: Empty name
	svc := &configv1.UpstreamServiceConfig{
		Name: proto.String(""),
		Id:   proto.String("test-id"),
	}
	err = store.SaveService(svc)
	if err == nil {
		t.Error("expected error for empty service name, got nil")
	} else if err.Error() != "service name is required" {
		t.Errorf("expected error 'service name is required', got '%v'", err)
	}
}

func TestNewStoreErrors(t *testing.T) {
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

	_, err = NewStore("sqlite", dbPath)
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

	// Open store to initialize DB and tables
	store, err := NewStore("sqlite", dbPath)
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}
	store.Close()

	// Manually insert invalid JSON using raw sql
	rawDB, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Skipf("skipping test because failing to open raw sqlite db: %v", err)
	}
	defer rawDB.Close()

	// Note: invalid json in 'config' column
	_, err = rawDB.Exec("INSERT INTO upstream_services (id, name, config, created_at, updated_at) VALUES (?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)", "bad-id", "bad-service", "{invalid-json")
	if err != nil {
		// Attempt without checking exact schema?
		// Gorm uses 'upstream_services' table and 'config' column.
		t.Fatalf("failed to insert invalid data: %v", err)
	}
	rawDB.Close()

	// Re-open store to Load
	store, err = NewStore("sqlite", dbPath)
	if err != nil {
		t.Fatalf("failed to re-open store: %v", err)
	}
	defer store.Close()

	// Try to load
	config, err := store.Load()
	if err != nil {
		t.Fatalf("expected nil error when loading invalid JSON (should be skipped), got %v", err)
	}
	if len(config.UpstreamServices) != 0 {
		t.Errorf("expected 0 services, got %d", len(config.UpstreamServices))
	}

	// Try GetService - this SHOULD fail because it targets the specific invalid item
	_, err = store.GetService("bad-service")
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
	store, err := NewStore("sqlite", dbPath)
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}

	// Close the DB to force errors
	store.Close()

	// Test Load
	_, err = store.Load()
	if err == nil {
		t.Error("expected error on Load with closed DB, got nil")
	}

	// Test SaveService
	svc := &configv1.UpstreamServiceConfig{
		Name: proto.String("test-service"),
	}
	err = store.SaveService(svc)
	if err == nil {
		t.Error("expected error on SaveService with closed DB, got nil")
	}

	// Test GetService
	_, err = store.GetService("test-service")
	if err == nil {
		t.Error("expected error on GetService with closed DB, got nil")
	}

	// Test DeleteService
	err = store.DeleteService("test-service")
	if err == nil {
		t.Error("expected error on DeleteService with closed DB, got nil")
	}
}
