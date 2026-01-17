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

func TestSkillServiceServer(t *testing.T) {
	// Setup temporary directory for skills
	tempDir, err := os.MkdirTemp("", "skill_test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create skill manager
	manager, err := skill.NewManager(tempDir)
	require.NoError(t, err)

	// Create server
	server := NewSkillServiceServer(manager)

	ctx := context.Background()

	t.Run("CreateSkill", func(t *testing.T) {
		req := &pb.CreateSkillRequest{
			Skill: &config_v1.Skill{
				Name:         strPtr("test-skill"),
				Description:  strPtr("A test skill"),
				Instructions: strPtr("Do something"),
			},
		}

		resp, err := server.CreateSkill(ctx, req)
		require.NoError(t, err)
		assert.NotNil(t, resp.Skill)
		assert.Equal(t, "test-skill", resp.Skill.GetName())
		assert.Equal(t, "A test skill", resp.Skill.GetDescription())
		assert.Equal(t, "Do something", resp.Skill.GetInstructions())

		// Verify file exists
		skillPath := filepath.Join(tempDir, "test-skill", "SKILL.md")
		assert.FileExists(t, skillPath)
	})

	t.Run("CreateSkill_Invalid", func(t *testing.T) {
		req := &pb.CreateSkillRequest{
			Skill: nil,
		}
		_, err := server.CreateSkill(ctx, req)
		assert.Error(t, err)
		assert.Equal(t, codes.InvalidArgument, status.Code(err))
	})

	t.Run("GetSkill", func(t *testing.T) {
		req := &pb.GetSkillRequest{
			Name: "test-skill",
		}
		resp, err := server.GetSkill(ctx, req)
		require.NoError(t, err)
		assert.NotNil(t, resp.Skill)
		assert.Equal(t, "test-skill", resp.Skill.GetName())
	})

	t.Run("GetSkill_NotFound", func(t *testing.T) {
		req := &pb.GetSkillRequest{
			Name: "non-existent-skill",
		}
		_, err := server.GetSkill(ctx, req)
		assert.Error(t, err)
		assert.Equal(t, codes.NotFound, status.Code(err))
	})

	t.Run("GetSkill_Invalid", func(t *testing.T) {
		req := &pb.GetSkillRequest{
			Name: "",
		}
		_, err := server.GetSkill(ctx, req)
		assert.Error(t, err)
		assert.Equal(t, codes.InvalidArgument, status.Code(err))
	})

	t.Run("ListSkills", func(t *testing.T) {
		req := &pb.ListSkillsRequest{}
		resp, err := server.ListSkills(ctx, req)
		require.NoError(t, err)
		assert.NotEmpty(t, resp.Skills)
		found := false
		for _, s := range resp.Skills {
			if s.GetName() == "test-skill" {
				found = true
				break
			}
		}
		assert.True(t, found)
	})

	t.Run("UpdateSkill", func(t *testing.T) {
		req := &pb.UpdateSkillRequest{
			Name: "test-skill",
			Skill: &config_v1.Skill{
				Name:         strPtr("updated-skill"), // Rename
				Description:  strPtr("Updated description"),
				Instructions: strPtr("Do something updated"),
			},
		}
		resp, err := server.UpdateSkill(ctx, req)
		require.NoError(t, err)
		assert.Equal(t, "updated-skill", resp.Skill.GetName())
		assert.Equal(t, "Updated description", resp.Skill.GetDescription())

		// Verify old skill is gone and new one exists
		assert.NoFileExists(t, filepath.Join(tempDir, "test-skill", "SKILL.md"))
		assert.FileExists(t, filepath.Join(tempDir, "updated-skill", "SKILL.md"))
	})

	t.Run("UpdateSkill_Invalid", func(t *testing.T) {
		req := &pb.UpdateSkillRequest{
			Name: "",
		}
		_, err := server.UpdateSkill(ctx, req)
		assert.Error(t, err)
		assert.Equal(t, codes.InvalidArgument, status.Code(err))
	})

	t.Run("UpdateSkill_MissingContent", func(t *testing.T) {
		req := &pb.UpdateSkillRequest{
			Name: "updated-skill",
			Skill: nil,
		}
		_, err := server.UpdateSkill(ctx, req)
		assert.Error(t, err)
		assert.Equal(t, codes.InvalidArgument, status.Code(err))
	})

	t.Run("DeleteSkill", func(t *testing.T) {
		req := &pb.DeleteSkillRequest{
			Name: "updated-skill",
		}
		resp, err := server.DeleteSkill(ctx, req)
		require.NoError(t, err)
		assert.NotNil(t, resp)

		// Verify file is gone
		assert.NoFileExists(t, filepath.Join(tempDir, "updated-skill", "SKILL.md"))
	})

	t.Run("DeleteSkill_NotFound", func(t *testing.T) {
		req := &pb.DeleteSkillRequest{
			Name: "non-existent-skill",
		}
		_, err := server.DeleteSkill(ctx, req)
		assert.Error(t, err)
		assert.Equal(t, codes.Internal, status.Code(err)) // Manager returns generic error, wrapped as Internal
	})

	t.Run("DeleteSkill_Invalid", func(t *testing.T) {
		req := &pb.DeleteSkillRequest{
			Name: "",
		}
		_, err := server.DeleteSkill(ctx, req)
		assert.Error(t, err)
		assert.Equal(t, codes.InvalidArgument, status.Code(err))
	})
}
