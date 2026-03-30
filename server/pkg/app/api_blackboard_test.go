// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/mcpany/core/server/pkg/middleware"
)

func TestBlackboardAPI(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "blackboard_test_api")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	dbPath := filepath.Join(tempDir, "blackboard.db")

	store, err := middleware.NewBlackboardStore(dbPath)
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}
	defer store.Close()

	// Seed data
	ctx := context.Background()
	_ = store.Set(ctx, "agent1", "key1", "value1")
	_ = store.Set(ctx, "agent2", "key2", "value2")

	app := NewApplication()
	app.standardMiddlewares = &middleware.StandardMiddlewares{
		Blackboard: store,
	}

	handler := app.handleBlackboardKeys()

	req := httptest.NewRequest(http.MethodGet, "/blackboard/keys", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	var entries []middleware.BlackboardEntry
	if err := json.Unmarshal(rec.Body.Bytes(), &entries); err != nil {
		t.Fatalf("failed to parse json response: %v", err)
	}

	if len(entries) != 2 {
		t.Errorf("expected 2 entries, got %d", len(entries))
	}

	foundAgent1 := false
	for _, entry := range entries {
		if entry.AgentID == "agent1" && entry.Key == "key1" && entry.Value == "value1" {
			foundAgent1 = true
			break
		}
	}

	if !foundAgent1 {
		t.Errorf("expected to find agent1 entry in response")
	}

	// Test method not allowed
	reqPost := httptest.NewRequest(http.MethodPost, "/blackboard/keys", nil)
	recPost := httptest.NewRecorder()

	handler.ServeHTTP(recPost, reqPost)

	if recPost.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected status %d, got %d", http.StatusMethodNotAllowed, recPost.Code)
	}
}
