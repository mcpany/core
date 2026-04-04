// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"context"
	"testing"
	"time"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/audit"
	"github.com/mcpany/core/server/pkg/middleware"
	"google.golang.org/protobuf/proto"
)

func TestSeedTraces(t *testing.T) {
	app := setupTestApp()

	// Need to initialize db and middleware
	mid := app.GetAuditMiddleware()
	if mid == nil {
		t.Skip("Audit middleware disabled in standard setupTestApp, skipping mock tests")
	}

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

func setupTestAppWithAudit(t *testing.T) *Application {
	app := setupTestApp()

	// Inject fake audit middleware
	auditCfg := configv1.AuditConfig_builder{
		Enabled: proto.Bool(true),
	}.Build()
	auditMid, _ := middleware.NewAuditMiddleware(auditCfg)
	app.standardMiddlewares = &middleware.StandardMiddlewares{Audit: auditMid}

	return app
}

func TestSeedTracesWithMockMiddleware(t *testing.T) {
	app := setupTestAppWithAudit(t)

	err := app.seedTraces(context.Background())
	if err != nil {
		t.Fatalf("expected no error seeding traces, got %v", err)
	}

	mid := app.GetAuditMiddleware()

	// Give broadcast channel a ms to drain
	time.Sleep(10 * time.Millisecond)

	history := mid.GetHistory()
	if len(history) != 5 { // Seed generates 5 entries
		t.Fatalf("expected seeded trace history of 5 entries, got %d", len(history))
	}
}
