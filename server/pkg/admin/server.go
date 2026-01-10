// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

// Package admin implements the admin server
package admin

import (
	"context"
	"strings"

	pb "github.com/mcpany/core/proto/admin/v1"
	configv1 "github.com/mcpany/core/proto/config/v1"
	mcprouterv1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/mcpany/core/server/pkg/middleware"
	"github.com/mcpany/core/server/pkg/storage"
	"github.com/mcpany/core/server/pkg/tool"
	"github.com/mcpany/core/server/pkg/util/passhash"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)

// Server implements the AdminServiceServer interface.
type Server struct {
	pb.UnimplementedAdminServiceServer
	cache       *middleware.CachingMiddleware
	toolManager tool.ManagerInterface
	storage     storage.Storage
}

// NewServer creates a new Admin Server.
//
// Parameters:
//   cache: The caching middleware to use for clearing cache.
//   toolManager: The tool manager to use for listing services and tools.
//   storage: The storage backend to use for user management.
//
// Returns:
//   *Server: A new instance of the Admin Server.
func NewServer(
	cache *middleware.CachingMiddleware,
	toolManager tool.ManagerInterface,
	storage storage.Storage,
) *Server {
	return &Server{
		cache:       cache,
		toolManager: toolManager,
		storage:     storage,
	}
}

// ClearCache clears the cache.
//
// Parameters:
//   ctx: The context for the request.
//   _ : The request object (unused).
//
// Returns:
//   *pb.ClearCacheResponse: Empty response on success.
//   error: An error if the cache clearing failed or is not enabled.
func (s *Server) ClearCache(ctx context.Context, _ *pb.ClearCacheRequest) (*pb.ClearCacheResponse, error) {
	if s.cache == nil {
		return nil, status.Error(codes.FailedPrecondition, "caching is not enabled")
	}
	if err := s.cache.Clear(ctx); err != nil {
		return nil, err
	}
	return &pb.ClearCacheResponse{}, nil
}

// ListServices returns all registered services.
//
// Parameters:
//   _ : The context for the request (unused).
//   _ : The request object (unused).
//
// Returns:
//   *pb.ListServicesResponse: The list of registered services.
//   error: An error if the operation failed.
func (s *Server) ListServices(_ context.Context, _ *pb.ListServicesRequest) (*pb.ListServicesResponse, error) {
	serviceInfos := s.toolManager.ListServices()
	var services []*configv1.UpstreamServiceConfig
	for _, info := range serviceInfos {
		if info.Config != nil {
			services = append(services, info.Config)
		}
	}
	return &pb.ListServicesResponse{Services: services}, nil
}

// GetService returns a specific service by ID.
//
// Parameters:
//   _ : The context for the request (unused).
//   req: The request object containing the service ID.
//
// Returns:
//   *pb.GetServiceResponse: The requested service configuration.
//   error: An error if the service was not found or has no config.
func (s *Server) GetService(_ context.Context, req *pb.GetServiceRequest) (*pb.GetServiceResponse, error) {
	info, ok := s.toolManager.GetServiceInfo(req.GetServiceId())
	if !ok {
		return nil, status.Error(codes.NotFound, "service not found")
	}
	if info.Config == nil {
		return nil, status.Error(codes.Internal, "service config not found")
	}
	return &pb.GetServiceResponse{Service: info.Config}, nil
}

// ListTools returns all registered tools.
//
// Parameters:
//   _ : The context for the request (unused).
//   _ : The request object (unused).
//
// Returns:
//   *pb.ListToolsResponse: The list of all registered tools.
//   error: An error if the operation failed.
func (s *Server) ListTools(_ context.Context, _ *pb.ListToolsRequest) (*pb.ListToolsResponse, error) {
	tools := s.toolManager.ListTools()
	responseTools := make([]*mcprouterv1.Tool, 0, len(tools))
	for _, t := range tools {
		responseTools = append(responseTools, t.Tool())
	}
	return &pb.ListToolsResponse{Tools: responseTools}, nil
}

// GetTool returns a specific tool by name.
//
// Parameters:
//   _ : The context for the request (unused).
//   req: The request object containing the tool name.
//
// Returns:
//   *pb.GetToolResponse: The requested tool definition.
//   error: An error if the tool was not found.
func (s *Server) GetTool(_ context.Context, req *pb.GetToolRequest) (*pb.GetToolResponse, error) {
	t, ok := s.toolManager.GetTool(req.GetToolName())
	if !ok {
		return nil, status.Error(codes.NotFound, "tool not found")
	}
	return &pb.GetToolResponse{Tool: t.Tool()}, nil
}

// CreateUser creates a new user.
//
// Parameters:
//   ctx: The context for the request.
//   req: The request object containing the user to create.
//
// Returns:
//   *pb.CreateUserResponse: The created user.
//   error: An error if the user creation failed.
func (s *Server) CreateUser(ctx context.Context, req *pb.CreateUserRequest) (*pb.CreateUserResponse, error) {
	if req.User == nil {
		return nil, status.Error(codes.InvalidArgument, "user is required")
	}
	// Hash password if needed
	if req.User.Authentication != nil {
		if basic := req.User.Authentication.GetBasicAuth(); basic != nil {
			if basic.GetPasswordHash() != "" && !strings.HasPrefix(basic.GetPasswordHash(), "$2") {
				hashed, err := passhash.Password(basic.GetPasswordHash())
				if err != nil {
					return nil, status.Errorf(codes.Internal, "failed to hash password: %v", err)
				}
				basic.PasswordHash = proto.String(hashed)
			}
		}
	}
	if err := s.storage.CreateUser(ctx, req.User); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create user: %v", err)
	}
	return &pb.CreateUserResponse{User: req.User}, nil
}

// GetUser retrieves a user by ID.
//
// Parameters:
//   ctx: The context for the request.
//   req: The request object containing the user ID.
//
// Returns:
//   *pb.GetUserResponse: The requested user.
//   error: An error if the user was not found.
func (s *Server) GetUser(ctx context.Context, req *pb.GetUserRequest) (*pb.GetUserResponse, error) {
	user, err := s.storage.GetUser(ctx, req.GetUserId())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get user: %v", err)
	}
	if user == nil {
		return nil, status.Error(codes.NotFound, "user not found")
	}
	return &pb.GetUserResponse{User: user}, nil
}

// ListUsers lists all users.
//
// Parameters:
//   ctx: The context for the request.
//   _ : The request object (unused).
//
// Returns:
//   *pb.ListUsersResponse: The list of all users.
//   error: An error if the operation failed.
func (s *Server) ListUsers(ctx context.Context, _ *pb.ListUsersRequest) (*pb.ListUsersResponse, error) {
	users, err := s.storage.ListUsers(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list users: %v", err)
	}
	return &pb.ListUsersResponse{Users: users}, nil
}

// UpdateUser updates an existing user.
//
// Parameters:
//   ctx: The context for the request.
//   req: The request object containing the updated user information.
//
// Returns:
//   *pb.UpdateUserResponse: The updated user.
//   error: An error if the update failed.
func (s *Server) UpdateUser(ctx context.Context, req *pb.UpdateUserRequest) (*pb.UpdateUserResponse, error) {
	if req.User == nil {
		return nil, status.Error(codes.InvalidArgument, "user is required")
	}
	// Hash password if needed
	if req.User.Authentication != nil {
		if basic := req.User.Authentication.GetBasicAuth(); basic != nil {
			if basic.GetPasswordHash() != "" && !strings.HasPrefix(basic.GetPasswordHash(), "$2") {
				hashed, err := passhash.Password(basic.GetPasswordHash())
				if err != nil {
					return nil, status.Errorf(codes.Internal, "failed to hash password: %v", err)
				}
				basic.PasswordHash = proto.String(hashed)
			}
		}
	}
	if err := s.storage.UpdateUser(ctx, req.User); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update user: %v", err)
	}
	return &pb.UpdateUserResponse{User: req.User}, nil
}

// DeleteUser deletes a user by ID.
//
// Parameters:
//   ctx: The context for the request.
//   req: The request object containing the user ID to delete.
//
// Returns:
//   *pb.DeleteUserResponse: Empty response on success.
//   error: An error if the deletion failed.
func (s *Server) DeleteUser(ctx context.Context, req *pb.DeleteUserRequest) (*pb.DeleteUserResponse, error) {
	if err := s.storage.DeleteUser(ctx, req.GetUserId()); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to delete user: %v", err)
	}
	return &pb.DeleteUserResponse{}, nil
}
