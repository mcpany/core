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

func TestResolvePath_BrokenSymlink(t *testing.T) {
	rootDir, err := os.MkdirTemp("", "fs_broken_symlink_test")
	require.NoError(t, err)
	defer os.RemoveAll(rootDir)

	// Create a broken symlink: link -> non_existent
	linkPath := filepath.Join(rootDir, "broken_link")
	err = os.Symlink("non_existent_target", linkPath)
	require.NoError(t, err)

	// Try to resolve path through broken symlink
	// /broken_link/some/file.txt
	rootPaths := map[string]string{"/": rootDir}
	p := NewLocalProvider(nil, rootPaths, nil, nil, 0)

	_, err = p.ResolvePath("/broken_link/some/file.txt")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "broken symlink")
}

func TestResolvePath_PermissionDenied(t *testing.T) {
	if os.Getuid() == 0 {
		t.Skip("Skipping permission denied test as root user")
	}

	rootDir, err := os.MkdirTemp("", "fs_perm_test")
	require.NoError(t, err)
	defer os.RemoveAll(rootDir)

	// Create a directory with no permissions
	lockedDir := filepath.Join(rootDir, "locked")
	err = os.Mkdir(lockedDir, 0755)
	require.NoError(t, err)

	// Create a file inside (before locking)
	err = os.WriteFile(filepath.Join(lockedDir, "secret.txt"), []byte("secret"), 0644)
	require.NoError(t, err)

	// Lock the directory
	err = os.Chmod(lockedDir, 0000)
	require.NoError(t, err)
	// Restore permissions on cleanup so RemoveAll works
	defer os.Chmod(lockedDir, 0755)

	// Try to resolve a file inside the locked directory
	// Note: resolveNonExistentPath logic relies on Stat() failing.
	// If we try to resolve "/locked/secret.txt", and "locked" is 0000:
	// Stat("/locked/secret.txt") fails (permission denied).
	// Ancestor search finds rootDir as existing path.
	// remainingPath is "locked/secret.txt".
	// nextComponent is "locked".
	// Lstat("/locked") succeeds.
	// It is not a symlink.
	// So it continues.
	// Then it joins existingPathCanonical (rootDir) with remainingPath.
	// returns rootDir/locked/secret.txt.
	// Then it continues to checkPathSecurity.

	// Wait, if Stat failed due to permission denied, maybe we want it to fail?
	// But `ResolvePath` is about path resolution, actual access (Read) comes later.
	// However, `ResolvePath` should probably fail if it can't verify safety?
	// The current logic seems to allow it if it's not a broken symlink.
	// BUT, if we can't traverse it, we can't verify if it resolves to a symlink pointing outside!
	// So "locked" is a directory. If we treat it as just a name, we assume it's safe.
	// But what if "locked" was a symlink to outside, but we couldn't read it?
	// Lstat would tell us if it is a symlink.
	// Since Lstat succeeds and says it's a dir, we know it is a dir (or at least not a symlink).
	// So it is safe to proceed in terms of symlink traversal.
	// The actual open() later will fail.

	// Let's test a case where we try to resolve a NON-EXISTENT file under a locked directory.
	// /locked/nonexistent.txt

	rootPaths := map[string]string{"/": rootDir}
	p := NewLocalProvider(nil, rootPaths, nil, nil, 0)

	_, err = p.ResolvePath("/locked/nonexistent.txt")
	// We expect an error because checking the path involves Lstat/EvalSymlinks which hits Permission Denied
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "permission denied")
}

func TestResolvePath_BestMatchEdgeCases(t *testing.T) {
	rootDir, err := os.MkdirTemp("", "fs_match_test")
	require.NoError(t, err)
	defer os.RemoveAll(rootDir)

	dirA := filepath.Join(rootDir, "a")
	dirAB := filepath.Join(rootDir, "ab")

	err = os.Mkdir(dirA, 0755)
	require.NoError(t, err)
	err = os.Mkdir(dirAB, 0755)
	require.NoError(t, err)

	rootPaths := map[string]string{
		"/a":  dirA,
		"/ab": dirAB,
	}
	p := NewLocalProvider(nil, rootPaths, nil, nil, 0)

	// Test exact match
	p1, err := p.ResolvePath("/a/file.txt")
	assert.NoError(t, err)
	assert.Equal(t, filepath.Join(dirA, "file.txt"), p1)

	// Test prefix overlap (/ab should not match /a)
	p2, err := p.ResolvePath("/ab/file.txt")
	assert.NoError(t, err)
	assert.Equal(t, filepath.Join(dirAB, "file.txt"), p2)

	// Test trailing slash in root definition handling (logic should handle it)
	rootPathsWithSlash := map[string]string{
		"/a/": dirA,
	}
	pSlash := NewLocalProvider(nil, rootPathsWithSlash, nil, nil, 0)
	p3, err := pSlash.ResolvePath("/a/file.txt")
	assert.NoError(t, err)
	assert.Equal(t, filepath.Join(dirA, "file.txt"), p3)
}

func TestResolvePath_NoRoots(t *testing.T) {
	p := NewLocalProvider(nil, nil, nil, nil, 0)
	_, err := p.ResolvePath("/anything")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no root paths defined")
}

func TestResolvePath_NoMatch(t *testing.T) {
	rootPaths := map[string]string{"/data": "/tmp/data"}
	p := NewLocalProvider(nil, rootPaths, nil, nil, 0)

	_, err := p.ResolvePath("/other/file")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not allowed")
}

func TestResolvePath_SymlinkLoop(t *testing.T) {
	// EvalSymlinks usually catches this and returns error, but let's verify.
	rootDir, err := os.MkdirTemp("", "fs_loop_test")
	require.NoError(t, err)
	defer os.RemoveAll(rootDir)

	link1 := filepath.Join(rootDir, "link1")
	link2 := filepath.Join(rootDir, "link2")

	err = os.Symlink(link2, link1)
	require.NoError(t, err)
	err = os.Symlink(link1, link2)
	require.NoError(t, err)

	rootPaths := map[string]string{"/": rootDir}
	p := NewLocalProvider(nil, rootPaths, nil, nil, 0)

	_, err = p.ResolvePath("/link1")
	assert.Error(t, err)
	// Error message comes from filepath.EvalSymlinks, typically "too many levels of symbolic links"
}

func TestResolvePath_ComplexGlobs(t *testing.T) {
	rootDir, err := os.MkdirTemp("", "fs_glob_test")
	require.NoError(t, err)
	defer os.RemoveAll(rootDir)

	// Structure:
	// /allowed/ok.txt
	// /allowed/sub/ok.txt
	// /allowed/sub/bad.log

	err = os.MkdirAll(filepath.Join(rootDir, "allowed", "sub"), 0755)
	require.NoError(t, err)

	rootPaths := map[string]string{"/": rootDir}
	allowed := []string{
		filepath.Join(rootDir, "allowed", "*.txt"),      // shallow
		filepath.Join(rootDir, "allowed", "**", "*.txt"), // deep (if supported by filepath.Match? No, filepath.Match doesn't support **)
		// Go's filepath.Match does NOT support **. It only supports *.
		// But let's check what patterns we can use.
		// "allowed/*/*.txt" matches allowed/sub/ok.txt
	}
	// Let's use simpler glob patterns that Go supports.
	allowed = []string{
		filepath.Join(rootDir, "allowed", "*.txt"),
		filepath.Join(rootDir, "allowed", "*", "*.txt"),
	}

	p := NewLocalProvider(nil, rootPaths, allowed, nil, 0)

	// allowed/ok.txt -> Matches first pattern
	_, err = p.ResolvePath("/allowed/ok.txt")
	assert.NoError(t, err)

	// allowed/sub/ok.txt -> Matches second pattern
	_, err = p.ResolvePath("/allowed/sub/ok.txt")
	assert.NoError(t, err)

	// allowed/sub/bad.log -> No match
	_, err = p.ResolvePath("/allowed/sub/bad.log")
	assert.Error(t, err)
}
