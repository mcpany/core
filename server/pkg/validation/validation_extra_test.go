// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package validation

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIsAllowedPath_Symlinks(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a safe directory
	safeDir := filepath.Join(tmpDir, "safe")
	err := os.Mkdir(safeDir, 0755)
	require.NoError(t, err)

	// Create a secret directory outside safeDir
	secretDir := filepath.Join(tmpDir, "secret")
	err = os.Mkdir(secretDir, 0755)
	require.NoError(t, err)

	// Create a secret file
	secretFile := filepath.Join(secretDir, "password.txt")
	err = os.WriteFile(secretFile, []byte("s3cr3t"), 0600)
	require.NoError(t, err)

	// Create a symlink inside safeDir pointing to secretDir
	symlinkPath := filepath.Join(safeDir, "link_to_secret")
	err = os.Symlink(secretDir, symlinkPath)
	require.NoError(t, err)

	// Set allowed paths to ONLY safeDir
	// Note: We need to temporarily set allowedPaths global variable
	// We can use SetAllowedPaths for this.
	defer SetAllowedPaths(nil)
	SetAllowedPaths([]string{safeDir})

	// 1. Accessing safeDir should be allowed
	err = IsAllowedPath(safeDir)
	require.NoError(t, err, "safeDir should be allowed")

	// 2. Accessing a file in safeDir should be allowed
	safeFile := filepath.Join(safeDir, "test.txt")
	err = os.WriteFile(safeFile, []byte("test"), 0600)
	require.NoError(t, err)
	err = IsAllowedPath(safeFile)
	require.NoError(t, err, "file in safeDir should be allowed")

	// 3. Accessing secretDir should be DENIED (it's not in allowed paths)
	err = IsAllowedPath(secretDir)
	require.Error(t, err, "secretDir should be denied")

	// 4. Accessing symlink inside safeDir pointing to secretDir should be DENIED
	// even though the symlink is inside safeDir, it resolves to secretDir which is outside.
	err = IsAllowedPath(symlinkPath)
	require.Error(t, err, "symlink to secretDir should be denied")

	// 5. Accessing file via symlink should be DENIED
	fileViaSymlink := filepath.Join(symlinkPath, "password.txt")
	err = IsAllowedPath(fileViaSymlink)
	require.Error(t, err, "file via symlink should be denied")
}

func TestIsAllowedPath_SymlinkLoop(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a loop
	link1 := filepath.Join(tmpDir, "link1")
	link2 := filepath.Join(tmpDir, "link2")

	_ = os.Symlink(link2, link1)
	_ = os.Symlink(link1, link2)

	defer SetAllowedPaths(nil)
	SetAllowedPaths([]string{tmpDir})

	// Accessing loop should fail or timeout, or return error from EvalSymlinks
	// We just want to make sure it doesn't crash or hang indefinitely (though EvalSymlinks handles loops)
	err := IsAllowedPath(link1)
	require.Error(t, err)
}
