// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package mcpserver

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/mcpany/core/server/pkg/resource"
	"github.com/mcpany/core/server/pkg/skill"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewSkillResource(t *testing.T) {
	s := &skill.Skill{
		Frontmatter: skill.Frontmatter{Name: "test-skill"},
	}
	r := NewSkillResource(s)
	assert.NotNil(t, r)
	assert.Equal(t, s, r.skill)
	assert.Empty(t, r.assetPath)
}

func TestNewSkillAssetResource(t *testing.T) {
	s := &skill.Skill{
		Frontmatter: skill.Frontmatter{Name: "test-skill"},
	}
	assetPath := "scripts/test.py"
	r := NewSkillAssetResource(s, assetPath)
	assert.NotNil(t, r)
	assert.Equal(t, s, r.skill)
	assert.Equal(t, assetPath, r.assetPath)
}

func TestSkillResource_Metadata(t *testing.T) {
	s := &skill.Skill{
		Frontmatter: skill.Frontmatter{
			Name:        "test-skill",
			Description: "A test skill",
		},
	}

	t.Run("Main Skill Resource", func(t *testing.T) {
		r := NewSkillResource(s)
		assert.Equal(t, "skills://test-skill/SKILL.md", r.URI())
		assert.Equal(t, "Skill: test-skill", r.Name())
		assert.Equal(t, "skills", r.Service())

		mcpRes := r.Resource()
		assert.Equal(t, "skills://test-skill/SKILL.md", mcpRes.URI)
		assert.Equal(t, "Skill: test-skill", mcpRes.Name)
		assert.Equal(t, "text/markdown", mcpRes.MIMEType)
		assert.Equal(t, "A test skill", mcpRes.Description)
	})

	t.Run("Asset Resource", func(t *testing.T) {
		r := NewSkillAssetResource(s, "data.json")
		assert.Equal(t, "skills://test-skill/data.json", r.URI())
		assert.Equal(t, "Skill Asset: data.json (test-skill)", r.Name())
		assert.Equal(t, "skills", r.Service())

		mcpRes := r.Resource()
		assert.Equal(t, "skills://test-skill/data.json", mcpRes.URI)
		assert.Equal(t, "application/json", mcpRes.MIMEType)
	})
}

func TestSkillResource_Read(t *testing.T) {
	// Setup temporary skill directory
	tmpDir := t.TempDir()
	skillDir := filepath.Join(tmpDir, "test-skill")
	err := os.MkdirAll(skillDir, 0755)
	require.NoError(t, err)

	// Create SKILL.md
	skillContent := "# Test Skill\nInstructions"
	err = os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte(skillContent), 0644)
	require.NoError(t, err)

	// Create a text asset
	scriptDir := filepath.Join(skillDir, "scripts")
	err = os.Mkdir(scriptDir, 0755)
	require.NoError(t, err)
	scriptContent := "print('hello')"
	err = os.WriteFile(filepath.Join(scriptDir, "hello.py"), []byte(scriptContent), 0644)
	require.NoError(t, err)

	// Create a binary asset (simulate)
	dataDir := filepath.Join(skillDir, "data")
	err = os.Mkdir(dataDir, 0755)
	require.NoError(t, err)
	binaryContent := []byte{0x00, 0x01, 0x02, 0x03}
	err = os.WriteFile(filepath.Join(dataDir, "data.bin"), binaryContent, 0644)
	require.NoError(t, err)

	// Create a symlink asset (valid)
	// target is relative to link location
	err = os.Symlink("../scripts/hello.py", filepath.Join(dataDir, "link_to_hello.py"))
	require.NoError(t, err)

	// Create a symlink asset (invalid - outside)
	outsideFile := filepath.Join(tmpDir, "outside.txt")
	err = os.WriteFile(outsideFile, []byte("secret"), 0644)
	require.NoError(t, err)
	err = os.Symlink("../../outside.txt", filepath.Join(dataDir, "link_to_outside.txt"))
	require.NoError(t, err)

	s := &skill.Skill{
		Frontmatter: skill.Frontmatter{Name: "test-skill"},
		Path:        skillDir,
	}

	ctx := context.Background()

	t.Run("Read SKILL.md", func(t *testing.T) {
		r := NewSkillResource(s)
		res, err := r.Read(ctx)
		require.NoError(t, err)
		require.Len(t, res.Contents, 1)
		assert.Equal(t, "skills://test-skill/SKILL.md", res.Contents[0].URI)
		assert.Equal(t, "text/markdown", res.Contents[0].MIMEType)
		assert.Equal(t, skillContent, res.Contents[0].Text)
	})

	t.Run("Read Text Asset", func(t *testing.T) {
		r := NewSkillAssetResource(s, "scripts/hello.py")
		res, err := r.Read(ctx)
		require.NoError(t, err)
		require.Len(t, res.Contents, 1)
		// MIME type detection might vary, usually text/plain or text/x-python if configured,
		// but Go's mime package uses OS specific /etc/mime.types or similar.
		// Since we don't control the environment fully, we check if it has Text content.
		assert.NotEmpty(t, res.Contents[0].Text)
		assert.Equal(t, scriptContent, res.Contents[0].Text)
	})

	t.Run("Read Binary Asset", func(t *testing.T) {
		r := NewSkillAssetResource(s, "data/data.bin")
		res, err := r.Read(ctx)
		require.NoError(t, err)
		require.Len(t, res.Contents, 1)
		assert.Empty(t, res.Contents[0].Text)
		assert.Equal(t, binaryContent, res.Contents[0].Blob)
	})

	t.Run("Read Symlink Asset (Valid)", func(t *testing.T) {
		r := NewSkillAssetResource(s, "data/link_to_hello.py")
		res, err := r.Read(ctx)
		require.NoError(t, err)
		require.Len(t, res.Contents, 1)
		assert.Equal(t, scriptContent, res.Contents[0].Text)
	})

	t.Run("Security: Path Traversal", func(t *testing.T) {
		r := NewSkillAssetResource(s, "../outside.txt")
		_, err := r.Read(ctx)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid asset path")
	})

	t.Run("Security: Symlink Traversal", func(t *testing.T) {
		r := NewSkillAssetResource(s, "data/link_to_outside.txt")
		_, err := r.Read(ctx)
		require.Error(t, err)
		// The error message might be "invalid path: points outside skill directory" or from EvalSymlinks if it fails logic
		assert.Contains(t, err.Error(), "invalid path")
	})

	t.Run("File Not Found", func(t *testing.T) {
		r := NewSkillAssetResource(s, "nonexistent.txt")
		_, err := r.Read(ctx)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "asset does not exist")
	})
}

func TestRegisterSkillResources(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a skill manually
	skillDir := filepath.Join(tmpDir, "my-skill")
	err := os.MkdirAll(skillDir, 0755)
	require.NoError(t, err)

	skillContent := "---\nname: my-skill\ndescription: A test skill\n---\n# Instructions"
	err = os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte(skillContent), 0644)
	require.NoError(t, err)

	// Create an asset
	assetPath := "asset.txt"
	err = os.WriteFile(filepath.Join(skillDir, assetPath), []byte("asset content"), 0644)
	require.NoError(t, err)

	// Init managers
	sm, err := skill.NewManager(tmpDir)
	require.NoError(t, err)

	rm := resource.NewManager()

	// Register
	err = RegisterSkillResources(rm, sm)
	require.NoError(t, err)

	// Verify resources
	resources := rm.ListResources()
	// Should have at least 2 resources: Main SKILL.md and asset.txt
	assert.GreaterOrEqual(t, len(resources), 2)

	// Check for main skill resource
	res, found := rm.GetResource("skills://my-skill/SKILL.md")
	assert.True(t, found)
	assert.Equal(t, "Skill: my-skill", res.Resource().Name)

	// Check for asset resource
	res, found = rm.GetResource("skills://my-skill/asset.txt")
	assert.True(t, found)
	assert.Equal(t, "Skill Asset: asset.txt (my-skill)", res.Resource().Name)
}
