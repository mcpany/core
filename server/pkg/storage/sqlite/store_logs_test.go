package sqlite

import (
	"context"
	"testing"
	"time"

	"github.com/mcpany/core/server/pkg/logging"
)

func TestStore_Logs(t *testing.T) {
	// Setup temporary DB
	tmpDir := t.TempDir()
	dbPath := tmpDir + "/test.db"
	db, err := NewDB(dbPath)
	if err != nil {
		t.Fatalf("Failed to create DB: %v", err)
	}
	defer db.Close()

	store := NewStore(db)
	ctx := context.Background()

	// 1. Save logs
	entry1 := &logging.LogEntry{
		ID:        "log-1",
		Timestamp: time.Now().UTC().Add(-2 * time.Minute).Format(time.RFC3339),
		Level:     "INFO",
		Message:   "Test log 1",
		Source:    "test-source",
		Metadata:  map[string]any{"foo": "bar"},
	}
	if err := store.SaveLog(ctx, entry1); err != nil {
		t.Fatalf("Failed to save log 1: %v", err)
	}

	entry2 := &logging.LogEntry{
		ID:        "log-2",
		Timestamp: time.Now().UTC().Add(-1 * time.Minute).Format(time.RFC3339),
		Level:     "ERROR",
		Message:   "Test log 2 error",
		Source:    "test-source",
	}
	if err := store.SaveLog(ctx, entry2); err != nil {
		t.Fatalf("Failed to save log 2: %v", err)
	}

	// 2. List all logs
	logs, err := store.ListLogs(ctx, 10, 0, "ALL", "ALL", "")
	if err != nil {
		t.Fatalf("Failed to list logs: %v", err)
	}
	if len(logs) != 2 {
		t.Errorf("Expected 2 logs, got %d", len(logs))
	}
	// Check order (DESC)
	if logs[0].ID != "log-2" {
		t.Errorf("Expected newest log first (log-2), got %s", logs[0].ID)
	}

	// 3. Filter by Level
	logs, err = store.ListLogs(ctx, 10, 0, "ERROR", "ALL", "")
	if err != nil {
		t.Fatalf("Failed to list logs by level: %v", err)
	}
	if len(logs) != 1 {
		t.Errorf("Expected 1 error log, got %d", len(logs))
	}
	if logs[0].ID != "log-2" {
		t.Errorf("Expected log-2, got %s", logs[0].ID)
	}

	// 4. Filter by Search
	logs, err = store.ListLogs(ctx, 10, 0, "ALL", "ALL", "log 1")
	if err != nil {
		t.Fatalf("Failed to search logs: %v", err)
	}
	if len(logs) != 1 {
		t.Errorf("Expected 1 log matching 'log 1', got %d", len(logs))
	}
	if logs[0].ID != "log-1" {
		t.Errorf("Expected log-1, got %s", logs[0].ID)
	}
}
