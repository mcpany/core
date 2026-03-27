// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"context"
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/mcpany/core/server/pkg/tool"
)

func TestCFIAMiddleware(t *testing.T) {
	// Setup test files and known hashes
	tempDir := t.TempDir()
	goodFile := filepath.Join(tempDir, "GEMINI.md")
	badFile := filepath.Join(tempDir, "settings.json")

	goodContent := []byte("Authorized project context")
	badContent := []byte("Unauthorized injected context")

	if err := os.WriteFile(goodFile, goodContent, 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	if err := os.WriteFile(badFile, badContent, 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	goodHash := fmt.Sprintf("%x", sha256.Sum256(goodContent))
	expectedBadHash := "fakehash123" // different from actual hash

	config := CFIAConfig{
		Enabled:     true,
		TargetTools: []string{"read_file"},
		AttestedHashes: map[string]string{
			goodFile: goodHash,
			badFile:  expectedBadHash,
		},
		ArgumentName: "filepath",
	}

	middleware := NewCFIAMiddleware(config)

	ctx := context.Background()
	mockNext := func(ctx context.Context, req *tool.ExecutionRequest) (any, error) {
		return "success", nil
	}

	tests := []struct {
		name      string
		req       *tool.ExecutionRequest
		expectErr bool
	}{
		{
			name: "Authorized Context - Match",
			req: &tool.ExecutionRequest{
				ToolName: "read_file",
				Arguments: map[string]interface{}{
					"filepath": goodFile,
				},
			},
			expectErr: false,
		},
		{
			name: "Unauthorized Context - Hash Mismatch",
			req: &tool.ExecutionRequest{
				ToolName: "read_file",
				Arguments: map[string]interface{}{
					"filepath": badFile,
				},
			},
			expectErr: true,
		},
		{
			name: "Ignored Tool - Pass Through",
			req: &tool.ExecutionRequest{
				ToolName: "other_tool",
				Arguments: map[string]interface{}{
					"filepath": badFile,
				},
			},
			expectErr: false,
		},
		{
			name: "Missing Filepath Argument - Pass Through",
			req: &tool.ExecutionRequest{
				ToolName: "read_file",
				Arguments: map[string]interface{}{
					"otherarg": "value",
				},
			},
			expectErr: false,
		},
		{
			name: "File Does Not Exist",
			req: &tool.ExecutionRequest{
				ToolName: "read_file",
				Arguments: map[string]interface{}{
					"filepath": filepath.Join(tempDir, "does-not-exist.md"), // not in attested map
				},
			},
			expectErr: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := middleware.Execute(ctx, tc.req, mockNext)
			if tc.expectErr && err == nil {
				t.Errorf("Expected an error for %s, got nil", tc.name)
			}
			if !tc.expectErr && err != nil {
				t.Errorf("Did not expect an error for %s, got: %v", tc.name, err)
			}
		})
	}
}

func TestCFIAMiddlewareDisabled(t *testing.T) {
	tempDir := t.TempDir()
	badFile := filepath.Join(tempDir, "settings.json")
	if err := os.WriteFile(badFile, []byte("content"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	config := CFIAConfig{
		Enabled:     false,
		TargetTools: []string{"read_file"},
		AttestedHashes: map[string]string{
			badFile: "fakehash",
		},
		ArgumentName: "filepath",
	}

	middleware := NewCFIAMiddleware(config)
	ctx := context.Background()
	req := &tool.ExecutionRequest{
		ToolName: "read_file",
		Arguments: map[string]interface{}{
			"filepath": badFile,
		},
	}
	mockNext := func(ctx context.Context, req *tool.ExecutionRequest) (any, error) {
		return "success", nil
	}

	_, err := middleware.Execute(ctx, req, mockNext)
	if err != nil {
		t.Errorf("Expected nil error when disabled, got: %v", err)
	}
}
