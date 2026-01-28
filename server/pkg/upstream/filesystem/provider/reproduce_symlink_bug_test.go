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

func TestSymlinkedAllowedPath(t *testing.T) {
	// Create a temporary directory for the test
	tmpDir := t.TempDir()

	// Evaluate symlinks for tmpDir to ensure we have the canonical path as base
	canonicalTmpDir, err := filepath.EvalSymlinks(tmpDir)
	require.NoError(t, err)

	// Create a real directory
	realDir := filepath.Join(canonicalTmpDir, "real_dir")
	err = os.Mkdir(realDir, 0755)
	require.NoError(t, err)

	// Create a file in the real directory
	testFile := filepath.Join(realDir, "test.txt")
	err = os.WriteFile(testFile, []byte("test"), 0644)
	require.NoError(t, err)

	// Create a symlink to the real directory
	linkDir := filepath.Join(canonicalTmpDir, "link_dir")
	err = os.Symlink(realDir, linkDir)
	require.NoError(t, err)

	// Configure the provider with the symlink as the allowed path
	rootPaths := map[string]string{
		"/": canonicalTmpDir,
	}

	// The user explicitly allows "link_dir" (the symlink)
	allowedPaths := []string{linkDir}
	deniedPaths := []string{}

	// Create provider
	p := NewLocalProvider(nil, rootPaths, allowedPaths, deniedPaths, configv1.FilesystemUpstreamService_SYMLINK_MODE_UNSPECIFIED)

	// Access via the symlink
	virtualPath := "/link_dir/test.txt"

	resolved, err := p.ResolvePath(virtualPath)

	// This should succeed if we handle symlinks in allowedPaths correctly.
	assert.NoError(t, err, "Should resolve path even if allowed path is a symlink")
	assert.Equal(t, testFile, resolved)
}
