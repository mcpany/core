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

func TestManager_Cache(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "skills-cache-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	m, err := NewManager(tempDir)
	require.NoError(t, err)

	// 1. Create a skill via Manager (Cold start / Invalidate)
	skill := &Skill{
		Frontmatter: Frontmatter{
			Name: "skill-1",
		},
		Instructions: "instr 1",
	}
	err = m.CreateSkill(skill)
	require.NoError(t, err)

	// 2. List (Cache Miss -> Populate Cache)
	skills, err := m.ListSkills()
	require.NoError(t, err)
	assert.Len(t, skills, 1)
	assert.Equal(t, "skill-1", skills[0].Name)

	// 3. Modify file on disk directly (Backdoor)
	// This simulates an external change or just proving we are reading from cache
	skillDir := filepath.Join(tempDir, "skill-1")
	// Add an asset file directly
	err = os.WriteFile(filepath.Join(skillDir, "backdoor.txt"), []byte("secret"), 0644)
	require.NoError(t, err)

	// 4. List again (Cache Hit)
	// The new asset should NOT be in the list because we are serving from cache
	skillsCached, err := m.ListSkills()
	require.NoError(t, err)
	assert.Len(t, skillsCached, 1)

	// We check assets. Since ListSkills calls loadSkill which walks dir,
	// if it wasn't cached, it would find "backdoor.txt".
	found := false
	for _, asset := range skillsCached[0].Assets {
		if asset == "backdoor.txt" {
			found = true
			break
		}
	}
	assert.False(t, found, "Should not find backdoor.txt if served from cache")

	// 5. Invalidate Cache via UpdateSkill
	update := &Skill{
		Frontmatter: Frontmatter{
			Name: "skill-1",
		},
		Instructions: "instr 2",
	}
	err = m.UpdateSkill("skill-1", update)
	require.NoError(t, err)

	// 6. List again (Cache Miss -> Reload)
	skillsReloaded, err := m.ListSkills()
	require.NoError(t, err)
	assert.Len(t, skillsReloaded, 1)
	assert.Equal(t, "instr 2", skillsReloaded[0].Instructions)

	// Now it SHOULD find the backdoor file because it re-scanned the directory
	found = false
	for _, asset := range skillsReloaded[0].Assets {
		if asset == "backdoor.txt" {
			found = true
			break
		}
	}
	assert.True(t, found, "Should find backdoor.txt after cache invalidation")
}
