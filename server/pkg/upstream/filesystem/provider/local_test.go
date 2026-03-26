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

func TestValidateLocalPath(t *testing.T) {
	// Create a temporary directory structure
	rootDir, err := os.MkdirTemp("", "fs_validate_test")
	require.NoError(t, err)
	defer os.RemoveAll(rootDir)

	// Create subdirectories and files
	dataDir := filepath.Join(rootDir, "data")
	err = os.Mkdir(dataDir, 0755)
	require.NoError(t, err)

	secretDir := filepath.Join(rootDir, "secret")
	err = os.Mkdir(secretDir, 0755)
	require.NoError(t, err)

	testFile := filepath.Join(dataDir, "test.txt")
	err = os.WriteFile(testFile, []byte("test"), 0644)
	require.NoError(t, err)

	secretFile := filepath.Join(secretDir, "secret.txt")
	err = os.WriteFile(secretFile, []byte("secret"), 0644)
	require.NoError(t, err)

	// Create symlink inside data pointing to secret
	symlinkPath := filepath.Join(dataDir, "link_to_secret")
	err = os.Symlink(secretFile, symlinkPath)
	require.NoError(t, err)

	// Create symlink inside data pointing to safe file
	safeLinkPath := filepath.Join(dataDir, "link_to_test")
	err = os.Symlink(testFile, safeLinkPath)
	require.NoError(t, err)

	tests := []struct {
		name        string
		virtualPath string
		rootPaths   map[string]string
		wantErr     bool
		expected    string // if empty, check if result matches real path of validation
	}{
		{
			name:        "Valid path",
			virtualPath: "/data/test.txt",
			rootPaths:   map[string]string{"/data": dataDir},
			wantErr:     false,
		},
		{
			name:        "Valid path with root /",
			virtualPath: "/data/test.txt",
			rootPaths:   map[string]string{"/": rootDir},
			wantErr:     false,
		},
		{
			name:        "Path traversal attempt",
			virtualPath: "/data/../secret/secret.txt",
			rootPaths:   map[string]string{"/data": dataDir},
			wantErr:     true, // Access denied or no matching root (since /secret is not under /data)
		},
		{
			name:        "Access denied via symlink",
			virtualPath: "/data/link_to_secret",
			rootPaths:   map[string]string{"/data": dataDir},
			wantErr:     true,
		},
		{
			name:        "Allowed symlink",
			virtualPath: "/data/link_to_test",
			rootPaths:   map[string]string{"/data": dataDir},
			wantErr:     false,
		},
		{
			name:        "No matching root",
			virtualPath: "/other/file.txt",
			rootPaths:   map[string]string{"/data": dataDir},
			wantErr:     true,
		},
		{
			name:        "Non-existent file valid parent",
			virtualPath: "/data/nonexistent.txt",
			rootPaths:   map[string]string{"/data": dataDir},
			wantErr:     false,
		},
		{
			name:        "Non-existent file traversal",
			virtualPath: "/data/../secret/nonexistent.txt",
			rootPaths:   map[string]string{"/data": dataDir},
			wantErr:     true,
		},
		{
			name:        "Root path fallback",
			virtualPath: "/test.txt",
			rootPaths:   map[string]string{"/": dataDir},
			wantErr:     false,
		},
		{
			name:        "Invalid prefix match",
			virtualPath: "/database/file.txt",
			rootPaths:   map[string]string{"/data": dataDir},
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewLocalProvider(nil, tt.rootPaths, nil, nil, 0)
			got, err := p.ResolvePath(tt.virtualPath)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				if tt.expected != "" {
					assert.Equal(t, tt.expected, got)
				} else {
					// Verify the path exists or is a valid path string
					assert.NotEmpty(t, got)
				}
			}
		})
	}
}

func TestLocalProviderAccessControl(t *testing.T) {
	// Create a temporary directory structure
	rootDir, err := os.MkdirTemp("", "fs_ac_test")
	require.NoError(t, err)
	defer os.RemoveAll(rootDir)

	// Create subdirectories
	publicDir := filepath.Join(rootDir, "public")
	err = os.Mkdir(publicDir, 0755)
	require.NoError(t, err)

	privateDir := filepath.Join(rootDir, "private")
	err = os.Mkdir(privateDir, 0755)
	require.NoError(t, err)

	// Create files
	publicFile := filepath.Join(publicDir, "info.txt")
	err = os.WriteFile(publicFile, []byte("public info"), 0644)
	require.NoError(t, err)

	privateFile := filepath.Join(privateDir, "secret.txt")
	err = os.WriteFile(privateFile, []byte("secret info"), 0644)
	require.NoError(t, err)

	// Root mapping: / -> rootDir
	// This exposes everything by default if no allow list.
	rootPaths := map[string]string{"/": rootDir}

	tests := []struct {
		name         string
		virtualPath  string
		allowedPaths []string
		deniedPaths  []string
		wantErr      bool
	}{
		{
			name:         "Default allow all",
			virtualPath:  "/public/info.txt",
			allowedPaths: nil,
			deniedPaths:  nil,
			wantErr:      false,
		},
		{
			name:         "Allow specific directory",
			virtualPath:  "/public/info.txt",
			allowedPaths: []string{publicDir},
			deniedPaths:  nil,
			wantErr:      false,
		},
		{
			name:         "Deny outside allowed directory",
			virtualPath:  "/private/secret.txt",
			allowedPaths: []string{publicDir},
			deniedPaths:  nil,
			wantErr:      true,
		},
		{
			name:         "Deny specific file",
			virtualPath:  "/private/secret.txt",
			allowedPaths: nil,
			deniedPaths:  []string{privateFile},
			wantErr:      true,
		},
		{
			name:         "Allow glob",
			virtualPath:  "/public/info.txt",
			allowedPaths: []string{filepath.Join(rootDir, "public", "*.txt")},
			deniedPaths:  nil,
			wantErr:      false,
		},
		{
			name:         "Deny glob",
			virtualPath:  "/private/secret.txt",
			allowedPaths: nil,
			deniedPaths:  []string{filepath.Join(rootDir, "private", "*.txt")},
			wantErr:      true,
		},
		{
			name:         "Deny overrides allow",
			virtualPath:  "/public/info.txt",
			allowedPaths: []string{publicDir},
			deniedPaths:  []string{publicFile},
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewLocalProvider(nil, rootPaths, tt.allowedPaths, tt.deniedPaths, 0)
			_, err := p.ResolvePath(tt.virtualPath)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
