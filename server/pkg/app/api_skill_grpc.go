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
type SkillServiceServer struct {
	pb.UnimplementedSkillServiceServer
	manager *skill.Manager
}

// NewSkillServiceServer creates a new SkillServiceServer.
//
// manager handles the resource management.
//
// Returns the result.
func NewSkillServiceServer(manager *skill.Manager) *SkillServiceServer {
	return &SkillServiceServer{
		manager: manager,
	}
}

// ListSkills lists all available skills.
//
// _ is an unused parameter.
// _ is an unused parameter.
//
// Returns the response.
// Returns an error if the operation fails.
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
// _ is an unused parameter.
// req is the request object.
//
// Returns the response.
// Returns an error if the operation fails.
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
// _ is an unused parameter.
// req is the request object.
//
// Returns the response.
// Returns an error if the operation fails.
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
// _ is an unused parameter.
// req is the request object.
//
// Returns the response.
// Returns an error if the operation fails.
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
// _ is an unused parameter.
// req is the request object.
//
// Returns the response.
// Returns an error if the operation fails.
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
