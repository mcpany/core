// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package mcpserver

import (
	"context"
	"os"
	"testing"

	"github.com/mcpany/core/server/pkg/resource"
	"github.com/mcpany/core/server/pkg/skill"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func setupTestSkillManager(t *testing.T) (*skill.Manager, string) {
	tempDir, err := os.MkdirTemp("", "skill_resource_test")
	require.NoError(t, err)

	manager, err := skill.NewManager(tempDir)
	require.NoError(t, err)

	t.Cleanup(func() {
		os.RemoveAll(tempDir)
	})

	return manager, tempDir
}

func TestSkillResource_URI(t *testing.T) {
	s := &skill.Skill{
		Frontmatter: skill.Frontmatter{Name: "my-skill"},
	}

	r := NewSkillResource(s)
	assert.Equal(t, "skills://my-skill/SKILL.md", r.URI())

	assetR := NewSkillAssetResource(s, "scripts/test.py")
	assert.Equal(t, "skills://my-skill/scripts/test.py", assetR.URI())
}

func TestSkillResource_Name(t *testing.T) {
	s := &skill.Skill{
		Frontmatter: skill.Frontmatter{Name: "my-skill"},
	}

	r := NewSkillResource(s)
	assert.Equal(t, "Skill: my-skill", r.Name())

	assetR := NewSkillAssetResource(s, "scripts/test.py")
	assert.Equal(t, "Skill Asset: scripts/test.py (my-skill)", assetR.Name())
}

func TestSkillResource_Read(t *testing.T) {
	manager, _ := setupTestSkillManager(t)

	err := manager.CreateSkill(&skill.Skill{
		Frontmatter: skill.Frontmatter{Name: "read-test"},
		Instructions: "hello world",
	})
	require.NoError(t, err)

	err = manager.SaveAsset("read-test", "data.txt", []byte("some data"))
	require.NoError(t, err)

	sk, err := manager.GetSkill("read-test")
	require.NoError(t, err)

	t.Run("Read SKILL.md", func(t *testing.T) {
		r := NewSkillResource(sk)
		res, err := r.Read(context.Background())
		require.NoError(t, err)
		require.Len(t, res.Contents, 1)
		assert.Contains(t, res.Contents[0].Text, "hello world")
		assert.Contains(t, res.Contents[0].Text, "name: read-test") // Frontmatter
		assert.Equal(t, "text/markdown", res.Contents[0].MIMEType)
	})

	t.Run("Read Asset", func(t *testing.T) {
		r := NewSkillAssetResource(sk, "data.txt")
		res, err := r.Read(context.Background())
		require.NoError(t, err)
		require.Len(t, res.Contents, 1)
		assert.Equal(t, "some data", res.Contents[0].Text)
		assert.Equal(t, "text/plain; charset=utf-8", res.Contents[0].MIMEType)
	})

	t.Run("Read Invalid Path", func(t *testing.T) {
		r := NewSkillAssetResource(sk, "../other/secret.txt")
		_, err := r.Read(context.Background())
		assert.Error(t, err)
	})
}

func TestRegisterSkillResources(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRM := resource.NewMockManagerInterface(ctrl)
	manager, _ := setupTestSkillManager(t)

	// Create a skill with an asset
	err := manager.CreateSkill(&skill.Skill{
		Frontmatter: skill.Frontmatter{Name: "reg-test"},
		Instructions: "inst",
	})
	require.NoError(t, err)
	err = manager.SaveAsset("reg-test", "script.py", []byte("print(1)"))
	require.NoError(t, err)

	// Expect 2 AddResource calls: one for SKILL.md, one for script.py
	mockRM.EXPECT().AddResource(gomock.Any()).Times(2).Do(func(r resource.Resource) {
		uri := r.Resource().URI
		if uri != "skills://reg-test/SKILL.md" && uri != "skills://reg-test/script.py" {
			t.Errorf("Unexpected URI: %s", uri)
		}
	})

	err = RegisterSkillResources(mockRM, manager)
	require.NoError(t, err)
}
