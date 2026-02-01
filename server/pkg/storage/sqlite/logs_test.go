// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package sqlite

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	loggingv1 "github.com/mcpany/core/proto/logging/v1"
)

func TestLogs(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "mcpany-logs-test-*")
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

	t.Run("SaveAndList", func(t *testing.T) {
		log := &loggingv1.LogEntry{
			Id:           "log-1",
			Timestamp:    time.Now().Format(time.RFC3339),
			Level:        "INFO",
			Message:      "Test message",
			Source:       "test-source",
			MetadataJson: `{"key":"value"}`,
		}

		if err := store.SaveLog(ctx, log); err != nil {
			t.Fatalf("failed to save log: %v", err)
		}

		// List all
		logs, err := store.ListLogs(ctx, nil)
		if err != nil {
			t.Fatalf("failed to list logs: %v", err)
		}
		if len(logs) != 1 {
			t.Errorf("expected 1 log, got %d", len(logs))
		}
		if logs[0].GetMessage() != "Test message" {
			t.Errorf("expected message 'Test message', got '%s'", logs[0].GetMessage())
		}
	})

	t.Run("Filter", func(t *testing.T) {
		// Add more logs
		store.SaveLog(ctx, &loggingv1.LogEntry{Id: "log-2", Timestamp: time.Now().Add(-1 * time.Minute).Format(time.RFC3339), Level: "ERROR", Message: "Error message", Source: "error-source"})
		store.SaveLog(ctx, &loggingv1.LogEntry{Id: "log-3", Timestamp: time.Now().Add(-2 * time.Minute).Format(time.RFC3339), Level: "INFO", Message: "Another info", Source: "test-source"})

		// Filter by Level
		filter := &loggingv1.LogFilter{Level: "ERROR"}
		logs, err := store.ListLogs(ctx, filter)
		if err != nil {
			t.Fatalf("failed to list logs with filter: %v", err)
		}
		if len(logs) != 1 {
			t.Errorf("expected 1 error log, got %d", len(logs))
		}
		if logs[0].GetId() != "log-2" {
			t.Errorf("expected log-2, got %s", logs[0].GetId())
		}

		// Filter by Source
		filter = &loggingv1.LogFilter{Source: "test-source"}
		logs, err = store.ListLogs(ctx, filter)
		if err != nil {
			t.Fatalf("failed to list logs with filter: %v", err)
		}
		// Expecting log-1 and log-3.
		// log-1 (latest), log-3 (older).
		// ListLogs returns oldest first (reversed).
		// Wait, my implementation reverses it?
		// "Reverse logs to be chronological (oldest to newest)"
		// Query: ORDER BY timestamp DESC (Newest, Oldest)
		// Reverse -> Oldest, Newest.
		// log-3 is older than log-1.
		// So order should be log-3, log-1.

		if len(logs) != 2 {
			t.Errorf("expected 2 logs, got %d", len(logs))
		}
		if len(logs) == 2 {
			if logs[0].GetId() != "log-3" {
				t.Errorf("expected first log to be log-3, got %s", logs[0].GetId())
			}
			if logs[1].GetId() != "log-1" {
				t.Errorf("expected second log to be log-1, got %s", logs[1].GetId())
			}
		}
	})
}
