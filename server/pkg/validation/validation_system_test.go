// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package validation

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIsAllowedSystemPath(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir := t.TempDir()

	// Create allowed directory
	allowedDir := filepath.Join(tmpDir, "allowed")
	require.NoError(t, os.Mkdir(allowedDir, 0755))

	// Create forbidden directory
	forbiddenDir := filepath.Join(tmpDir, "forbidden")
	require.NoError(t, os.Mkdir(forbiddenDir, 0755))

	// Set allowed paths
	defer SetAllowedPaths(nil)
	SetAllowedPaths([]string{allowedDir})

	tests := []struct {
		name        string
		path        string
		setup       func(path string) // Optional setup (create file)
		expectError bool
		errorMsg    string
	}{
		{
			name:        "Allowed Path - DB File",
			path:        filepath.Join(allowedDir, "audit.db"),
			expectError: false,
		},
		{
			name:        "Allowed Path - SQLite File",
			path:        filepath.Join(allowedDir, "data.sqlite"),
			expectError: false,
		},
		{
			name:        "Allowed Path - Log File",
			path:        filepath.Join(allowedDir, "server.log"),
			expectError: false,
		},
		{
			name:        "Blocked - .env file in allowed dir",
			path:        filepath.Join(allowedDir, ".env"),
			expectError: true,
			errorMsg:    "access to sensitive file",
		},
		{
			name:        "Blocked - config.yaml in allowed dir",
			path:        filepath.Join(allowedDir, "config.yaml"),
			expectError: true,
			errorMsg:    "access to sensitive file",
		},
		{
			name:        "Blocked - Private Key (.pem) in allowed dir",
			path:        filepath.Join(allowedDir, "key.pem"),
			expectError: true,
			errorMsg:    "access to sensitive file",
		},
		{
			name:        "Blocked - Forbidden Directory",
			path:        filepath.Join(forbiddenDir, "audit.db"),
			expectError: true,
			errorMsg:    "not allowed",
		},
		{
			name:        "Blocked - Traversal to Forbidden",
			path:        filepath.Join(allowedDir, "../forbidden/audit.db"),
			expectError: true,
			errorMsg:    "not allowed", // filepath.Join resolves '..', so it fails allowed dir check
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				tt.setup(tt.path)
			}
			err := IsAllowedSystemPath(tt.path)
			if tt.expectError {
				require.Error(t, err)
				if tt.errorMsg != "" {
					require.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestIsAllowedSystemPath_Symlinks(t *testing.T) {
	tmpDir := t.TempDir()
	allowedDir := filepath.Join(tmpDir, "allowed")
	forbiddenDir := filepath.Join(tmpDir, "forbidden")
	require.NoError(t, os.Mkdir(allowedDir, 0755))
	require.NoError(t, os.Mkdir(forbiddenDir, 0755))

	defer SetAllowedPaths(nil)
	SetAllowedPaths([]string{allowedDir})

	// 1. Symlink inside allowed pointing to forbidden file
	forbiddenFile := filepath.Join(forbiddenDir, "audit.db")
	require.NoError(t, os.WriteFile(forbiddenFile, []byte("data"), 0644))

	linkPath := filepath.Join(allowedDir, "link_to_db")
	require.NoError(t, os.Symlink(forbiddenFile, linkPath))

	err := IsAllowedSystemPath(linkPath)
	require.Error(t, err, "Should block symlink to forbidden dir")
	require.Contains(t, err.Error(), "not allowed")

	// 2. Symlink inside allowed pointing to forbidden sensitive file
	forbiddenEnv := filepath.Join(forbiddenDir, ".env")
	require.NoError(t, os.WriteFile(forbiddenEnv, []byte("SECRET=1"), 0644))

	linkEnv := filepath.Join(allowedDir, "link_to_env")
	require.NoError(t, os.Symlink(forbiddenEnv, linkEnv))

	err = IsAllowedSystemPath(linkEnv)
	require.Error(t, err, "Should block symlink to sensitive file")
	require.Contains(t, err.Error(), "access to sensitive file")
}
