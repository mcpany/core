// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package provider

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResolvePath_EdgeCases(t *testing.T) {
	// Create a temporary directory structure
	rootDir, err := os.MkdirTemp("", "fs_edge_test")
	require.NoError(t, err)
	defer os.RemoveAll(rootDir)

	// Create data directory
	dataDir := filepath.Join(rootDir, "data")
	err = os.Mkdir(dataDir, 0755)
	require.NoError(t, err)

	// Create secret directory
	secretDir := filepath.Join(rootDir, "secret")
	err = os.Mkdir(secretDir, 0755)
	require.NoError(t, err)

	// Create a broken symlink in dataDir pointing to non-existent file
	brokenLink := filepath.Join(dataDir, "broken_link")
	err = os.Symlink(filepath.Join(rootDir, "nonexistent"), brokenLink)
	require.NoError(t, err)

	// Create a symlink in dataDir pointing to secretDir (valid target, but denied by policy potentially)
	secretLink := filepath.Join(dataDir, "secret_link")
	err = os.Symlink(secretDir, secretLink)
	require.NoError(t, err)

	// Create a "broken" symlink that points to an existing file but we will try to traverse it
	// Actually, if it points to a file, we can't traverse it as a directory.
	// Let's create a symlink to a directory that exists, then remove the directory?
	// No, that's just a broken symlink.

	rootPaths := map[string]string{"/data": dataDir}

	tests := []struct {
		name        string
		virtualPath string
		wantErr     bool
		errContains string
	}{
		{
			name:        "Broken symlink access",
			virtualPath: "/data/broken_link",
			wantErr:     true,
			// Depending on implementation, might be "file does not exist" or "broken symlink"
			// The code checks for broken symlinks in resolveNonExistentPath only if we are traversing?
			// If we ask for the link itself, it might resolve it and fail to Stat the target?
			// Let's see. logic: filepath.EvalSymlinks fails for broken symlink.
			// Then it calls resolveNonExistentPath.
			// resolveNonExistentPath checks if components are broken symlinks.
		},
		{
			name:        "Traversal through broken symlink",
			virtualPath: "/data/broken_link/file.txt",
			wantErr:     true,
			errContains: "broken symlink",
		},
		{
			name:        "Non-existent path with valid parent",
			virtualPath: "/data/missing.txt",
			wantErr:     false,
		},
		{
			name:        "Non-existent path deep traversal",
			virtualPath: "/data/subdir/missing.txt",
			wantErr:     false,
		},
		{
			name:        "Traversal through symlink to denied area",
			virtualPath: "/data/secret_link/file.txt",
			wantErr:     true, // access denied: path traversal detected (because it resolves to secretDir which is not under dataDir)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewLocalProvider(nil, rootPaths, nil, nil, 0)
			_, err := p.ResolvePath(tt.virtualPath)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
