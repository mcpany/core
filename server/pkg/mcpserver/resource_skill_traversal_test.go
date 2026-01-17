// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

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

func TestSkillResource_PathTraversal(t *testing.T) {
	// Setup
	tmpDir, err := os.MkdirTemp("", "skill-test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	skillDir := filepath.Join(tmpDir, "myskill")
	err = os.Mkdir(skillDir, 0755)
	require.NoError(t, err)

	// Create a secret file outside the skill directory
	secretFile := filepath.Join(tmpDir, "secret.txt")
	err = os.WriteFile(secretFile, []byte("super secret"), 0644)
	require.NoError(t, err)

	// Create a sibling directory that shares prefix
	siblingDir := filepath.Join(tmpDir, "myskill-sibling")
	err = os.Mkdir(siblingDir, 0755)
	require.NoError(t, err)
	siblingFile := filepath.Join(siblingDir, "secret.txt")
	err = os.WriteFile(siblingFile, []byte("sibling secret"), 0644)
	require.NoError(t, err)

	// Create a valid skill asset
	assetFile := filepath.Join(skillDir, "asset.txt")
	err = os.WriteFile(assetFile, []byte("public asset"), 0644)
	require.NoError(t, err)

	s := &skill.Skill{
		Frontmatter: skill.Frontmatter{
			Name: "myskill",
		},
		Path: skillDir,
	}

	// Test Case 1: Valid Access
	resource := mcpserver.NewSkillAssetResource(s, "asset.txt")
	result, err := resource.Read(context.Background())
	require.NoError(t, err)
	assert.Equal(t, "public asset", result.Contents[0].Text)

	// Test Case 2: Path Traversal Attack (Dot Dot)
	// We try to access ../secret.txt
	resource = mcpserver.NewSkillAssetResource(s, "../secret.txt")
	result, err = resource.Read(context.Background())

	// Expectation: Should fail
	if err == nil {
		assert.Fail(t, "Should prevent path traversal with ..")
	} else {
		assert.ErrorContains(t, err, "invalid asset path")
	}

	// Test Case 3: Sibling Directory Attack
	// We try to access ../myskill-sibling/secret.txt
	relPath, err := filepath.Rel(skillDir, siblingFile)
	require.NoError(t, err)

	resource = mcpserver.NewSkillAssetResource(s, relPath)
	result, err = resource.Read(context.Background())

	// Expectation: Should fail
	if err == nil {
		assert.Fail(t, "Should prevent access to sibling directory")
	} else {
		assert.ErrorContains(t, err, "invalid asset path")
	}
}
