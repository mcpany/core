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

func strPtrTest(s string) *string {
	return &s
}

func TestSkillServiceServer_CreateSkill(t *testing.T) {
	manager, _ := setupTestManager(t)

	server := NewSkillServiceServer(manager)

	ctx := context.Background()

	t.Run("Create valid skill", func(t *testing.T) {
		req := &pb.CreateSkillRequest{
			Skill: &config_v1.Skill{
				Name:         strPtrTest("test-skill"),
				Description:  strPtrTest("A test skill"),
				Instructions: strPtrTest("Do the test"),
			},
		}

		resp, err := server.CreateSkill(ctx, req)
		require.NoError(t, err)
		assert.Equal(t, "test-skill", resp.Skill.GetName())
		assert.Equal(t, "A test skill", resp.Skill.GetDescription())
	})

	t.Run("Create skill missing input", func(t *testing.T) {
		req := &pb.CreateSkillRequest{
			Skill: nil,
		}
		_, err := server.CreateSkill(ctx, req)
		assert.Error(t, err)
		st, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.InvalidArgument, st.Code())
	})

	t.Run("Create skill with invalid name", func(t *testing.T) {
		req := &pb.CreateSkillRequest{
			Skill: &config_v1.Skill{
				Name:         strPtrTest("Invalid Name"), // Uppercase and space not allowed
				Description:  strPtrTest("Invalid"),
				Instructions: strPtrTest("Invalid"),
			},
		}
		_, err := server.CreateSkill(ctx, req)
		assert.Error(t, err)
		st, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.Internal, st.Code()) // Manager returns error, wrapped as Internal
	})

	t.Run("Create duplicate skill", func(t *testing.T) {
		req := &pb.CreateSkillRequest{
			Skill: &config_v1.Skill{
				Name:         strPtrTest("duplicate-skill"),
				Description:  strPtrTest("First"),
				Instructions: strPtrTest("First"),
			},
		}
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
		req := &pb.GetSkillRequest{Name: "get-test"}
		resp, err := server.GetSkill(ctx, req)
		require.NoError(t, err)
		assert.Equal(t, "get-test", resp.Skill.GetName())
	})

	t.Run("Get non-existent skill", func(t *testing.T) {
		req := &pb.GetSkillRequest{Name: "non-existent"}
		_, err := server.GetSkill(ctx, req)
		assert.Error(t, err)
		st, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.NotFound, st.Code())
	})

	t.Run("Get skill with empty name", func(t *testing.T) {
		req := &pb.GetSkillRequest{Name: ""}
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
	assert.Len(t, resp.Skills, 2)

	names := make(map[string]bool)
	for _, s := range resp.Skills {
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
		req := &pb.UpdateSkillRequest{
			Name: "update-me",
			Skill: &config_v1.Skill{
				Name:         strPtrTest("update-me"),
				Description:  strPtrTest("New Desc"),
				Instructions: strPtrTest("New Inst"),
			},
		}
		resp, err := server.UpdateSkill(ctx, req)
		require.NoError(t, err)
		assert.Equal(t, "New Desc", resp.Skill.GetDescription())

		// Verify persistence
		sk, err := manager.GetSkill("update-me")
		require.NoError(t, err)
		assert.Equal(t, "New Desc", sk.Description)
	})

	t.Run("Rename skill", func(t *testing.T) {
		req := &pb.UpdateSkillRequest{
			Name: "update-me",
			Skill: &config_v1.Skill{
				Name:         strPtrTest("renamed-skill"),
				Description:  strPtrTest("Renamed"),
				Instructions: strPtrTest("Renamed Inst"),
			},
		}
		resp, err := server.UpdateSkill(ctx, req)
		require.NoError(t, err)
		assert.Equal(t, "renamed-skill", resp.Skill.GetName())

		// Verify old name gone
		_, err = manager.GetSkill("update-me")
		assert.Error(t, err)

		// Verify new name exists
		sk, err := manager.GetSkill("renamed-skill")
		require.NoError(t, err)
		assert.Equal(t, "renamed-skill", sk.Name)
	})

	t.Run("Update non-existent skill", func(t *testing.T) {
		req := &pb.UpdateSkillRequest{
			Name: "missing",
			Skill: &config_v1.Skill{
				Name: strPtrTest("missing"),
				Instructions: strPtrTest("Foo"),
			},
		}
		_, err := server.UpdateSkill(ctx, req)
		assert.Error(t, err)
		st, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.Internal, st.Code()) // Manager returns "skill not found" error
	})

	t.Run("Update with invalid inputs", func(t *testing.T) {
		_, err := server.UpdateSkill(ctx, &pb.UpdateSkillRequest{Name: "", Skill: &config_v1.Skill{}})
		assert.Error(t, err)
		assert.Equal(t, codes.InvalidArgument, status.Code(err))

		_, err = server.UpdateSkill(ctx, &pb.UpdateSkillRequest{Name: "foo", Skill: nil})
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
		req := &pb.DeleteSkillRequest{Name: "delete-me"}
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
		req := &pb.DeleteSkillRequest{Name: "ghost"}
		_, err := server.DeleteSkill(ctx, req)
		assert.Error(t, err)
		st, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.Internal, st.Code())
	})

	t.Run("Delete with empty name", func(t *testing.T) {
		req := &pb.DeleteSkillRequest{Name: ""}
		_, err := server.DeleteSkill(ctx, req)
		assert.Error(t, err)
		assert.Equal(t, codes.InvalidArgument, status.Code(err))
	})
}
