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

func TestSaveAsset_PathTraversal_Protection(t *testing.T) {
	tempDir := t.TempDir()
	// Create a critical file outside skills directory
	criticalFile := filepath.Join(tempDir, "config.yaml")
	require.NoError(t, os.WriteFile(criticalFile, []byte("original config"), 0644))

	skillsDir := filepath.Join(tempDir, "skills")
	manager, err := NewManager(skillsDir)
	require.NoError(t, err)

	// Attempt to overwrite criticalFile using traversal in skillName
	// skillName = "../"
	// relPath = "config.yaml"
	// Target: skillsDir/../config.yaml -> tempDir/config.yaml

	err = manager.SaveAsset("..", "config.yaml", []byte("hacked config"))

	// Expectation: This SHOULD fail with invalid skill name
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid skill name")

	// Verify file was NOT overwritten
	content, _ := os.ReadFile(criticalFile)
	assert.Equal(t, "original config", string(content))
}

func TestGetSkill_PathTraversal_Protection(t *testing.T) {
	tempDir := t.TempDir()
	skillsDir := filepath.Join(tempDir, "skills")
	manager, err := NewManager(skillsDir)
	require.NoError(t, err)

	// Attempt to read something via traversal
	_, err = manager.GetSkill("..")

	// Expectation: This SHOULD fail with invalid skill name
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid skill name")
}
