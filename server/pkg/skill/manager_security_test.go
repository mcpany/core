// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package skill

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSaveAsset_Traversal(t *testing.T) {
	// Create a temp root dir for skills
	tmpDir := t.TempDir()

	// Initialize Manager
	// We use a subdirectory for skills so we can verify writing to tmpDir (parent)
	skillsRoot := filepath.Join(tmpDir, "skills")
	err := os.Mkdir(skillsRoot, 0755)
	require.NoError(t, err)

	m, err := NewManager(skillsRoot)
	require.NoError(t, err)

	// Create a legitimate skill
	validSkill := &Skill{
		Frontmatter: Frontmatter{
			Name: "test-skill",
			Description: "Test Skill",
		},
		Instructions: "Do something.",
	}
	err = m.CreateSkill(validSkill)
	require.NoError(t, err)

	// Attempt to save an asset using path traversal via skillName
	// This simulates an attack where the skillName is manipulated to be ".."

	evilContent := []byte("pwned")

	// Now try to write to tmpDir/pwned.txt using skillName=".."
	// skillName=".." -> skillDir = skillsRoot/.. = tmpDir
	// relPath="pwned.txt" -> fullPath = tmpDir/pwned.txt
	err = m.SaveAsset("..", "pwned.txt", evilContent)

	// If vulnerable, err is nil.
	// We expect an error here.
	if err == nil {
		t.Fatalf("Vulnerability confirmed: Successfully wrote outside skill directory")
	}
	t.Logf("Secure: Failed to write outside skill directory: %v", err)
}
