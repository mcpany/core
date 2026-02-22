// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package provider

import (
	"os"
	"path/filepath"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLocalProvider_SensitiveFiles(t *testing.T) {
	// Create a temp directory
	tmpDir := t.TempDir()

	// Create .git directory and config file
	gitDir := filepath.Join(tmpDir, ".git")
	err := os.Mkdir(gitDir, 0755)
	require.NoError(t, err)
	configFile := filepath.Join(gitDir, "config")
	err = os.WriteFile(configFile, []byte("core"), 0644)
	require.NoError(t, err)

	// Create a normal file
	normalFile := filepath.Join(tmpDir, "README.md")
	err = os.WriteFile(normalFile, []byte("hello"), 0644)
	require.NoError(t, err)

	// Setup LocalProvider with root at tmpDir
	roots := map[string]string{
		"/": tmpDir,
	}
	p := NewLocalProvider(nil, roots, nil, nil, configv1.FilesystemUpstreamService_ALLOW)

	// Test access to normal file
	// Since root is "/", virtual path "/README.md" maps to tmpDir/README.md
	// The provider findBestMatch returns "/", tmpDir.
	// ResolveSymlinks: virtualPath "/README.md". relativePath "README.md". targetPath tmpDir/README.md.
	resolved, err := p.ResolvePath("/README.md")
	// On Mac/Linux, resolved path might involve symlinks (/var/folders -> /private/var/folders)
	// We rely on EvalSymlinks in ResolvePath.
	// We should compare Canonical paths.
	realNormal, _ := filepath.EvalSymlinks(normalFile)

	assert.NoError(t, err)
	assert.Equal(t, realNormal, resolved)

	// Test access to sensitive file (.git/config)
	// This should fail now
	_, err = p.ResolvePath("/.git/config")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "access to sensitive directory \".git\" is denied")

    // Test access to .git directory itself
    _, err = p.ResolvePath("/.git")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "access to sensitive directory \".git\" is denied")
}
