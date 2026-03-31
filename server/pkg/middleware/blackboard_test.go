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

	store, err := NewBlackboardStore(dbPath)
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

	// Test ListAll
	err = store.Set(ctx, "agent2", "key2", "value_for_agent2")
	if err != nil {
		t.Fatalf("failed to set value: %v", err)
	}

	entries, err := store.ListAll(ctx)
	if err != nil {
		t.Fatalf("failed to list all entries: %v", err)
	}

	if len(entries) != 2 {
		t.Errorf("expected 2 entries, got %d", len(entries))
	}

	foundAgent1 := false
	foundAgent2 := false
	for _, entry := range entries {
		if entry.AgentID == "agent1" && entry.Key == "key1" && entry.Value == "value2" {
			foundAgent1 = true
		}
		if entry.AgentID == "agent2" && entry.Key == "key2" && entry.Value == "value_for_agent2" {
			foundAgent2 = true
		}
	}

	if !foundAgent1 {
		t.Errorf("expected to find agent1's entry in ListAll")
	}
	if !foundAgent2 {
		t.Errorf("expected to find agent2's entry in ListAll")
	}
}
