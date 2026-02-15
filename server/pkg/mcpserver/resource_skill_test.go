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
		Frontmatter: skill.Frontmatter{
			Name: "test-skill",
		},
	}

	r := NewSkillResource(s)
	assert.NotNil(t, r)
	assert.Equal(t, "skills://test-skill/SKILL.md", r.URI())
}

func TestNewSkillAssetResource(t *testing.T) {
	s := &skill.Skill{
		Frontmatter: skill.Frontmatter{
			Name: "test-skill",
		},
	}

	r := NewSkillAssetResource(s, "scripts/test.py")
	assert.NotNil(t, r)
	assert.Equal(t, "skills://test-skill/scripts/test.py", r.URI())
}

func TestSkillResource_Metadata(t *testing.T) {
	s := &skill.Skill{
		Frontmatter: skill.Frontmatter{
			Name:        "test-skill",
			Description: "A test skill",
		},
	}

	// Main Resource
	r1 := NewSkillResource(s)
	assert.Equal(t, "skills://test-skill/SKILL.md", r1.URI())
	assert.Equal(t, "Skill: test-skill", r1.Name())
	assert.Equal(t, "skills", r1.Service())

	mcpRes1 := r1.Resource()
	assert.Equal(t, "text/markdown", mcpRes1.MIMEType)
	assert.Equal(t, "A test skill", mcpRes1.Description)

	// Asset Resource
	r2 := NewSkillAssetResource(s, "data.json")
	assert.Equal(t, "skills://test-skill/data.json", r2.URI())
	assert.Equal(t, "Skill Asset: data.json (test-skill)", r2.Name())

	mcpRes2 := r2.Resource()
	assert.Equal(t, "application/json", mcpRes2.MIMEType)
}

func TestSkillResource_Read(t *testing.T) {
	// Setup temporary skill directory
	tempDir := t.TempDir()
	skillDir := filepath.Join(tempDir, "test-skill")
	err := os.MkdirAll(skillDir, 0755)
	require.NoError(t, err)

	// Create SKILL.md
	skillContent := "# Test Skill\n\nInstructions..."
	err = os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte(skillContent), 0644)
	require.NoError(t, err)

	// Create assets
	assetsDir := filepath.Join(skillDir, "assets")
	err = os.MkdirAll(assetsDir, 0755)
	require.NoError(t, err)

	jsonContent := `{"key": "value"}`
	err = os.WriteFile(filepath.Join(assetsDir, "data.json"), []byte(jsonContent), 0644)
	require.NoError(t, err)

	binContent := []byte{0x00, 0x01, 0x02, 0x03}
	err = os.WriteFile(filepath.Join(assetsDir, "data.bin"), binContent, 0644)
	require.NoError(t, err)

	// Create a file outside the skill directory
	outsideFile := filepath.Join(tempDir, "outside.txt")
	err = os.WriteFile(outsideFile, []byte("SECRET"), 0644)
	require.NoError(t, err)

	// Create a symlink inside the skill pointing to the outside file
	err = os.Symlink(outsideFile, filepath.Join(assetsDir, "bad_link"))
	require.NoError(t, err)

	// Create a valid internal symlink
	err = os.Symlink("data.json", filepath.Join(assetsDir, "good_link.json"))
	require.NoError(t, err)

	// Prepare Skill struct
	s := &skill.Skill{
		Frontmatter: skill.Frontmatter{Name: "test-skill"},
		Path:        skillDir,
	}

	t.Run("Read SKILL.md", func(t *testing.T) {
		r := NewSkillResource(s)
		res, err := r.Read(context.Background())
		require.NoError(t, err)
		require.Len(t, res.Contents, 1)
		assert.Equal(t, "text/markdown", res.Contents[0].MIMEType)
		assert.Equal(t, skillContent, res.Contents[0].Text)
	})

	t.Run("Read Asset JSON", func(t *testing.T) {
		r := NewSkillAssetResource(s, "assets/data.json")
		res, err := r.Read(context.Background())
		require.NoError(t, err)
		require.Len(t, res.Contents, 1)
		assert.Equal(t, "application/json", res.Contents[0].MIMEType)
		assert.Equal(t, jsonContent, res.Contents[0].Text)
	})

	t.Run("Read Asset Binary", func(t *testing.T) {
		r := NewSkillAssetResource(s, "assets/data.bin")
		res, err := r.Read(context.Background())
		require.NoError(t, err)
		require.Len(t, res.Contents, 1)
		assert.Equal(t, "application/octet-stream", res.Contents[0].MIMEType)
		assert.Equal(t, binContent, res.Contents[0].Blob)
	})

	t.Run("Read Symlink Internal (Valid)", func(t *testing.T) {
		r := NewSkillAssetResource(s, "assets/good_link.json")
		res, err := r.Read(context.Background())
		require.NoError(t, err)
		require.Len(t, res.Contents, 1)
		assert.Equal(t, jsonContent, res.Contents[0].Text)
	})

	t.Run("Security: Path Traversal", func(t *testing.T) {
		r := NewSkillAssetResource(s, "../outside.txt")
		_, err := r.Read(context.Background())
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid asset path")
	})

	t.Run("Security: Symlink Escape", func(t *testing.T) {
		r := NewSkillAssetResource(s, "assets/bad_link")
		_, err := r.Read(context.Background())
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "points outside skill directory")
	})

	t.Run("Error: File Not Found", func(t *testing.T) {
		r := NewSkillAssetResource(s, "assets/missing.txt")
		_, err := r.Read(context.Background())
		assert.Error(t, err)
		assert.ErrorIs(t, err, os.ErrNotExist)
	})
}

func TestRegisterSkillResources(t *testing.T) {
	// Setup Managers
	tempDir := t.TempDir()
	sm, err := skill.NewManager(tempDir)
	require.NoError(t, err)

	rm := resource.NewManager()

	// Create a skill
	s := &skill.Skill{
		Frontmatter: skill.Frontmatter{
			Name: "my-skill",
			Description: "desc",
		},
		Instructions: "Do things",
	}
	err = sm.CreateSkill(s)
	require.NoError(t, err)

	// Create an asset
	err = sm.SaveAsset("my-skill", "script.py", []byte("print('hi')"))
	require.NoError(t, err)

	// Register
	err = RegisterSkillResources(rm, sm)
	require.NoError(t, err)

	// Verify
	resources := rm.ListResources()
	// Should have main skill + 1 asset = 2 resources
	assert.Len(t, resources, 2)

	// Check main skill
	res1, found := rm.GetResource("skills://my-skill/SKILL.md")
	assert.True(t, found)
	assert.Equal(t, "Skill: my-skill", res1.Resource().Name)

	// Check asset
	res2, found := rm.GetResource("skills://my-skill/script.py")
	assert.True(t, found)
	assert.Equal(t, "Skill Asset: script.py (my-skill)", res2.Resource().Name)
}
