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

func TestSecurity_PathTraversal(t *testing.T) {
	// Create a temp directory for the test
	tempDir, err := os.MkdirTemp("", "skill-security-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create a "skills" directory inside the temp directory
	skillsDir := filepath.Join(tempDir, "skills")
	err = os.Mkdir(skillsDir, 0755)
	require.NoError(t, err)

	m, err := NewManager(skillsDir)
	require.NoError(t, err)

	t.Run("SaveAsset Traversal via SkillName", func(t *testing.T) {
		targetFile := "hacked.txt"
		err := m.SaveAsset("..", targetFile, []byte("pwned"))

		// Expect error. If nil, we have a vulnerability.
		assert.Error(t, err, "Should fail to save asset with traversal name")
		if err != nil {
			assert.Contains(t, err.Error(), "invalid skill name")
		} else {
			// Clean up if we accidentally wrote it
			_ = os.Remove(filepath.Join(tempDir, targetFile))
		}
	})

	t.Run("GetSkill Traversal", func(t *testing.T) {
		// Create a fake SKILL.md in tempDir
		fakeSkill := filepath.Join(tempDir, SkillFileName)
		err := os.WriteFile(fakeSkill, []byte("---\nname: hacked\n---\nbody"), 0644)
		require.NoError(t, err)

		s, err := m.GetSkill("..")
		assert.Error(t, err, "Should fail to get skill with traversal name")
		if err == nil {
			assert.NotEqual(t, "..", s.Name, "Should not return skill with name '..'")
		} else {
			assert.Contains(t, err.Error(), "invalid skill name")
		}
	})

	t.Run("DeleteSkill Traversal", func(t *testing.T) {
		// Create a target dir
		targetDir := filepath.Join(tempDir, "target-to-delete")
		err := os.Mkdir(targetDir, 0755)
		require.NoError(t, err)

		err = m.DeleteSkill("../target-to-delete")
		assert.Error(t, err, "Should fail to delete skill with traversal name")
		if err != nil {
			 assert.Contains(t, err.Error(), "invalid skill name")
		}
	})
}
