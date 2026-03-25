// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestBlackboardStore(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "blackboard_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	dbPath := filepath.Join(tempDir, "blackboard.db")

	store, err := NewBlackboardStore(dbPath, "agent_aware")
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}
	defer store.Close()

	ctx := context.Background()

	// Test Set
	err = store.Set(ctx, "agent1", "key1", "value1")
	if err != nil {
		t.Fatalf("failed to set value: %v", err)
	}

	// Test Get
	val, err := store.Get(ctx, "agent1", "key1")
	if err != nil {
		t.Fatalf("failed to get value: %v", err)
	}
	if val != "value1" {
		t.Errorf("expected 'value1', got '%s'", val)
	}

	// Test Agent Isolation
	_, err = store.Get(ctx, "agent2", "key1")
	if err == nil {
		t.Errorf("expected error getting value for different agent")
	}

	// Test Overwrite
	err = store.Set(ctx, "agent1", "key1", "value2")
	if err != nil {
		t.Fatalf("failed to overwrite value: %v", err)
	}

	val, err = store.Get(ctx, "agent1", "key1")
	if err != nil {
		t.Fatalf("failed to get value: %v", err)
	}
	if val != "value2" {
		t.Errorf("expected 'value2', got '%s'", val)
	}

	// Test Agent Aware isolation blocks empty agent
	err = store.Set(ctx, "", "key1", "value3")
	if err == nil {
		t.Errorf("expected error setting value without agent ID in agent_aware mode")
	}

	_, err = store.Get(ctx, "", "key1")
	if err == nil {
		t.Errorf("expected error getting value without agent ID in agent_aware mode")
	}
}
