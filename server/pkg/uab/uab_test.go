// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package uab

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestUniversalAgentBus(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "uab_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	dbPath := filepath.Join(tempDir, "uab.db")
	uab, err := NewUniversalAgentBus(dbPath)
	if err != nil {
		t.Fatalf("failed to initialize UAB: %v", err)
	}
	defer uab.Close()

	ctx := context.Background()

	_ = uab.RegisterSession(ctx, "session-1")
	_ = uab.RegisterSession(ctx, "session-2")
	_ = uab.RegisterTransport(ctx, "transport-a")

	sessions, err := uab.GetSessionCount(ctx)
	if err != nil {
		t.Fatalf("failed to get sessions: %v", err)
	}
	if sessions != 2 {
		t.Errorf("Expected 2, got %d", sessions)
	}

	transports, err := uab.GetTransportCount(ctx)
	if err != nil {
		t.Fatalf("failed to get transports: %v", err)
	}
	if transports != 1 {
		t.Errorf("Expected 1, got %d", transports)
	}
}
