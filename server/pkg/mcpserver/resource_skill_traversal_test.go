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

	// Test Case 2: Path Traversal Attack
	// We try to access ../secret.txt
	resource = mcpserver.NewSkillAssetResource(s, "../secret.txt")
	result, err = resource.Read(context.Background())

	// Expectation: Should fail (but currently succeeds due to vulnerability)
	if err == nil {
		// If it succeeds, check if we read the secret
		if result.Contents[0].Text == "super secret" {
			t.Log("VULNERABILITY CONFIRMED: Path traversal allowed access to secret file")
			// We assert fail here to confirm the test is working as a reproduction
			assert.Fail(t, "Should prevent path traversal")
		} else {
			t.Logf("Read succeeded but content unexpected: %s", result.Contents[0].Text)
		}
	} else {
		// Once fixed, it should error
		assert.ErrorContains(t, err, "invalid asset path")
	}
}
