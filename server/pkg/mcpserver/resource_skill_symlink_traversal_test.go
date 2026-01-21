package mcpserver_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/mcpany/core/server/pkg/mcpserver"
	"github.com/mcpany/core/server/pkg/skill"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSkillResource_SymlinkTraversal(t *testing.T) {
	// Setup
	tmpDir, err := os.MkdirTemp("", "skill-symlink-test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	skillDir := filepath.Join(tmpDir, "myskill")
	err = os.Mkdir(skillDir, 0755)
	require.NoError(t, err)

	// Create a secret file outside the skill directory
	secretFile := filepath.Join(tmpDir, "secret.txt")
	err = os.WriteFile(secretFile, []byte("super secret"), 0644)
	require.NoError(t, err)

	// Create a symlink inside the skill directory pointing to the secret file
	linkFile := filepath.Join(skillDir, "link_to_secret.txt")
	err = os.Symlink(secretFile, linkFile)
	require.NoError(t, err)

	s := &skill.Skill{
		Frontmatter: skill.Frontmatter{
			Name: "myskill",
		},
		Path: skillDir,
	}

	// Test Case: Symlink Traversal Attack
	// We try to access link_to_secret.txt which is a valid filename inside the directory,
	// but points outside.
	resource := mcpserver.NewSkillAssetResource(s, "link_to_secret.txt")
	_, err = resource.Read(context.Background())

	// Expectation: Should fail with an error indicating invalid path
	assert.Error(t, err)
	assert.ErrorContains(t, err, "invalid path")
}
