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
	"google.golang.org/protobuf/proto"
)

// SkillServiceServer implements the SkillService gRPC interface.
//
// Summary: Handles gRPC requests for managing skills (CRUD operations).
//
// Fields:
//   - UnimplementedSkillServiceServer: Embeds the base gRPC server implementation.
//   - manager: The skill manager instance for handling business logic.
type SkillServiceServer struct {
	pb.UnimplementedSkillServiceServer
	manager *skill.Manager
}

// NewSkillServiceServer creates a new SkillServiceServer.
//
// Summary: Initializes a new gRPC handler for the Skill Service.
//
// Parameters:
//   - manager: *skill.Manager. The underlying manager for skill data persistence.
//
// Returns:
//   - *SkillServiceServer: The initialized gRPC server handler.
func NewSkillServiceServer(manager *skill.Manager) *SkillServiceServer {
	return &SkillServiceServer{
		manager: manager,
	}
}

// ListSkills lists all available skills.
//
// Summary: Retrieves a list of all registered skills.
//
// Parameters:
//   - _ : context.Context. Request context (unused).
//   - _ : *pb.ListSkillsRequest. List request parameters (unused).
//
// Returns:
//   - *pb.ListSkillsResponse: The response containing the list of skills.
//   - error: An error if the retrieval fails.
func (s *SkillServiceServer) ListSkills(_ context.Context, _ *pb.ListSkillsRequest) (*pb.ListSkillsResponse, error) {
	skills, err := s.manager.ListSkills()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list skills: %v", err)
	}

	pbSkills := make([]*config_v1.Skill, len(skills))
	for i, sk := range skills {
		pbSkills[i] = toProtoSkill(sk)
	}

	return pb.ListSkillsResponse_builder{
		Skills: pbSkills,
	}.Build(), nil
}

// GetSkill retrieves a specific skill by name.
//
// Summary: Fetches the details of a skill identified by its name.
//
// Parameters:
//   - _ : context.Context. Request context.
//   - req: *pb.GetSkillRequest. The request containing the skill name.
//
// Returns:
//   - *pb.GetSkillResponse: The response containing the skill details.
//   - error: An error if the skill is not found or retrieval fails.
func (s *SkillServiceServer) GetSkill(_ context.Context, req *pb.GetSkillRequest) (*pb.GetSkillResponse, error) {
	if req.GetName() == "" {
		return nil, status.Error(codes.InvalidArgument, "skill name is required")
	}

	sk, err := s.manager.GetSkill(req.GetName())
	if err != nil {
		// convert fs errors to status codes?
		// For simplicity, just return Internal or NotFound if we check error type
		return nil, status.Errorf(codes.NotFound, "skill not found: %v", err)
	}

	return pb.GetSkillResponse_builder{
		Skill: toProtoSkill(sk),
	}.Build(), nil
}

// CreateSkill creates a new skill.
//
// Summary: Stores a new skill definition.
//
// Parameters:
//   - _ : context.Context. Request context.
//   - req: *pb.CreateSkillRequest. The request containing the skill definition to create.
//
// Returns:
//   - *pb.CreateSkillResponse: The response containing the created skill.
//   - error: An error if validation or creation fails.
func (s *SkillServiceServer) CreateSkill(_ context.Context, req *pb.CreateSkillRequest) (*pb.CreateSkillResponse, error) {
	if req.GetSkill() == nil {
		return nil, status.Error(codes.InvalidArgument, "skill is required")
	}

	sk := fromProtoSkill(req.GetSkill())
	if err := s.manager.CreateSkill(sk); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create skill: %v", err)
	}

	return pb.CreateSkillResponse_builder{
		Skill: toProtoSkill(sk),
	}.Build(), nil
}

// UpdateSkill updates an existing skill.
//
// Summary: Modifies an existing skill definition.
//
// Parameters:
//   - _ : context.Context. Request context.
//   - req: *pb.UpdateSkillRequest. The request containing the skill name and updated definition.
//
// Returns:
//   - *pb.UpdateSkillResponse: The response containing the updated skill.
//   - error: An error if the skill is not found or update fails.
func (s *SkillServiceServer) UpdateSkill(_ context.Context, req *pb.UpdateSkillRequest) (*pb.UpdateSkillResponse, error) {
	if req.GetName() == "" {
		return nil, status.Error(codes.InvalidArgument, "skill name is required")
	}
	if req.GetSkill() == nil {
		return nil, status.Error(codes.InvalidArgument, "skill content is required")
	}

	sk := fromProtoSkill(req.GetSkill())
	// Ensure name matches param? Rest convention matches path.
	// Manager UpdateSkill takes oldName and newSkill.
	if err := s.manager.UpdateSkill(req.GetName(), sk); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update skill: %v", err)
	}

	return pb.UpdateSkillResponse_builder{
		Skill: toProtoSkill(sk),
	}.Build(), nil
}

// DeleteSkill deletes a skill.
//
// Summary: Removes a skill definition.
//
// Parameters:
//   - _ : context.Context. Request context.
//   - req: *pb.DeleteSkillRequest. The request containing the name of the skill to delete.
//
// Returns:
//   - *pb.DeleteSkillResponse: Empty response on success.
//   - error: An error if the deletion fails.
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
