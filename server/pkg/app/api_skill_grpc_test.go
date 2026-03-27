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
	"google.golang.org/protobuf/proto"
)

func setupTestManager(t *testing.T) (*skill.Manager, string) {
	tmpDir, err := os.MkdirTemp("", "skill_test")
	require.NoError(t, err)
	t.Cleanup(func() {
		os.RemoveAll(tmpDir)
	})

	manager, err := skill.NewManager(tmpDir)
	require.NoError(t, err)

	return manager, tmpDir
}


func TestSkillServiceServer_CreateSkill(t *testing.T) {
	manager, _ := setupTestManager(t)

	server := NewSkillServiceServer(manager)

	ctx := context.Background()

	t.Run("Create valid skill", func(t *testing.T) {
		req := pb.CreateSkillRequest_builder{
			Skill: config_v1.Skill_builder{
				Name:         proto.String("test-skill"),
				Description:  proto.String("A test skill"),
				Instructions: proto.String("Do the test"),
			}.Build(),
		}.Build()

		resp, err := server.CreateSkill(ctx, req)
		require.NoError(t, err)
		assert.Equal(t, "test-skill", resp.GetSkill().GetName())
		assert.Equal(t, "A test skill", resp.GetSkill().GetDescription())
	})

	t.Run("Create skill missing input", func(t *testing.T) {
		req := pb.CreateSkillRequest_builder{
			Skill: nil,
		}.Build()
		_, err := server.CreateSkill(ctx, req)
		assert.Error(t, err)
		st, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.InvalidArgument, st.Code())
	})

	t.Run("Create skill with invalid name", func(t *testing.T) {
		req := pb.CreateSkillRequest_builder{
			Skill: config_v1.Skill_builder{
				Name:         proto.String("Invalid Name"), // Uppercase and space not allowed
				Description:  proto.String("Invalid"),
				Instructions: proto.String("Invalid"),
			}.Build(),
		}.Build()
		_, err := server.CreateSkill(ctx, req)
		assert.Error(t, err)
		st, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.Internal, st.Code()) // Manager returns error, wrapped as Internal
	})

	t.Run("Create duplicate skill", func(t *testing.T) {
		req := pb.CreateSkillRequest_builder{
			Skill: config_v1.Skill_builder{
				Name:         proto.String("duplicate-skill"),
				Description:  proto.String("First"),
				Instructions: proto.String("First"),
			}.Build(),
		}.Build()
		_, err := server.CreateSkill(ctx, req)
		require.NoError(t, err)

		_, err = server.CreateSkill(ctx, req)
		assert.Error(t, err)
		st, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.Internal, st.Code())
	})
}

func TestSkillServiceServer_GetSkill(t *testing.T) {
	manager, _ := setupTestManager(t)
	server := NewSkillServiceServer(manager)
	ctx := context.Background()

	// Seed a skill
	testSkill := &skill.Skill{
		Frontmatter: skill.Frontmatter{
			Name:        "get-test",
			Description: "Get Test",
		},
		Instructions: "Instructions",
	}
	require.NoError(t, manager.CreateSkill(testSkill))

	t.Run("Get existing skill", func(t *testing.T) {
		req := pb.GetSkillRequest_builder{Name: "get-test"}.Build()
		resp, err := server.GetSkill(ctx, req)
		require.NoError(t, err)
		assert.Equal(t, "get-test", resp.GetSkill().GetName())
	})

	t.Run("Get non-existent skill", func(t *testing.T) {
		req := pb.GetSkillRequest_builder{Name: "non-existent"}.Build()
		_, err := server.GetSkill(ctx, req)
		assert.Error(t, err)
		st, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.NotFound, st.Code())
	})

	t.Run("Get skill with empty name", func(t *testing.T) {
		req := pb.GetSkillRequest_builder{Name: ""}.Build()
		_, err := server.GetSkill(ctx, req)
		assert.Error(t, err)
		st, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.InvalidArgument, st.Code())
	})
}

func TestSkillServiceServer_ListSkills(t *testing.T) {
	manager, _ := setupTestManager(t)
	server := NewSkillServiceServer(manager)
	ctx := context.Background()

	// Seed skills
	s1 := &skill.Skill{Frontmatter: skill.Frontmatter{Name: "skill-1"}, Instructions: "I1"}
	s2 := &skill.Skill{Frontmatter: skill.Frontmatter{Name: "skill-2"}, Instructions: "I2"}
	require.NoError(t, manager.CreateSkill(s1))
	require.NoError(t, manager.CreateSkill(s2))

	resp, err := server.ListSkills(ctx, &pb.ListSkillsRequest{})
	require.NoError(t, err)
	assert.Len(t, resp.GetSkills(), 2)

	names := make(map[string]bool)
	for _, s := range resp.GetSkills() {
		names[s.GetName()] = true
	}
	assert.True(t, names["skill-1"])
	assert.True(t, names["skill-2"])
}

func TestSkillServiceServer_UpdateSkill(t *testing.T) {
	manager, _ := setupTestManager(t)
	server := NewSkillServiceServer(manager)
	ctx := context.Background()

	// Seed a skill
	require.NoError(t, manager.CreateSkill(&skill.Skill{
		Frontmatter: skill.Frontmatter{Name: "update-me", Description: "Old"},
		Instructions: "Old Instructions",
	}))

	t.Run("Update existing skill", func(t *testing.T) {
		req := pb.UpdateSkillRequest_builder{
			Name: "update-me",
			Skill: config_v1.Skill_builder{
				Name:         proto.String("update-me"),
				Description:  proto.String("New Desc"),
				Instructions: proto.String("New Inst"),
			}.Build(),
		}.Build()
		resp, err := server.UpdateSkill(ctx, req)
		require.NoError(t, err)
		assert.Equal(t, "New Desc", resp.GetSkill().GetDescription())

		// Verify persistence
		sk, err := manager.GetSkill("update-me")
		require.NoError(t, err)
		assert.Equal(t, "New Desc", sk.Description)
	})

	t.Run("Rename skill", func(t *testing.T) {
		req := pb.UpdateSkillRequest_builder{
			Name: "update-me",
			Skill: config_v1.Skill_builder{
				Name:         proto.String("renamed-skill"),
				Description:  proto.String("Renamed"),
				Instructions: proto.String("Renamed Inst"),
			}.Build(),
		}.Build()
		resp, err := server.UpdateSkill(ctx, req)
		require.NoError(t, err)
		assert.Equal(t, "renamed-skill", resp.GetSkill().GetName())

		// Verify old name gone
		_, err = manager.GetSkill("update-me")
		assert.Error(t, err)

		// Verify new name exists
		sk, err := manager.GetSkill("renamed-skill")
		require.NoError(t, err)
		assert.Equal(t, "renamed-skill", sk.Name)
	})

	t.Run("Update non-existent skill", func(t *testing.T) {
		req := pb.UpdateSkillRequest_builder{
			Name: "missing",
			Skill: config_v1.Skill_builder{
				Name: proto.String("missing"),
				Instructions: proto.String("Foo"),
			}.Build(),
		}.Build()
		_, err := server.UpdateSkill(ctx, req)
		assert.Error(t, err)
		st, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.Internal, st.Code()) // Manager returns "skill not found" error
	})

	t.Run("Update with invalid inputs", func(t *testing.T) {
		_, err := server.UpdateSkill(ctx, pb.UpdateSkillRequest_builder{Name: "", Skill: config_v1.Skill_builder{}.Build()}.Build())
		assert.Error(t, err)
		assert.Equal(t, codes.InvalidArgument, status.Code(err))

		_, err = server.UpdateSkill(ctx, pb.UpdateSkillRequest_builder{Name: "foo", Skill: nil}.Build())
		assert.Error(t, err)
		assert.Equal(t, codes.InvalidArgument, status.Code(err))
	})
}

func TestSkillServiceServer_DeleteSkill(t *testing.T) {
	manager, tmpDir := setupTestManager(t)
	server := NewSkillServiceServer(manager)
	ctx := context.Background()

	require.NoError(t, manager.CreateSkill(&skill.Skill{
		Frontmatter: skill.Frontmatter{Name: "delete-me"},
		Instructions: "Bye",
	}))

	t.Run("Delete existing skill", func(t *testing.T) {
		req := pb.DeleteSkillRequest_builder{Name: "delete-me"}.Build()
		_, err := server.DeleteSkill(ctx, req)
		require.NoError(t, err)

		// Verify gone
		_, err = manager.GetSkill("delete-me")
		assert.Error(t, err)

		// Check filesystem
		_, err = os.Stat(filepath.Join(tmpDir, "delete-me"))
		assert.True(t, os.IsNotExist(err))
	})

	t.Run("Delete non-existent skill", func(t *testing.T) {
		req := pb.DeleteSkillRequest_builder{Name: "ghost"}.Build()
		_, err := server.DeleteSkill(ctx, req)
		assert.Error(t, err)
		st, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.Internal, st.Code())
	})

	t.Run("Delete with empty name", func(t *testing.T) {
		req := pb.DeleteSkillRequest_builder{Name: ""}.Build()
		_, err := server.DeleteSkill(ctx, req)
		assert.Error(t, err)
		assert.Equal(t, codes.InvalidArgument, status.Code(err))
	})
}
