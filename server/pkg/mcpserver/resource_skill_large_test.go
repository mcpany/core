package mcpserver

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/mcpany/core/server/pkg/consts"
	"github.com/mcpany/core/server/pkg/skill"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSkillResource_Read_LargeFile(t *testing.T) {
	// Create a temporary directory for the skill
	tempDir, err := os.MkdirTemp("", "skill_large_test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create a dummy large file
	// We make it 1 byte larger than the limit
	limit := int64(consts.DefaultMaxResourceSizeBytes)
	largeFileSize := limit + 1
	largeFileName := "large_asset.txt"
	largeFilePath := filepath.Join(tempDir, largeFileName)

	f, err := os.Create(largeFilePath)
	require.NoError(t, err)

	// Seek to the end - 1 and write a byte to make it sparse but large
	// This is faster than writing 10MB
	_, err = f.Seek(largeFileSize-1, 0)
	require.NoError(t, err)
	_, err = f.Write([]byte{0})
	require.NoError(t, err)
	f.Close()

	// Create the skill object
	s := &skill.Skill{
		Frontmatter: skill.Frontmatter{
			Name: "test-skill",
		},
		Path: tempDir,
	}

	// Create the resource
	r := NewSkillAssetResource(s, largeFileName)

	// Attempt to read
	_, err = r.Read(context.Background())

	// Assert error
	require.Error(t, err)
	assert.Contains(t, err.Error(), "resource too large")
}
