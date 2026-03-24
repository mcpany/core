// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestLazyMCPMiddleware(t *testing.T) {
	config := LazyMCPConfig{Enabled: true, Threshold: 0.85}
	middleware := NewLazyMCPMiddleware(config)

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
