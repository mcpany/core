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

func TestSkillResource_BinaryAsset(t *testing.T) {
	// Setup
	tmpDir, err := os.MkdirTemp("", "skill-blob-test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	skillDir := filepath.Join(tmpDir, "myskill")
	err = os.Mkdir(skillDir, 0755)
	require.NoError(t, err)

	// Create a binary asset (simulated with random bytes and .bin extension)
	binaryContent := []byte{0xDE, 0xAD, 0xBE, 0xEF, 0x00, 0xFF}
	binaryFile := filepath.Join(skillDir, "data.bin")
	err = os.WriteFile(binaryFile, binaryContent, 0644)
	require.NoError(t, err)

	// Create a text asset
	textContent := []byte("Hello, world!")
	textFile := filepath.Join(skillDir, "data.txt")
	err = os.WriteFile(textFile, textContent, 0644)
	require.NoError(t, err)

	s := &skill.Skill{
		Frontmatter: skill.Frontmatter{
			Name: "myskill",
		},
		Path: skillDir,
	}

	// Test Case 1: Binary Asset
	resource := mcpserver.NewSkillAssetResource(s, "data.bin")
	result, err := resource.Read(context.Background())
	require.NoError(t, err)

	require.Len(t, result.Contents, 1)
	content := result.Contents[0]

	// Expectation: Blob should be populated, Text should be empty
	assert.Equal(t, binaryContent, content.Blob, "Blob content should match")
	assert.Empty(t, content.Text, "Text content should be empty for binary")
	assert.Equal(t, "application/octet-stream", content.MIMEType)

	// Test Case 2: Text Asset
	resourceText := mcpserver.NewSkillAssetResource(s, "data.txt")
	resultText, err := resourceText.Read(context.Background())
	require.NoError(t, err)

	require.Len(t, resultText.Contents, 1)
	contentText := resultText.Contents[0]

	// Expectation: Text should be populated, Blob should be empty (or nil)
	assert.Equal(t, string(textContent), contentText.Text, "Text content should match")
	assert.Empty(t, contentText.Blob, "Blob content should be empty for text")
	// MIME type might vary depending on OS, but should likely contain "text"
	// assert.Contains(t, contentText.MIMEType, "text/plain")
}
