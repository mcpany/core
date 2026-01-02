// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package filesystem

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateLocalPath(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "mcp-filesystem-test")
	assert.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Create some files and directories
	err = os.Mkdir(filepath.Join(tmpDir, "allowed"), 0755)
	assert.NoError(t, err)
	err = os.WriteFile(filepath.Join(tmpDir, "allowed", "file.txt"), []byte("content"), 0644)
	assert.NoError(t, err)
	err = os.Mkdir(filepath.Join(tmpDir, "forbidden"), 0755)
	assert.NoError(t, err)

	rootPaths := map[string]string{
		"/workspace": filepath.Join(tmpDir, "allowed"),
	}

	u := &Upstream{}

	tests := []struct {
		name        string
		virtualPath string
		expectErr   bool
		expectedPathSuffix string
	}{
		{
			name:        "Valid path",
			virtualPath: "/workspace/file.txt",
			expectErr:   false,
			expectedPathSuffix: "allowed/file.txt",
		},
		// This test case is broken because without leading slash, it tries to match keys.
		// If keys in rootPaths are like "/workspace", then "workspace/file.txt" won't match "/workspace".
		// It will fallback to finding root "/". If not present -> Error.
		// However, validateLocalPath does NOT prepend slash to virtualPath automatically?
		// checkPath := virtualPath
		// if !strings.HasPrefix(checkPath, "/") { checkPath = "/" + checkPath }
		// So "workspace/file.txt" becomes "/workspace/file.txt".
		// And should match.
		// BUT the error in previous run was:
		// "/tmp/mcp-filesystem-test.../allowed/workspace/file.txt" does not contain "allowed/file.txt"
		// Wait, if it matched, relative path calculation:
		// relativePath = strings.TrimPrefix(virtualPath, bestMatchVirtual)
		// virtualPath is "workspace/file.txt". bestMatchVirtual is "/workspace".
		// Prefix doesn't match string-wise if one has slash and other doesn't?
		// `strings.HasPrefix("workspace/file.txt", "/workspace")` is False.
		// The logic in local.go says:
		// checkPath := virtualPath... (adds slash)
		// if strings.HasPrefix(checkPath, cleanVRoot) { ... match found ... }
		// THEN:
		// relativePath := strings.TrimPrefix(virtualPath, bestMatchVirtual)
		// If virtualPath was "workspace/file.txt" and bestMatchVirtual is "/workspace", TrimPrefix does nothing!
		// So relativePath remains "workspace/file.txt".
		// Then targetPath = Join(realRoot, "workspace/file.txt") -> .../allowed/workspace/file.txt
		// Which is WRONG. It should be .../allowed/file.txt.
		// FIX: We should trim prefix from checkPath, or ensure we use the matched part correctly.
		{
			name:        "Valid path without leading slash",
			virtualPath: "workspace/file.txt",
			expectErr:   false,
			expectedPathSuffix: "allowed/file.txt",
		},
		{
			name:        "Path traversal attempt 1",
			virtualPath: "/workspace/../forbidden/file.txt",
			expectErr:   true,
		},
		{
			name:        "Path traversal attempt 2",
			virtualPath: "/workspace/../../etc/passwd",
			expectErr:   true,
		},
		{
			name:        "Unknown root",
			virtualPath: "/other/file.txt",
			expectErr:   true,
		},
		{
			name:        "Root directory itself",
			virtualPath: "/workspace",
			expectErr:   false,
			expectedPathSuffix: "allowed",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			path, err := u.validateLocalPath(tc.virtualPath, rootPaths)
			if tc.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.True(t, filepath.IsAbs(path))
				// Check if path ends with expected suffix
				// We need to normalize separators
				expected := filepath.FromSlash(tc.expectedPathSuffix)
				assert.Contains(t, path, expected)
			}
		})
	}
}
