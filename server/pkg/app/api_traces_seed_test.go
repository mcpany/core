// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"context"
	"testing"
	"time"

	"github.com/mcpany/core/server/pkg/audit"
)

func TestSeedTraces_Custom(t *testing.T) {
	app, mid := setupTracesTestApp(t)

	// Initial seed shouldn't panic, writes to db
	err := app.seedTraces(context.Background())
	if err != nil {
		t.Fatalf("expected no error seeding traces, got %v", err)
	}

	history := mid.GetHistory()
	if len(history) == 0 {
		t.Fatalf("expected seeded trace history, got empty")
	}

	// Check that specific mocked entries are mapped correctly
	var hasRefactor bool
	for _, entryRaw := range history {
		if entry, ok := entryRaw.(audit.Entry); ok {
			if entry.ToolName == "code-refactor" {
				hasRefactor = true
			}
		}
	}
	if !hasRefactor {
		t.Errorf("expected 'code-refactor' tool in seed traces")
	}
}

func TestSeedTracesWithMockMiddleware_Custom(t *testing.T) {
	app, mid := setupTracesTestApp(t)

	err := app.seedTraces(context.Background())
	if err != nil {
		t.Fatalf("expected no error seeding traces, got %v", err)
	}

	// Give broadcast channel a ms to drain
	time.Sleep(10 * time.Millisecond)

	history := mid.GetHistory()
	if len(history) != 5 { // Seed generates 5 entries
		t.Fatalf("expected seeded trace history of 5 entries, got %d", len(history))
	}
}
