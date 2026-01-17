// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	pb "github.com/mcpany/core/proto/api/v1"
	config_v1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/skill"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func setupTestSkillManager(t *testing.T) (*skill.Manager, string) {
	tempDir, err := os.MkdirTemp("", "skill_test")
	require.NoError(t, err)

	manager, err := skill.NewManager(tempDir)
	require.NoError(t, err)

	t.Cleanup(func() {
		os.RemoveAll(tempDir)
	})

	return manager, tempDir
}

func TestSkillServiceServer_CreateSkill(t *testing.T) {
	manager, _ := setupTestSkillManager(t)
	server := NewSkillServiceServer(manager)

	ctx := context.Background()

	t.Run("Create valid skill", func(t *testing.T) {
		req := &pb.CreateSkillRequest{
			Skill: &config_v1.Skill{
				Name:         ptr("test-skill"),
				Description:  ptr("Test Skill Description"),
				Instructions: ptr("Do something"),
			},
		}

		resp, err := server.CreateSkill(ctx, req)
		require.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, "test-skill", resp.Skill.GetName())
		assert.Equal(t, "Test Skill Description", resp.Skill.GetDescription())

		// Verify it exists in manager
		sk, err := manager.GetSkill("test-skill")
		require.NoError(t, err)
		assert.Equal(t, "test-skill", sk.Name)
	})

	t.Run("Create skill with missing input", func(t *testing.T) {
		req := &pb.CreateSkillRequest{
			Skill: nil,
		}
		_, err := server.CreateSkill(ctx, req)
		assert.Error(t, err)
		assert.Equal(t, codes.InvalidArgument, status.Code(err))
	})

	t.Run("Create duplicate skill", func(t *testing.T) {
		req := &pb.CreateSkillRequest{
			Skill: &config_v1.Skill{
				Name:         ptr("test-skill"), // Already created
				Description:  ptr("Duplicate"),
				Instructions: ptr("Do something"),
			},
		}
		_, err := server.CreateSkill(ctx, req)
		assert.Error(t, err)
		assert.Equal(t, codes.Internal, status.Code(err))
	})
}

func TestSkillServiceServer_GetSkill(t *testing.T) {
	manager, _ := setupTestSkillManager(t)
	server := NewSkillServiceServer(manager)
	ctx := context.Background()

	// Setup initial skill
	err := manager.CreateSkill(&skill.Skill{
		Frontmatter: skill.Frontmatter{Name: "existing-skill"},
		Instructions: "instruct",
	})
	require.NoError(t, err)

	t.Run("Get existing skill", func(t *testing.T) {
		req := &pb.GetSkillRequest{Name: "existing-skill"}
		resp, err := server.GetSkill(ctx, req)
		require.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, "existing-skill", resp.Skill.GetName())
	})

	t.Run("Get non-existent skill", func(t *testing.T) {
		req := &pb.GetSkillRequest{Name: "non-existent"}
		_, err := server.GetSkill(ctx, req)
		assert.Error(t, err)
		assert.Equal(t, codes.NotFound, status.Code(err))
	})

	t.Run("Get skill with empty name", func(t *testing.T) {
		req := &pb.GetSkillRequest{Name: ""}
		_, err := server.GetSkill(ctx, req)
		assert.Error(t, err)
		assert.Equal(t, codes.InvalidArgument, status.Code(err))
	})
}

func TestSkillServiceServer_ListSkills(t *testing.T) {
	manager, _ := setupTestSkillManager(t)
	server := NewSkillServiceServer(manager)
	ctx := context.Background()

	// Create a few skills
	skills := []string{"skill-1", "skill-2", "skill-3"}
	for _, name := range skills {
		err := manager.CreateSkill(&skill.Skill{
			Frontmatter: skill.Frontmatter{Name: name},
			Instructions: "instruct",
		})
		require.NoError(t, err)
	}

	t.Run("List skills", func(t *testing.T) {
		req := &pb.ListSkillsRequest{}
		resp, err := server.ListSkills(ctx, req)
		require.NoError(t, err)
		assert.Len(t, resp.Skills, 3)

		// Order is not guaranteed by file system listing usually, but manager implementation might.
		// Let's just check existence
		names := make([]string, 0, 3)
		for _, s := range resp.Skills {
			names = append(names, s.GetName())
		}
		assert.ElementsMatch(t, skills, names)
	})
}

func TestSkillServiceServer_UpdateSkill(t *testing.T) {
	manager, _ := setupTestSkillManager(t)
	server := NewSkillServiceServer(manager)
	ctx := context.Background()

	// Create initial skill
	err := manager.CreateSkill(&skill.Skill{
		Frontmatter: skill.Frontmatter{Name: "to-update"},
		Instructions: "v1",
	})
	require.NoError(t, err)

	t.Run("Update existing skill", func(t *testing.T) {
		req := &pb.UpdateSkillRequest{
			Name: "to-update",
			Skill: &config_v1.Skill{
				Name:         ptr("to-update"),
				Description:  ptr("Updated Description"),
				Instructions: ptr("v2"),
			},
		}
		resp, err := server.UpdateSkill(ctx, req)
		require.NoError(t, err)
		assert.Equal(t, "Updated Description", resp.Skill.GetDescription())

		sk, err := manager.GetSkill("to-update")
		require.NoError(t, err)
		assert.Equal(t, "v2", sk.Instructions)
	})

	t.Run("Rename skill", func(t *testing.T) {
		req := &pb.UpdateSkillRequest{
			Name: "to-update",
			Skill: &config_v1.Skill{
				Name:         ptr("renamed-skill"),
				Instructions: ptr("v2"),
			},
		}
		resp, err := server.UpdateSkill(ctx, req)
		require.NoError(t, err)
		assert.Equal(t, "renamed-skill", resp.Skill.GetName())

		_, err = manager.GetSkill("to-update")
		assert.Error(t, err) // Old name should be gone

		sk, err := manager.GetSkill("renamed-skill")
		require.NoError(t, err)
		assert.Equal(t, "renamed-skill", sk.Name)
	})

	t.Run("Update non-existent skill", func(t *testing.T) {
		req := &pb.UpdateSkillRequest{
			Name: "fake",
			Skill: &config_v1.Skill{
				Name: ptr("fake"),
			},
		}
		_, err := server.UpdateSkill(ctx, req)
		assert.Error(t, err)
		assert.Equal(t, codes.Internal, status.Code(err))
	})

	t.Run("Update with invalid args", func(t *testing.T) {
		_, err := server.UpdateSkill(ctx, &pb.UpdateSkillRequest{Name: ""})
		assert.Error(t, err)
		assert.Equal(t, codes.InvalidArgument, status.Code(err))

		_, err = server.UpdateSkill(ctx, &pb.UpdateSkillRequest{Name: "valid", Skill: nil})
		assert.Error(t, err)
		assert.Equal(t, codes.InvalidArgument, status.Code(err))
	})
}

func TestSkillServiceServer_DeleteSkill(t *testing.T) {
	manager, tempDir := setupTestSkillManager(t)
	server := NewSkillServiceServer(manager)
	ctx := context.Background()

	err := manager.CreateSkill(&skill.Skill{
		Frontmatter: skill.Frontmatter{Name: "to-delete"},
		Instructions: "bye",
	})
	require.NoError(t, err)

	t.Run("Delete existing skill", func(t *testing.T) {
		req := &pb.DeleteSkillRequest{Name: "to-delete"}
		_, err := server.DeleteSkill(ctx, req)
		require.NoError(t, err)

		_, err = manager.GetSkill("to-delete")
		assert.Error(t, err)

		// Verify dir is gone
		_, err = os.Stat(filepath.Join(tempDir, "to-delete"))
		assert.True(t, os.IsNotExist(err))
	})

	t.Run("Delete non-existent skill", func(t *testing.T) {
		req := &pb.DeleteSkillRequest{Name: "ghost"}
		_, err := server.DeleteSkill(ctx, req)
		assert.Error(t, err)
		assert.Equal(t, codes.Internal, status.Code(err))
	})

	t.Run("Delete with empty name", func(t *testing.T) {
		req := &pb.DeleteSkillRequest{Name: ""}
		_, err := server.DeleteSkill(ctx, req)
		assert.Error(t, err)
		assert.Equal(t, codes.InvalidArgument, status.Code(err))
	})
}
