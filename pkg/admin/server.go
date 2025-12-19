// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

// Package admin provides the administrative API for managing the MCP server.
package admin

import (
	"context"

	"github.com/mcpany/core/pkg/middleware"
	"github.com/mcpany/core/pkg/tool"
	pb "github.com/mcpany/core/proto/admin/v1"
	mcp_router_v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/protobuf/proto"
	"google.golang.org/grpc/status"
)

// Server implements the AdminServiceServer interface.
type Server struct {
	pb.UnimplementedAdminServiceServer
	cache       *middleware.CachingMiddleware
	toolManager tool.ManagerInterface
}

// NewServer creates a new Admin Server.
func NewServer(cache *middleware.CachingMiddleware, toolManager tool.ManagerInterface) *Server {
	return &Server{cache: cache, toolManager: toolManager}
}

// ClearCache clears the cache.
func (s *Server) ClearCache(ctx context.Context, _ *pb.ClearCacheRequest) (*pb.ClearCacheResponse, error) {
	if err := s.cache.Clear(ctx); err != nil {
		return nil, err
	}
	return &pb.ClearCacheResponse{}, nil
}

// ListServices returns all registered upstream services.
func (s *Server) ListServices(_ context.Context, _ *pb.ListServicesRequest) (*pb.ListServicesResponse, error) {
	services := s.toolManager.ListServices()
	serviceSummaries := make([]*pb.ServiceSummary, 0, len(services))
	for _, svc := range services {
		// Use SanitizedName as ID because toolManager is indexed by SanitizedName
		id := svc.Config.GetSanitizedName()
		if id == "" {
			id = svc.Name
		}
		serviceSummaries = append(serviceSummaries, &pb.ServiceSummary{
			Id:   proto.String(id),
			Name: proto.String(svc.Name),
		})
	}
	return &pb.ListServicesResponse{Services: serviceSummaries}, nil
}

// GetService returns the configuration of a specific upstream service.
func (s *Server) GetService(_ context.Context, req *pb.GetServiceRequest) (*pb.GetServiceResponse, error) {
	info, ok := s.toolManager.GetServiceInfo(req.GetId())
	if !ok {
		return nil, status.Errorf(codes.NotFound, "service not found: %s", req.GetId())
	}
	return &pb.GetServiceResponse{Config: info.Config}, nil
}

// ListTools returns all registered tools.
func (s *Server) ListTools(_ context.Context, req *pb.ListToolsRequest) (*pb.ListToolsResponse, error) {
	allTools := s.toolManager.ListTools()
	tools := make([]*mcp_router_v1.Tool, 0, len(allTools))

	for _, t := range allTools {
		protoTool := t.Tool()
		if req.GetServiceId() != "" && protoTool.GetServiceId() != req.GetServiceId() {
			continue
		}
		tools = append(tools, protoTool)
	}
	return &pb.ListToolsResponse{Tools: tools}, nil
}

// GetTool returns the definition of a specific tool.
func (s *Server) GetTool(_ context.Context, req *pb.GetToolRequest) (*pb.GetToolResponse, error) {
	t, ok := s.toolManager.GetTool(req.GetName())
	if !ok {
		return nil, status.Errorf(codes.NotFound, "tool not found: %s", req.GetName())
	}
	return &pb.GetToolResponse{Tool: t.Tool()}, nil
}
