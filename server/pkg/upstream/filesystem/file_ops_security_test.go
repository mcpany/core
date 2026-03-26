// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package filesystem

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSafeRename_TOCTOU_Mitigation(t *testing.T) {
	// Setup filesystem
	baseDir := t.TempDir()
	sourceDir := filepath.Join(baseDir, "source")
	destDir := filepath.Join(baseDir, "dest")
	sensitiveDir := filepath.Join(baseDir, "sensitive")

	err := os.Mkdir(sourceDir, 0755)
	require.NoError(t, err)
	err = os.Mkdir(destDir, 0755)
	require.NoError(t, err)
	err = os.Mkdir(sensitiveDir, 0755)
	require.NoError(t, err)

	// Create source file
	sourceFile := filepath.Join(sourceDir, "file.txt")
	err = os.WriteFile(sourceFile, []byte("data"), 0600)
	require.NoError(t, err)

	// Create a subdirectory in destination that we will swap
	destSubDir := filepath.Join(destDir, "subdir")
	err = os.Mkdir(destSubDir, 0755)
	require.NoError(t, err)

	// The intended destination path
	targetPath := filepath.Join(destSubDir, "moved.txt")

	// ATTACK: Swap destSubDir with symlink to sensitiveDir
	err = os.RemoveAll(destSubDir)
	require.NoError(t, err)
	err = os.Symlink(sensitiveDir, destSubDir)
	require.NoError(t, err)

	// Attempt safeRename
	fs := afero.NewOsFs()
	err = safeRename(fs, sourceFile, targetPath)

	// Verify error
	assert.Error(t, err)
	if err != nil {
		assert.Contains(t, err.Error(), "destination path verification failed")
		assert.Contains(t, err.Error(), "path integrity violation")
	}

	// Verify file was NOT moved to sensitiveDir
	_, err = os.Stat(filepath.Join(sensitiveDir, "moved.txt"))
	assert.True(t, os.IsNotExist(err), "File should not be moved to sensitive directory")
}

func TestSafeRemove_TOCTOU_Mitigation(t *testing.T) {
	// Setup filesystem
	baseDir := t.TempDir()
	targetDir := filepath.Join(baseDir, "target")
	sensitiveDir := filepath.Join(baseDir, "sensitive")

	err := os.Mkdir(targetDir, 0755)
	require.NoError(t, err)
	err = os.Mkdir(sensitiveDir, 0755)
	require.NoError(t, err)

	// Create a file in sensitive dir
	sensitiveFile := filepath.Join(sensitiveDir, "config")
	err = os.WriteFile(sensitiveFile, []byte("secret"), 0600)
	require.NoError(t, err)

	// Create a subdirectory in target that we will swap
	targetSubDir := filepath.Join(targetDir, "subdir")
	err = os.Mkdir(targetSubDir, 0755)
	require.NoError(t, err)

	// The intended path to remove (as if we resolved it before swap)
	// Note: We are testing safeRemove logic which takes a path.
	// In the real tool, ResolvePath would have returned the path string.
	// If TOCTOU happens, the path string "targetSubDir/config" now points to "sensitiveDir/config".
	pathToRemove := filepath.Join(targetSubDir, "config")

	// ATTACK: Swap targetSubDir with symlink to sensitiveDir
	err = os.RemoveAll(targetSubDir)
	require.NoError(t, err)
	err = os.Symlink(sensitiveDir, targetSubDir)
	require.NoError(t, err)

	// Attempt safeRemove
	fs := afero.NewOsFs()
	err = safeRemove(fs, pathToRemove, false)

	// Verify error
	assert.Error(t, err)
	if err != nil {
		assert.Contains(t, err.Error(), "path verification failed")
		assert.Contains(t, err.Error(), "path integrity violation")
	}

	// Verify sensitive file was NOT removed
	_, err = os.Stat(sensitiveFile)
	assert.NoError(t, err, "Sensitive file should not be removed")
}
