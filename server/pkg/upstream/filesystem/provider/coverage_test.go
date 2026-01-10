// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package provider

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLocalProvider_ResolvePath_EdgeCases(t *testing.T) {
	// Setup: create a root dir
	tmpDir := t.TempDir()
	p, err := NewLocalProvider(nil, map[string]string{"/data": tmpDir})
	require.NoError(t, err)

	// Case: no root paths defined
	pEmpty, _ := NewLocalProvider(nil, nil)
	_, err = pEmpty.ResolvePath("/foo")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no root paths defined")

	// Case: virtual path with no slash prefix
	// The code handles this by adding /, but we should verifying it works
	p2, err := NewLocalProvider(nil, map[string]string{"data": tmpDir})
	require.NoError(t, err)
	path, err := p2.ResolvePath("data/file.txt")
	assert.NoError(t, err)
	assert.Contains(t, path, tmpDir)

	// Case: root path does not exist
	// filepath.Abs should succeed, but EvalSymlinks should fail
	nonExistentRoot := filepath.Join(tmpDir, "does-not-exist")
	p3, err := NewLocalProvider(nil, map[string]string{"/": nonExistentRoot})
	// Now NewLocalProvider returns error for non-existent root
	assert.Error(t, err)
	assert.Nil(t, p3)
	assert.Contains(t, err.Error(), "root path does not exist")

	// Case: Non-existent file deep in path
	// This exercises the "deepest existing ancestor" logic
	// Create a dir
	subDir := filepath.Join(tmpDir, "sub")
	require.NoError(t, os.Mkdir(subDir, 0755))

	p4, err := NewLocalProvider(nil, map[string]string{"/": tmpDir})
	require.NoError(t, err)
	resolved, err := p4.ResolvePath("/sub/non/existent/file.txt")
	assert.NoError(t, err)
	expected := filepath.Join(subDir, "non/existent/file.txt")
	// On some systems EvalSymlinks might resolve /private/tmp to /tmp, so we compare Canonical paths
	expectedCanonical, _ := filepath.EvalSymlinks(subDir)
	if expectedCanonical == "" {
		expectedCanonical = subDir
	}
	expected = filepath.Join(expectedCanonical, "non/existent/file.txt")

	assert.Equal(t, expected, resolved)

	// Case: Access denied path traversal detected (target not in root)
	// We need a symlink that points outside the root, but the symlink itself is inside
	// AND the target exists.
	outsideDir := t.TempDir()
	outsideFile := filepath.Join(outsideDir, "outside.txt")
	os.WriteFile(outsideFile, []byte("outside"), 0644)

	symlink := filepath.Join(tmpDir, "badlink")
	os.Symlink(outsideFile, symlink)

	p5, err := NewLocalProvider(nil, map[string]string{"/": tmpDir})
	require.NoError(t, err)
	_, err = p5.ResolvePath("/badlink")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "access denied: path traversal detected")

	// Check p is not unused
	require.NotNil(t, p)
}

func TestLocalProvider_ResolvePath_PermissionError(t *testing.T) {
	if os.Getuid() == 0 {
		t.Skip("Skipping permission denied test as root")
	}

	tmpDir := t.TempDir()
	// Create a directory with no permissions
	noPermDir := filepath.Join(tmpDir, "noperm")
	err := os.Mkdir(noPermDir, 0000)
	require.NoError(t, err)
	// Make sure we can remove it later
	defer os.Chmod(noPermDir, 0755)

	// Test EvalSymlinks failure on target path
	p, err := NewLocalProvider(nil, map[string]string{"/": tmpDir})
	require.NoError(t, err)
	_, err = p.ResolvePath(filepath.Join("noperm", "file.txt"))
	// Depending on OS, this might fail with "permission denied" or work if parent has permissions.
	// Actually, accessing "noperm/file.txt" where noperm is 0000 should fail stat.
	// If it fails with something other than NotExist, we hit line 142.
	// However, filepath.EvalSymlinks calls lstat, which requires exec permission on parent.
	// If parent has no exec, lstat fails with permission denied.

	assert.Error(t, err, "Expected permission denied or similar error")
}

func TestZipProvider_ManualClose(t *testing.T) {
	p := &ZipProvider{fs: nil, closer: nil}
	assert.NoError(t, p.Close())
}
