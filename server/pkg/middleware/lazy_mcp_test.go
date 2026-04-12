// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestLazyMCPMiddleware(t *testing.T) {
	config := LazyMCPConfig{Enabled: true, Threshold: 0.85, CacheTTL: 600}
	middleware := NewLazyMCPMiddleware(config)

	if middleware.config.CacheTTL != 600 {
		t.Errorf("expected CacheTTL to be 600, got %d", middleware.config.CacheTTL)
	}

	originalResult := &mcp.ListToolsResult{
		Tools: []*mcp.Tool{
			{Name: "fs:read", Description: "Read a file from disk."},
			{Name: "fs:write", Description: "Write a file to disk."},
			{Name: "db:query", Description: "Execute a SQL query against the database."},
			{Name: "aws:s3:upload", Description: "Upload a file to S3 bucket."},
		},
	}

	tests := []struct {
		name          string
		intent        string
		expectedCount int
	}{
		{
			name:          "filter for filesystem tools",
			intent:        "disk",
			expectedCount: 2,
		},
		{
			name:          "filter for db tools",
			intent:        "SQL",
			expectedCount: 1,
		},
		{
			name:          "filter for matching tool name",
			intent:        "fs:",
			expectedCount: 2,
		},
		{
			name:          "no match returns empty",
			intent:        "kubernetes",
			expectedCount: 0,
		},
		{
			name:          "empty intent returns original list",
			intent:        "",
			expectedCount: 4,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filtered := middleware.FilterTools(originalResult, tt.intent)

			if len(filtered.Tools) != tt.expectedCount {
				t.Errorf("expected %d tools, got %d", tt.expectedCount, len(filtered.Tools))
			}
		})
	}
}

func TestLazyMCPMiddlewareCache(t *testing.T) {
	config := LazyMCPConfig{Enabled: true, Threshold: 0.85, CacheTTL: 1} // 1 second TTL
	middleware := NewLazyMCPMiddleware(config)

	originalResult := &mcp.ListToolsResult{
		Tools: []*mcp.Tool{
			{Name: "fs:read", Description: "Read a file from disk."},
		},
	}

	// First call should populate cache
	filtered1 := middleware.FilterTools(originalResult, "disk")
	if len(filtered1.Tools) != 1 {
		t.Fatalf("expected 1 tool on first call, got %d", len(filtered1.Tools))
	}

	// Verify cache is populated
	middleware.mu.RLock()
	entry, found := middleware.cache["disk"]
	middleware.mu.RUnlock()
	if !found {
		t.Fatalf("expected cache to be populated")
	}

	// Second call should return from cache (we can verify this by checking if the returned object is the same pointer)
	filtered2 := middleware.FilterTools(originalResult, "disk")
	if filtered1 != filtered2 {
		t.Fatalf("expected second call to return the exact same pointer from cache")
	}

	// Check expiry
	time.Sleep(1500 * time.Millisecond)

	// Change originalResult to see if we get a new result
	originalResult2 := &mcp.ListToolsResult{
		Tools: []*mcp.Tool{
			{Name: "fs:read", Description: "Read a file from disk."},
			{Name: "fs:write", Description: "Write a file to disk."},
		},
	}

	filtered3 := middleware.FilterTools(originalResult2, "disk")
	if len(filtered3.Tools) != 2 {
		t.Fatalf("expected 2 tools on third call after cache expiry, got %d", len(filtered3.Tools))
	}
	if filtered3 == filtered1 {
		t.Fatalf("expected third call to return a new pointer after cache expiry")
	}

	// Verify cache is updated
	middleware.mu.RLock()
	entry3, found3 := middleware.cache["disk"]
	middleware.mu.RUnlock()
	if !found3 {
		t.Fatalf("expected cache to be updated")
	}
	if entry3.result != filtered3 {
		t.Fatalf("expected cache to hold the new result")
	}
}
