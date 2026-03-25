// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"context"

	pb "github.com/mcpany/core/proto/api/v1"
// NewSkillServiceServer creates a new SkillServiceServer.
//
// Summary: Initializes a new gRPC server for Skill management.
// ListSkills lists all available skills.
//
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
// Summary: Retrieves a list of all skills managed by the server.
//
// Parameters:
//   - ctx: context.Context. The request context.
// Returns:
//   - execution result or state changes.
// Errors:
//   - triggers relevant error states on failure.
// Side Effects:
//   - updates relevant subsystem state or network conditions.
//   - req: *pb.ListSkillsRequest. The request object (currently empty).
//
// Returns:
//   - *pb.ListSkillsResponse: The response containing the list of skills.
//   - error: An error if the operation fails.
// Errors:
//   - triggers relevant error states on failure.
// Side Effects:
//   - updates relevant subsystem state or network conditions.
func (s *SkillServiceServer) ListSkills(_ context.Context, _ *pb.ListSkillsRequest) (*pb.ListSkillsResponse, error) {
	skills, err := s.manager.ListSkills()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list skills: %v", err)
// GetSkill retrieves a specific skill by name.
//
// Summary: Retrieves details of a specific skill.
//
// Parameters:
//   - ctx: context.Context. The request context.
//   - req: *pb.GetSkillRequest. The request containing the skill name.
//
// Returns:
//   - *pb.GetSkillResponse: The response containing the skill details.
//   - error: An error if the skill is not found or the operation fails.
// Errors:
//   - triggers relevant error states on failure.
// Side Effects:
//   - updates relevant subsystem state or network conditions.
func (s *SkillServiceServer) GetSkill(_ context.Context, req *pb.GetSkillRequest) (*pb.GetSkillResponse, error) {
	if req.GetName() == "" {
		return nil, status.Error(codes.InvalidArgument, "skill name is required")
	}

// CreateSkill creates a new skill.
//
// Summary: Creates a new skill from the provided definition.
//
// Parameters:
//   - ctx: context.Context. The request context.
//   - req: *pb.CreateSkillRequest. The request containing the new skill definition.
//
// Returns:
//   - *pb.CreateSkillResponse: The response containing the created skill.
//   - error: An error if the operation fails (e.g., validation error, storage error).
// Errors:
//   - triggers relevant error states on failure.
// Side Effects:
//   - updates relevant subsystem state or network conditions.
func (s *SkillServiceServer) CreateSkill(_ context.Context, req *pb.CreateSkillRequest) (*pb.CreateSkillResponse, error) {
	if req.GetSkill() == nil {
		return nil, status.Error(codes.InvalidArgument, "skill is required")
// UpdateSkill updates an existing skill.
//
// Summary: Updates an existing skill definition.
//
// Parameters:
//   - ctx: context.Context. The request context.
//   - req: *pb.UpdateSkillRequest. The request containing the skill name and new definition.
//
// Returns:
//   - *pb.UpdateSkillResponse: The response containing the updated skill.
//   - error: An error if the skill is not found or update fails.
// Errors:
//   - triggers relevant error states on failure.
// Side Effects:
//   - updates relevant subsystem state or network conditions.
func (s *SkillServiceServer) UpdateSkill(_ context.Context, req *pb.UpdateSkillRequest) (*pb.UpdateSkillResponse, error) {
	if req.GetName() == "" {
		return nil, status.Error(codes.InvalidArgument, "skill name is required")
	}
	if req.GetSkill() == nil {
		return nil, status.Error(codes.InvalidArgument, "skill content is required")
	}

// DeleteSkill deletes a skill.
//
// Summary: Deletes a skill by name.
//
// Parameters:
//   - ctx: context.Context. The request context.
//   - req: *pb.DeleteSkillRequest. The request containing the name of the skill to delete.
//
// Returns:
//   - *pb.DeleteSkillResponse: An empty response on success.
//   - error: An error if the operation fails.
// Errors:
//   - triggers relevant error states on failure.
// Side Effects:
//   - updates relevant subsystem state or network conditions.
func (s *SkillServiceServer) DeleteSkill(_ context.Context, req *pb.DeleteSkillRequest) (*pb.DeleteSkillResponse, error) {
	if req.GetName() == "" {
		return nil, status.Error(codes.InvalidArgument, "skill name is required")
	}

	if err := s.manager.DeleteSkill(req.GetName()); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to delete skill: %v", err)
	}

	return &pb.DeleteSkillResponse{}, nil
}

// Helper functions

func toProtoSkill(sk *skill.Skill) *config_v1.Skill {
	return config_v1.Skill_builder{
		Name:         proto.String(sk.Name),
		Description:  proto.String(sk.Description),
		License:      proto.String(sk.License),
		Instructions: proto.String(sk.Instructions),
		AllowedTools: sk.AllowedTools,
		Assets:       sk.Assets,
		Metadata:     sk.Metadata,
	}.Build()
}

func fromProtoSkill(pbSkill *config_v1.Skill) *skill.Skill {
	return &skill.Skill{
		Frontmatter: skill.Frontmatter{
			Name:         pbSkill.GetName(),
			Description:  pbSkill.GetDescription(),
			License:      pbSkill.GetLicense(),
			AllowedTools: pbSkill.GetAllowedTools(),
			Metadata:     pbSkill.GetMetadata(),
		},
		Instructions: pbSkill.GetInstructions(),
		Assets:       pbSkill.GetAssets(),
	}
}
