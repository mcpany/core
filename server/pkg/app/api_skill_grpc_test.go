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
	// Setup
	rootDir := t.TempDir()
	manager, err := skill.NewManager(rootDir)
	require.NoError(t, err)

	server := NewSkillServiceServer(manager)
	ctx := context.Background()

	// 1. CreateSkill
	t.Run("CreateSkill", func(t *testing.T) {
		req := &pb.CreateSkillRequest{
			Skill: &config_v1.Skill{
				Name:         strPtr("test-skill"),
				Description:  strPtr("A test skill"),
				Instructions: strPtr("Do something useful"),
			},
		}
		resp, err := server.CreateSkill(ctx, req)
		require.NoError(t, err)
		assert.NotNil(t, resp.Skill)
		assert.Equal(t, "test-skill", resp.Skill.GetName())
		assert.Equal(t, "A test skill", resp.Skill.GetDescription())

		// Verify file existence
		_, err = os.Stat(filepath.Join(rootDir, "test-skill", "SKILL.md"))
		assert.NoError(t, err)
	})

	// 2. GetSkill
	t.Run("GetSkill", func(t *testing.T) {
		req := &pb.GetSkillRequest{
			Name: "test-skill",
		}
		resp, err := server.GetSkill(ctx, req)
		require.NoError(t, err)
		assert.NotNil(t, resp.Skill)
		assert.Equal(t, "test-skill", resp.Skill.GetName())
	})

	// 3. ListSkills
	t.Run("ListSkills", func(t *testing.T) {
		req := &pb.ListSkillsRequest{}
		resp, err := server.ListSkills(ctx, req)
		require.NoError(t, err)
		assert.Len(t, resp.Skills, 1)
		assert.Equal(t, "test-skill", resp.Skills[0].GetName())
	})

	// 4. UpdateSkill
	t.Run("UpdateSkill", func(t *testing.T) {
		req := &pb.UpdateSkillRequest{
			Name: "test-skill",
			Skill: &config_v1.Skill{
				Name:         strPtr("test-skill-updated"), // Rename
				Description:  strPtr("Updated description"),
				Instructions: strPtr("Do something else"),
			},
		}
		resp, err := server.UpdateSkill(ctx, req)
		require.NoError(t, err)
		assert.NotNil(t, resp.Skill)
		assert.Equal(t, "test-skill-updated", resp.Skill.GetName())
		assert.Equal(t, "Updated description", resp.Skill.GetDescription())

		// Verify old directory gone, new exists
		_, err = os.Stat(filepath.Join(rootDir, "test-skill"))
		assert.True(t, os.IsNotExist(err))
		_, err = os.Stat(filepath.Join(rootDir, "test-skill-updated", "SKILL.md"))
		assert.NoError(t, err)
	})

	// 5. DeleteSkill
	t.Run("DeleteSkill", func(t *testing.T) {
		req := &pb.DeleteSkillRequest{
			Name: "test-skill-updated",
		}
		resp, err := server.DeleteSkill(ctx, req)
		require.NoError(t, err)
		assert.NotNil(t, resp)

		// Verify directory gone
		_, err = os.Stat(filepath.Join(rootDir, "test-skill-updated"))
		assert.True(t, os.IsNotExist(err))
	})
}

func TestSkillServiceServer_Errors(t *testing.T) {
	rootDir := t.TempDir()
	manager, _ := skill.NewManager(rootDir)
	server := NewSkillServiceServer(manager)
	ctx := context.Background()

	t.Run("CreateSkill Invalid Request", func(t *testing.T) {
		_, err := server.CreateSkill(ctx, &pb.CreateSkillRequest{Skill: nil})
		assert.Error(t, err)
		st, _ := status.FromError(err)
		assert.Equal(t, codes.InvalidArgument, st.Code())
	})

	t.Run("GetSkill Missing Name", func(t *testing.T) {
		_, err := server.GetSkill(ctx, &pb.GetSkillRequest{Name: ""})
		assert.Error(t, err)
		st, _ := status.FromError(err)
		assert.Equal(t, codes.InvalidArgument, st.Code())
	})

	t.Run("GetSkill Not Found", func(t *testing.T) {
		_, err := server.GetSkill(ctx, &pb.GetSkillRequest{Name: "non-existent"})
		assert.Error(t, err)
		st, _ := status.FromError(err)
		assert.Equal(t, codes.NotFound, st.Code())
	})

	t.Run("UpdateSkill Missing Name", func(t *testing.T) {
		_, err := server.UpdateSkill(ctx, &pb.UpdateSkillRequest{Name: "", Skill: &config_v1.Skill{}})
		assert.Error(t, err)
		st, _ := status.FromError(err)
		assert.Equal(t, codes.InvalidArgument, st.Code())
	})

	t.Run("UpdateSkill Missing Skill", func(t *testing.T) {
		_, err := server.UpdateSkill(ctx, &pb.UpdateSkillRequest{Name: "foo", Skill: nil})
		assert.Error(t, err)
		st, _ := status.FromError(err)
		assert.Equal(t, codes.InvalidArgument, st.Code())
	})

	t.Run("DeleteSkill Missing Name", func(t *testing.T) {
		_, err := server.DeleteSkill(ctx, &pb.DeleteSkillRequest{Name: ""})
		assert.Error(t, err)
		st, _ := status.FromError(err)
		assert.Equal(t, codes.InvalidArgument, st.Code())
	})
}

func TestHelpers(t *testing.T) {
	// Cover strPtr explicitly if needed, though used in other tests
	s := "test"
	p := strPtr(s)
	assert.Equal(t, s, *p)

	// toProtoSkill and fromProtoSkill are covered by Create/Update tests implicitly
}
