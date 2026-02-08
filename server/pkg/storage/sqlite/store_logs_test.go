// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package sqlite

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/mcpany/core/server/pkg/logging"
)

func TestStoreLogs(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "mcpany-test-logs-*")
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

	t.Run("SaveAndGetLogs", func(t *testing.T) {
		ctx := context.Background()

		// 1. Save logs
		logs := []logging.LogEntry{
			{
				ID:        "log-1",
				Timestamp: time.Now().Add(-2 * time.Minute).Format(time.RFC3339),
				Level:     "INFO",
				Message:   "First log",
				Source:    "test",
				Metadata:  map[string]any{"key": "value1"},
			},
			{
				ID:        "log-2",
				Timestamp: time.Now().Add(-1 * time.Minute).Format(time.RFC3339),
				Level:     "WARN",
				Message:   "Second log",
				Source:    "test",
				Metadata:  map[string]any{"key": "value2"},
			},
			{
				ID:        "log-3",
				Timestamp: time.Now().Format(time.RFC3339),
				Level:     "ERROR",
				Message:   "Third log",
				Source:    "test",
				Metadata:  map[string]any{"key": "value3"},
			},
		}

		for _, l := range logs {
			if err := store.SaveLog(ctx, l); err != nil {
				t.Fatalf("failed to save log %s: %v", l.ID, err)
			}
		}

		// 2. Get logs (limit 2)
		// Expecting recent 2 logs: log-3, log-2 (in chronological order: log-2, log-3)
		got, err := store.GetLogs(ctx, 2, 0)
		if err != nil {
			t.Fatalf("failed to get logs: %v", err)
		}

		if len(got) != 2 {
			t.Errorf("expected 2 logs, got %d", len(got))
		}

		if got[0].ID != "log-2" {
			t.Errorf("expected first log to be log-2, got %s", got[0].ID)
		}
		if got[1].ID != "log-3" {
			t.Errorf("expected second log to be log-3, got %s", got[1].ID)
		}

		// Check metadata
		if val, ok := got[0].Metadata["key"]; !ok || val != "value2" {
			t.Errorf("expected metadata key=value2, got %v", val)
		}

		// 3. Get logs with offset
		// Offset 1, limit 2.
		// Total logs: 3. Sorted DESC by time: log-3, log-2, log-1.
		// Offset 1: log-2, log-1.
		// Limit 2: log-2, log-1.
		// Reversed: log-1, log-2.
		gotOffset, err := store.GetLogs(ctx, 2, 1)
		if err != nil {
			t.Fatalf("failed to get logs with offset: %v", err)
		}
		if len(gotOffset) != 2 {
			t.Errorf("expected 2 logs with offset, got %d", len(gotOffset))
		}
		if gotOffset[0].ID != "log-1" {
			t.Errorf("expected first log to be log-1, got %s", gotOffset[0].ID)
		}
		if gotOffset[1].ID != "log-2" {
			t.Errorf("expected second log to be log-2, got %s", gotOffset[1].ID)
		}
	})
}
