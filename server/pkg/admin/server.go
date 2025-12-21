// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

// Package admin implements the admin server
package admin

import (
	"context"

	"github.com/mcpany/core/pkg/middleware"
	"github.com/mcpany/core/pkg/tool"
	pb "github.com/mcpany/core/proto/admin/v1"
	configv1 "github.com/mcpany/core/proto/config/v1"
	mcprouterv1 "github.com/mcpany/core/proto/mcp_router/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Server implements the AdminServiceServer interface.
type Server struct {
	pb.UnimplementedAdminServiceServer
	cache       *middleware.CachingMiddleware
	toolManager tool.ManagerInterface
}

// NewServer creates a new Admin Server.
//
// Parameters:
//
//	cache: The caching middleware to manage.
//	toolManager: The tool manager for retrieving service and tool information.
//
// Returns:
//
//	A new instance of the Admin Server.
func NewServer(cache *middleware.CachingMiddleware, toolManager tool.ManagerInterface) *Server {
	return &Server{
		cache:       cache,
		toolManager: toolManager,
	}
}

// ClearCache clears the cache.
//
// Parameters:
//
//	ctx: The context for the request.
//	_ : The request proto (unused).
//
// Returns:
//
//	The response proto.
//	An error if clearing the cache fails.
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
//
//	_ : The context for the request (unused).
//	_ : The request proto (unused).
//
// Returns:
//
//	The list of registered services.
//	An error if listing services fails (currently always nil).
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
//
//	_ : The context for the request (unused).
//	req: The request proto containing the service ID.
//
// Returns:
//
//	The service configuration.
//	An error if the service is not found or has no config.
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
//
//	_ : The context for the request (unused).
//	_ : The request proto (unused).
//
// Returns:
//
//	The list of registered tools.
//	An error if listing tools fails (currently always nil).
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
//
//	_ : The context for the request (unused).
//	req: The request proto containing the tool name.
//
// Returns:
//
//	The tool details.
//	An error if the tool is not found.
func (s *Server) GetTool(_ context.Context, req *pb.GetToolRequest) (*pb.GetToolResponse, error) {
	t, ok := s.toolManager.GetTool(req.GetToolName())
	if !ok {
		return nil, status.Error(codes.NotFound, "tool not found")
	}
	return &pb.GetToolResponse{Tool: t.Tool()}, nil
}
