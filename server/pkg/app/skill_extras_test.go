// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"context"
	"testing"

	v1 "github.com/mcpany/core/proto/api/v1"
	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/skill"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func TestSkillCRUD_Coverage(t *testing.T) {
	// app.SkillManager is initialized in NewApplication
	// But ListSkills etc are in api_skill_grpc.go which is part of *SkillServiceServer (unexported struct? No, NewSkillServiceServer returns v1.SkillServiceServer)

	// We need to instantiate SkillServiceServer.
	// It relies on *Application? No, it relies on *skill.Manager.

	app := NewApplication()
	// app.SkillManager uses "skills" dir. We might want to mock it or use tmp dir?
	// NewManager("skills") might fail if dir doesn't exist?
	// Actually NewApplication initialized it.

	// Let's replace SkillManager with one using temp dir
	tmpDir := t.TempDir()
	skillManager, err := skill.NewManager(tmpDir)
	require.NoError(t, err)
	app.SkillManager = skillManager

	// Create server
	// NewSkillServiceServer is exported in api_skill_grpc.go?
	// `func NewSkillServiceServer(manager *skill.Manager) v1.SkillServiceServer`
	// Wait, checking api_skill_grpc.go content (I haven't fully read it, just coverage said it exists).
	// If it's in package app, I can call `NewSkillServiceServer`.

	srv := NewSkillServiceServer(app.SkillManager)

	ctx := context.Background()

	// Create
	req := &v1.CreateSkillRequest{
		Skill: &configv1.Skill{
			Name:        proto.String("test-skill"),
			Description: proto.String("desc"),
			// Parameters removed as it doesn't exist in configv1.Skill
		},
	}
	resp, err := srv.CreateSkill(ctx, req)
	require.NoError(t, err)
	assert.Equal(t, "test-skill", resp.Skill.GetName())

	// Get
	getResp, err := srv.GetSkill(ctx, &v1.GetSkillRequest{Name: "test-skill"})
	require.NoError(t, err)
	assert.Equal(t, "test-skill", getResp.Skill.GetName())

	// List
	listResp, err := srv.ListSkills(ctx, &v1.ListSkillsRequest{})
	require.NoError(t, err)
	assert.Len(t, listResp.Skills, 1)

	// Update
	// UpdateSkill might require full object?
	updateReq := &v1.UpdateSkillRequest{
		Name: "test-skill",
		Skill: &configv1.Skill{
			Name:        proto.String("test-skill"),
			Description: proto.String("updated-desc"),
		},
	}
	updateResp, err := srv.UpdateSkill(ctx, updateReq)
	require.NoError(t, err)
	assert.Equal(t, "updated-desc", updateResp.Skill.GetDescription())

	// Delete
	_, err = srv.DeleteSkill(ctx, &v1.DeleteSkillRequest{Name: "test-skill"})
	require.NoError(t, err)

	// Verify delete
	_, err = srv.GetSkill(ctx, &v1.GetSkillRequest{Name: "test-skill"})
	assert.Error(t, err)
}
