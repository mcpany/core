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

	// Test Set without AgentAware
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

	// Test Agent Isolation (wrong agent, same scope)
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

	// Test with AgentAware and IntentScope
	ctxAware := context.WithValue(context.Background(), AgentAwareKey, true)
	ctxAware = context.WithValue(ctxAware, IntentScopeKey, "scope1")

	err = store.Set(ctxAware, "agent1", "key1", "value_scoped")
	if err != nil {
		t.Fatalf("failed to set scoped value: %v", err)
	}

	val, err = store.Get(ctxAware, "agent1", "key1")
	if err != nil {
		t.Fatalf("failed to get scoped value: %v", err)
	}
	if val != "value_scoped" {
		t.Errorf("expected 'value_scoped', got '%s'", val)
	}

	// Test Intent Scope Isolation
	ctxAwareWrongScope := context.WithValue(context.Background(), AgentAwareKey, true)
	ctxAwareWrongScope = context.WithValue(ctxAwareWrongScope, IntentScopeKey, "scope2")

	_, err = store.Get(ctxAwareWrongScope, "agent1", "key1")
	if err == nil {
		t.Errorf("expected error getting value for different intent scope")
	}

	// Test cross-contamination: non-scoped read should not see scoped writes
	// Non-scoped context for agent1
	ctxNonScoped := context.Background()
	val, err = store.Get(ctxNonScoped, "agent1", "key1")
	if err != nil {
		t.Fatalf("failed to get non-scoped value: %v", err)
	}
	// It should read the "value2" we set earlier in the test, not "value_scoped"
	if val != "value2" {
		t.Errorf("expected non-scoped value to be 'value2', got '%s'", val)
	}
}
