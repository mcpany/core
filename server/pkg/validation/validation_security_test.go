// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package validation

import (
	"os"
	"path/filepath"
	"testing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIsAllowedPath_SymlinkTraversal_Block(t *testing.T) {
	// Setup: Create a temp directory for the test
	tempDir := t.TempDir()
	cwd, _ := os.Getwd()
	defer os.Chdir(cwd)
	require.NoError(t, os.Chdir(tempDir))

	// Create a secret file outside the "safe" directory
	// In this test, "safe" is the tempDir (CWD)
	// We need another directory outside of CWD
	outsideDir, err := os.MkdirTemp("", "outside")
	require.NoError(t, err)
	defer os.RemoveAll(outsideDir)

	secretFile := filepath.Join(outsideDir, "secret.txt")
	require.NoError(t, os.WriteFile(secretFile, []byte("super secret"), 0644))

	// Create a symlink in the current directory pointing to the secret file
	symlinkName := "harmless_link"
	err = os.Symlink(secretFile, symlinkName)
	require.NoError(t, err)

	// Verify IsAllowedPath blocks it
	err = IsAllowedPath(symlinkName)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "is not allowed")
}

func TestIsAllowedPath_NonExistentFile_Safe(t *testing.T) {
	// Setup
	tempDir := t.TempDir()
	cwd, _ := os.Getwd()
	defer os.Chdir(cwd)
	require.NoError(t, os.Chdir(tempDir))

	// Test a non-existent file in CWD
	err := IsAllowedPath("non_existent.txt")
	assert.NoError(t, err)

	// Test a non-existent file in a non-existent subdir
	err = IsAllowedPath("subdir/non_existent.txt")
	assert.NoError(t, err)
}

func TestIsAllowedPath_NonExistentFile_UnsafeSymlinkParent(t *testing.T) {
	// Setup
	tempDir := t.TempDir()
	cwd, _ := os.Getwd()
	defer os.Chdir(cwd)
	require.NoError(t, os.Chdir(tempDir))

	// Create a symlink to outside
	outsideDir, err := os.MkdirTemp("", "outside")
	require.NoError(t, err)
	defer os.RemoveAll(outsideDir)

	err = os.Symlink(outsideDir, "link_to_outside")
	require.NoError(t, err)

	// Try to access a non-existent file through the symlink
	// IsAllowedPath("link_to_outside/missing.txt")
	// "link_to_outside" exists and is a symlink to outside.
	// The logic should walk up, find "link_to_outside" exists, resolve it -> /tmp/outside
	// Then append /missing.txt -> /tmp/outside/missing.txt
	// This is outside CWD, so it should fail.

	err = IsAllowedPath("link_to_outside/missing.txt")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "is not allowed")
}

func TestIsAllowedPath_EvalSymlinksError_Permission(t *testing.T) {
	// Setup
	tempDir := t.TempDir()
	cwd, _ := os.Getwd()
	defer os.Chdir(cwd)
	require.NoError(t, os.Chdir(tempDir))

	// Create a directory with no permissions
	noPermDir := "noperm"
	require.NoError(t, os.Mkdir(noPermDir, 0000))
	defer os.Chmod(noPermDir, 0755) // Cleanup

	// Try to resolve a path inside it.
	// filepath.EvalSymlinks should fail with permission denied on some systems
	// But on Linux root often bypasses permissions.
	// If running as non-root, this might work to trigger error.

	// If we are root, this test might not trigger the specific error branch we want,
	// but it's worth a try for coverage.

	err := IsAllowedPath(filepath.Join(noPermDir, "file.txt"))

	// If EvalSymlinks fails with permission, IsAllowedPath should return an error wrapping it.
	// However, our code has `if os.IsNotExist(err)`. Permission error is NOT NotExist.
	// So it should go to `else { return fmt.Errorf("failed to resolve symlinks ...") }`

	// If it succeeds (because root), then err will be nil (if IsPathTraversalSafe passes and it resolves to inside CWD).
	// But "noperm/file.txt" -> abs path inside CWD -> isInside CWD -> OK.

	// So to trigger error, we need EvalSymlinks to FAIL.
	// Creating a loop?
	// symlink loop -> EvalSymlinks returns error "too many levels of symbolic links"

	loopLink := "loop"
	require.NoError(t, os.Symlink(loopLink, loopLink))

	err = IsAllowedPath(loopLink)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to resolve symlinks")
}
