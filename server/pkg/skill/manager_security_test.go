// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package skill

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestManagerSecurity(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "skills-sec-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	m, err := NewManager(tempDir)
	require.NoError(t, err)

	// Setup: Create a file outside the skills directory to verify we can't overwrite it
	outsideFile := filepath.Join(tempDir, "..", "target_file.txt")
	// Make sure we are not in root
	if filepath.Dir(tempDir) == "/" {
		// Fallback for safety in some envs
		outsideFile = filepath.Join(os.TempDir(), "target_file.txt")
	}

	// Ensure we start clean
	_ = os.Remove(outsideFile)

	t.Run("SaveAsset SkillName Traversal", func(t *testing.T) {
		// Try to use ".." as skillName to write to parent directory
		// The path we are trying to write to is: <rootDir>/../asset.txt
		// If skillName="..", relPath="asset.txt"

		// This should fail because ".." is not a valid skill name
		err := m.SaveAsset("..", "asset.txt", []byte("hacked"))
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid skill name")

		// Verify file was not created
		_, err = os.Stat(filepath.Join(tempDir, "..", "asset.txt"))
		assert.True(t, os.IsNotExist(err))
	})

	t.Run("SaveAsset Absolute SkillName", func(t *testing.T) {
		// Try to use absolute path as skillName
		// This should fail because "/" is not allowed in skill name
		err := m.SaveAsset("/etc", "passwd", []byte("hacked"))
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid skill name")
	})

	t.Run("SaveAsset Complex Traversal", func(t *testing.T) {
		// Try to use a skill name that looks valid but contains traversal if not validated properly
		// But validateName enforces specific regex.
		// "skill/../../"
		err := m.SaveAsset("skill/../../", "asset.txt", []byte("hacked"))
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid skill name")
	})
}
