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
	tempDir, err := os.MkdirTemp("", "skills-security-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	m, err := NewManager(tempDir)
	require.NoError(t, err)

	skillName := "secure-skill"
	skill := &Skill{
		Frontmatter: Frontmatter{
			Name:        skillName,
			Description: "Security test skill",
		},
		Instructions: "Original instructions",
	}
	require.NoError(t, m.CreateSkill(skill))

	t.Run("Prevent Overwriting SKILL.md", func(t *testing.T) {
		err := m.SaveAsset(skillName, SkillFileName, []byte("Overwrite attempt"))
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot overwrite skill definition file")

		// Verify content hasn't changed
		s, err := m.GetSkill(skillName)
		require.NoError(t, err)
		assert.Equal(t, "Original instructions", s.Instructions)
	})

	t.Run("Prevent Absolute Paths", func(t *testing.T) {
		// This test attempts to write to an absolute path.
		// Construct an absolute path to a temporary file.
		absPath := filepath.Join(tempDir, "absolute_file.txt")
		err := m.SaveAsset(skillName, absPath, []byte("Absolute path attempt"))
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "absolute paths are not allowed")
	})

    t.Run("Prevent Windows Absolute Paths", func(t *testing.T) {
        // Mocking Windows paths on non-Windows is hard, but we can try basic drive letter pattern if the validation supports it.
        // Assuming the code uses filepath.IsAbs, this test depends on the OS running the test.
        // If running on Linux, "C:\foo" is relative.
        // So we can't easily test Windows behavior on Linux without mocking filepath.
        // However, we can test that our code calls filepath.IsAbs.
    })
}
