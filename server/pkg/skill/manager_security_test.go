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

func TestSaveAsset_PathTraversal_SkillName(t *testing.T) {
	// Create a temp directory for skills
	rootDir := t.TempDir()
	manager, err := NewManager(rootDir)
	require.NoError(t, err)

	// Create a legitimate skill
	skill := &Skill{
		Frontmatter: Frontmatter{
			Name: "legit-skill",
		},
		Instructions: "Do legitimate things.",
	}
	require.NoError(t, manager.CreateSkill(skill))

	// Attempt to write outside the skill directory by using ".." as skillName

	// Create a target file outside rootDir to verify overwrite/creation
	tempTargetDir := t.TempDir()
	// rootDir is inside some temp folder. tempTargetDir is another.

	// Construct a relative path from rootDir to tempTargetDir
	rel, err := filepath.Rel(rootDir, tempTargetDir)
	require.NoError(t, err)

	// maliciousSkillName = rel (e.g. "../other_temp")
	maliciousSkillName := rel

	assetPath := "pwned.txt"
	content := []byte("PWNED")

	// This should FAIL if security checks are in place.
	err = manager.SaveAsset(maliciousSkillName, assetPath, content)

	// Assert that it SHOULD fail
	assert.Error(t, err, "SaveAsset should fail when skillName contains traversal characters or points outside allowed area")

	// Verification: check if file was written
	pwnedPath := filepath.Join(tempTargetDir, assetPath)
	if _, err := os.Stat(pwnedPath); err == nil {
		t.Logf("VULNERABILITY CONFIRMED: File written to %s", pwnedPath)
	}
}
