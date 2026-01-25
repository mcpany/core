// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"context"

	pb "github.com/mcpany/core/proto/api/v1"
	config_v1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/skill"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// SkillServiceServer provides the gRPC implementation for the SkillService.
// It handles requests for creating, listing, updating, and deleting skills by delegating to the Skill Manager.
type SkillServiceServer struct {
	pb.UnimplementedSkillServiceServer
	manager *skill.Manager
}

// NewSkillServiceServer creates a new instance of SkillServiceServer.
// It wraps the provided skill manager to handle gRPC requests for skills.
//
// Parameters:
//   - manager: The skill manager instance responsible for the business logic of skill operations.
//
// Returns:
//   - A pointer to the initialized SkillServiceServer.
func NewSkillServiceServer(manager *skill.Manager) *SkillServiceServer {
	return &SkillServiceServer{
		manager: manager,
	}
}

// ListSkills retrieves a list of all available skills.
//
// Parameters:
//   - ctx: The context for the RPC call.
//   - req: The request object (empty for this method).
//
// Returns:
//   - A response containing the list of skills.
//   - An error if the operation fails.
func (s *SkillServiceServer) ListSkills(_ context.Context, _ *pb.ListSkillsRequest) (*pb.ListSkillsResponse, error) {
	skills, err := s.manager.ListSkills()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list skills: %v", err)
	}

	pbSkills := make([]*config_v1.Skill, len(skills))
	for i, sk := range skills {
		pbSkills[i] = toProtoSkill(sk)
	}

	return &pb.ListSkillsResponse{
		Skills: pbSkills,
	}, nil
}

// GetSkill retrieves the details of a specific skill identified by its name.
//
// Parameters:
//   - ctx: The context for the RPC call.
//   - req: The request object containing the name of the skill to retrieve.
//
// Returns:
//   - A response containing the skill details.
//   - An error if the skill is not found or the operation fails.
func (s *SkillServiceServer) GetSkill(_ context.Context, req *pb.GetSkillRequest) (*pb.GetSkillResponse, error) {
	if req.Name == "" {
		return nil, status.Error(codes.InvalidArgument, "skill name is required")
	}

	sk, err := s.manager.GetSkill(req.Name)
	if err != nil {
		// convert fs errors to status codes?
		// For simplicity, just return Internal or NotFound if we check error type
		return nil, status.Errorf(codes.NotFound, "skill not found: %v", err)
	}

	return &pb.GetSkillResponse{
		Skill: toProtoSkill(sk),
	}, nil
}

// CreateSkill registers a new skill with the system.
//
// Parameters:
//   - ctx: The context for the RPC call.
//   - req: The request object containing the definition of the skill to create.
//
// Returns:
//   - A response containing the created skill.
//   - An error if the creation fails.
func (s *SkillServiceServer) CreateSkill(_ context.Context, req *pb.CreateSkillRequest) (*pb.CreateSkillResponse, error) {
	if req.Skill == nil {
		return nil, status.Error(codes.InvalidArgument, "skill is required")
	}

	sk := fromProtoSkill(req.Skill)
	if err := s.manager.CreateSkill(sk); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create skill: %v", err)
	}

	return &pb.CreateSkillResponse{
		Skill: toProtoSkill(sk),
	}, nil
}

// UpdateSkill updates the definition of an existing skill.
//
// Parameters:
//   - ctx: The context for the RPC call.
//   - req: The request object containing the name of the skill and the new definition.
//
// Returns:
//   - A response containing the updated skill.
//   - An error if the skill does not exist or the update fails.
func (s *SkillServiceServer) UpdateSkill(_ context.Context, req *pb.UpdateSkillRequest) (*pb.UpdateSkillResponse, error) {
	if req.Name == "" {
		return nil, status.Error(codes.InvalidArgument, "skill name is required")
	}
	if req.Skill == nil {
		return nil, status.Error(codes.InvalidArgument, "skill content is required")
	}

	sk := fromProtoSkill(req.Skill)
	// Ensure name matches param? Rest convention matches path.
	// Manager UpdateSkill takes oldName and newSkill.
	if err := s.manager.UpdateSkill(req.Name, sk); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update skill: %v", err)
	}

	return &pb.UpdateSkillResponse{
		Skill: toProtoSkill(sk),
	}, nil
}

// DeleteSkill removes a skill from the system.
//
// Parameters:
//   - ctx: The context for the RPC call.
//   - req: The request object containing the name of the skill to delete.
//
// Returns:
//   - An empty response on success.
//   - An error if the deletion fails.
func (s *SkillServiceServer) DeleteSkill(_ context.Context, req *pb.DeleteSkillRequest) (*pb.DeleteSkillResponse, error) {
	if req.Name == "" {
		return nil, status.Error(codes.InvalidArgument, "skill name is required")
	}

	if err := s.manager.DeleteSkill(req.Name); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to delete skill: %v", err)
	}

	return &pb.DeleteSkillResponse{}, nil
}

// Helper functions

func strPtr(s string) *string {
	return &s
}

func toProtoSkill(sk *skill.Skill) *config_v1.Skill {
	return &config_v1.Skill{
		Name:         strPtr(sk.Name),
		Description:  strPtr(sk.Description),
		License:      strPtr(sk.License),
		Instructions: strPtr(sk.Instructions),
		AllowedTools: sk.AllowedTools,
		Assets:       sk.Assets,
		Metadata:     sk.Metadata,
	}
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
