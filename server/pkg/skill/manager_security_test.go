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

func TestManager_Security_PathTraversal(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "skills-sec-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	_, err = NewManager(tempDir)
	require.NoError(t, err)

	t.Run("SaveAsset Path Traversal via SkillName", func(t *testing.T) {
		// Attempt to write to a file outside the skills directory by using ".." in skillName
		// We try to write to "pwned.txt" in the tempDir (parent of skills dir)
		// skills dir is tempDir/skills (if we initialized it that way, but NewManager takes rootDir)
		// So if rootDir is tempDir, ".." goes to system temp.
		// Let's create a dedicated root for manager.
		managerRoot := filepath.Join(tempDir, "skills")
		m2, err := NewManager(managerRoot)
		require.NoError(t, err)

		// Create a valid skill first to ensure dir structure exists?
		// No, we are attacking the skillName.
		// If skillName is "../", it points to tempDir.

		// The vulnerability: SaveAsset joins rootDir + skillName.
		// If skillName is "..", path is tempDir.
		// Then joins with assetPath "pwned.txt".
		// Result: tempDir/pwned.txt.

		// If we fixed it, this should fail because ".." is invalid name.
		err = m2.SaveAsset("..", "pwned.txt", []byte("pwned"))

		// Assert that it returns an error
		assert.Error(t, err, "SaveAsset should reject '..' as skill name")

		// Verify file was NOT created
		_, statErr := os.Stat(filepath.Join(tempDir, "pwned.txt"))
		assert.True(t, os.IsNotExist(statErr), "File should not exist outside skills directory")
	})

	t.Run("GetSkill Path Traversal", func(t *testing.T) {
		managerRoot := filepath.Join(tempDir, "skills_get")
		m3, err := NewManager(managerRoot)
		require.NoError(t, err)

		// Create a file outside
		outsideFile := filepath.Join(tempDir, "SKILL.md") // tempDir is parent
		err = os.WriteFile(outsideFile, []byte("---\nname: pwn\n---\nbody"), 0644)
		require.NoError(t, err)

		// Try to read it via ".."
		_, err = m3.GetSkill("..")
		assert.Error(t, err, "GetSkill should reject '..' as skill name")
	})

	t.Run("DeleteSkill Path Traversal", func(t *testing.T) {
		managerRoot := filepath.Join(tempDir, "skills_del")
		m4, err := NewManager(managerRoot)
		require.NoError(t, err)

		// Create a file outside
		targetDir := filepath.Join(tempDir, "target_to_delete")
		err = os.Mkdir(targetDir, 0755)
		require.NoError(t, err)

		// Try to delete it via "../target_to_delete"
		// Wait, DeleteSkill joins root + name.
		// If name is "../target_to_delete", it resolves to tempDir/target_to_delete.

		err = m4.DeleteSkill("../target_to_delete")
		assert.Error(t, err, "DeleteSkill should reject traversal in skill name")

		// Verify dir still exists
		_, statErr := os.Stat(targetDir)
		assert.False(t, os.IsNotExist(statErr), "Directory should still exist")
	})
}
