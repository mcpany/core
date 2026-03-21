// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"context"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/mcpany/core/server/pkg/logging"
	"github.com/mcpany/core/server/pkg/storage/sqlite"
)

func TestLogPersistence_CatchUp(t *testing.T) {
	// Setup Logger
	logging.ForTestsOnlyResetLogger()
	logging.Init(slog.LevelInfo, os.Stderr, "")

	// 1. Generate logs BEFORE persistence starts
	log := logging.GetLogger()
	log.Info("Startup log 1")
	log.Info("Startup log 2")

	// Verify Broadcaster has them
	history := logging.GlobalBroadcaster.GetHistory()
	if len(history) < 2 {
		t.Fatalf("Broadcaster missing startup logs, got %d", len(history))
	}

	// 2. Setup SQLite DB
	db, err := sqlite.NewDB(":memory:")
	if err != nil {
		t.Fatalf("Failed to create memory db: %v", err)
	}
	defer db.Close()
	store := sqlite.NewStore(db)

	// 3. Start Persistence (Catch-up)
	app := NewApplication()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Call startLogPersistence directly
	app.startLogPersistence(ctx, store)

	// Wait for async flush
	time.Sleep(500 * time.Millisecond)

	// 4. Verify logs are in DB
	recentLogs, err := store.GetRecentLogs(ctx, 100)
	if err != nil {
		t.Fatalf("Failed to get logs from DB: %v", err)
	}

	found1 := false
	found2 := false
	for _, l := range recentLogs {
		if l.Message == "Startup log 1" {
			found1 = true
		}
		if l.Message == "Startup log 2" {
			found2 = true
		}
	}

	if !found1 {
		t.Error("Startup log 1 not persisted to DB")
	}
	if !found2 {
		t.Error("Startup log 2 not persisted to DB")
	}

	// 5. Log AFTER persistence started
	log.Info("Runtime log 3")
	time.Sleep(100 * time.Millisecond)

	recentLogs, _ = store.GetRecentLogs(ctx, 100)
	found3 := false
	for _, l := range recentLogs {
		if l.Message == "Runtime log 3" {
			found3 = true
		}
	}
	if !found3 {
		t.Error("Runtime log 3 not persisted to DB")
	}
}

func TestLogPersistence_Idempotency(t *testing.T) {
	// Setup Logger
	logging.ForTestsOnlyResetLogger()
	logging.Init(slog.LevelInfo, os.Stderr, "")

	// 1. Setup SQLite DB
	db, err := sqlite.NewDB(":memory:")
	if err != nil {
		t.Fatalf("Failed to create memory db: %v", err)
	}
	defer db.Close()
	store := sqlite.NewStore(db)

	// 2. Pre-populate DB with a log
	entry := &logging.LogEntry{
		ID:        "duplicate-id",
		Timestamp: time.Now().Format(time.RFC3339),
		Level:     "INFO",
		Message:   "Existing log",
	}
	if err := store.SaveLog(context.Background(), entry); err != nil {
		t.Fatalf("Failed to save initial log: %v", err)
	}

	// 3. Hydrate Broadcaster with this log (simulate initializeLogPersistence)
	logging.GlobalBroadcaster.Hydrate([]any{*entry})

	// 4. Start Persistence (Catch-up)
	// It should try to save "duplicate-id" again.
	app := NewApplication()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	app.startLogPersistence(ctx, store)

	// Wait for async flush
	time.Sleep(500 * time.Millisecond)

	// 5. Verify no crash and log still exists
	recentLogs, err := store.GetRecentLogs(ctx, 100)
	if err != nil {
		t.Fatalf("Failed to get logs from DB: %v", err)
	}

	count := 0
	for _, l := range recentLogs {
		if l.ID == "duplicate-id" {
			count++
		}
	}

	if count != 1 {
		t.Errorf("Expected 1 log with duplicate-id, got %d", count)
	}
}
