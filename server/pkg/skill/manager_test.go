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

func TestManager(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "skills-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	m, err := NewManager(tempDir)
	require.NoError(t, err)

	t.Run("Create and List", func(t *testing.T) {
		skill := &Skill{
			Frontmatter: Frontmatter{
				Name:         "test-skill",
				Description:  "A test skill",
				AllowedTools: []string{"tool1"},
			},
			Instructions: "# Instructions\nDo something.",
		}

		err := m.CreateSkill(skill)
		require.NoError(t, err)

		skills, err := m.ListSkills()
		require.NoError(t, err)
		assert.Len(t, skills, 1)
		assert.Equal(t, "test-skill", skills[0].Name)
		assert.Equal(t, "A test skill", skills[0].Description)
		assert.Equal(t, "tool1", skills[0].AllowedTools[0])
		assert.Contains(t, skills[0].Instructions, "# Instructions")

		// Verify file content
		content, err := os.ReadFile(filepath.Join(tempDir, "test-skill", SkillFileName))
		require.NoError(t, err)
		assert.Contains(t, string(content), "name: test-skill")
	})

	t.Run("Get Skill", func(t *testing.T) {
		s, err := m.GetSkill("test-skill")
		require.NoError(t, err)
		assert.Equal(t, "test-skill", s.Name)
	})

	t.Run("Update Skill", func(t *testing.T) {
		skill := &Skill{
			Frontmatter: Frontmatter{
				Name:        "test-skill-updated",
				Description: "Updated description",
			},
			Instructions: "Updated instructions",
		}

		err := m.UpdateSkill("test-skill", skill)
		require.NoError(t, err)

		// Old name should be gone
		_, err = m.GetSkill("test-skill")
		assert.Error(t, err)

		// New name should exist
		s, err := m.GetSkill("test-skill-updated")
		require.NoError(t, err)
		assert.Equal(t, "test-skill-updated", s.Name)
		assert.Equal(t, "Updated description", s.Description)
	})

	t.Run("Upload Asset", func(t *testing.T) {
		err := m.SaveAsset("test-skill-updated", "scripts/test.py", []byte("print('hello')"))
		require.NoError(t, err)

		// Check file exists
		content, err := os.ReadFile(filepath.Join(tempDir, "test-skill-updated", "scripts", "test.py"))
		require.NoError(t, err)
		assert.Equal(t, "print('hello')", string(content))

		// Check asset listing
		s, err := m.GetSkill("test-skill-updated")
		require.NoError(t, err)
		assert.Contains(t, s.Assets, "scripts/test.py")
	})

	t.Run("Delete Skill", func(t *testing.T) {
		err := m.DeleteSkill("test-skill-updated")
		require.NoError(t, err)

		skills, err := m.ListSkills()
		require.NoError(t, err)
		assert.Empty(t, skills)

		// Directory should be gone
		_, err = os.Stat(filepath.Join(tempDir, "test-skill-updated"))
		assert.True(t, os.IsNotExist(err))
	})

	t.Run("Create Invalid Name", func(t *testing.T) {
		invalidNames := []string{"", "Test", "test space", "test--dash", "-start", "end-"}
		for _, name := range invalidNames {
			skill := &Skill{Frontmatter: Frontmatter{Name: name}}
			err := m.CreateSkill(skill)
			assert.Error(t, err, "name %q should be invalid", name)
		}
	})

	t.Run("Create Duplicate", func(t *testing.T) {
		skill := &Skill{Frontmatter: Frontmatter{Name: "dup-skill"}}
		err := m.CreateSkill(skill)
		require.NoError(t, err)

		err = m.CreateSkill(skill)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "already exists")
	})

	t.Run("Update Not Found", func(t *testing.T) {
		skill := &Skill{Frontmatter: Frontmatter{Name: "non-existent"}}
		err := m.UpdateSkill("non-existent", skill)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})

	t.Run("Update to Existing Name", func(t *testing.T) {
		// Create two skills
		s1 := &Skill{Frontmatter: Frontmatter{Name: "s1"}}
		require.NoError(t, m.CreateSkill(s1))
		s2 := &Skill{Frontmatter: Frontmatter{Name: "s2"}}
		require.NoError(t, m.CreateSkill(s2))

		// Try to rename s1 to s2
		s1Update := &Skill{Frontmatter: Frontmatter{Name: "s2"}}
		err := m.UpdateSkill("s1", s1Update)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "already exists")
	})

	t.Run("Asset Path Traversal", func(t *testing.T) {
		// Create a skill first
		s := &Skill{Frontmatter: Frontmatter{Name: "asset-skill"}}
		require.NoError(t, m.CreateSkill(s))

		err := m.SaveAsset("asset-skill", "../outside.txt", []byte("bad"))
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid asset path")

		err = m.SaveAsset("asset-skill", "/abs/path", []byte("bad"))
		assert.Error(t, err)
	})

	t.Run("Load Malformed Skill", func(t *testing.T) {
		// Manually create a bad skill file
		dir := filepath.Join(tempDir, "bad-skill")
		err := os.MkdirAll(dir, 0755)
		require.NoError(t, err)

		// Write invalid YAML
		err = os.WriteFile(filepath.Join(dir, SkillFileName), []byte("---: bad\n---\nBody"), 0644)
		require.NoError(t, err)

		_, err = m.GetSkill("bad-skill")
		assert.Error(t, err)
	})
}
