// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package admin

import (
	"context"

	"github.com/mcpany/core/pkg/health"
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
	health      *health.Manager
}

// NewServer creates a new Admin Server.
func NewServer(cache *middleware.CachingMiddleware, toolManager tool.ManagerInterface, health *health.Manager) *Server {
	return &Server{
		cache:       cache,
		toolManager: toolManager,
		health:      health,
	}
}

// ClearCache clears the cache.
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
func (s *Server) ListServices(_ context.Context, _ *pb.ListServicesRequest) (*pb.ListServicesResponse, error) {
	serviceInfos := s.toolManager.ListServices()
	var services []*configv1.UpstreamServiceConfig
	var serviceStates []*pb.ServiceState

	for _, info := range serviceInfos {
		if info.Config != nil {
			services = append(services, info.Config)

			// Try to get status from health manager
			id := info.Config.GetId()
			if id == "" {
				id = info.Name
			}

			if s.health != nil {
				if state, ok := s.health.GetState(id); ok {
					serviceStates = append(serviceStates, state)
					continue
				}
			}

			// Fallback
			status := pb.ServiceStatus_SERVICE_STATUS_UNKNOWN
			serviceStates = append(serviceStates, &pb.ServiceState{
				Config: info.Config,
				Status: &status,
			})
		}
	}
	return &pb.ListServicesResponse{
		Services:      services,
		ServiceStates: serviceStates,
	}, nil
}

// GetService returns a specific service by ID.
func (s *Server) GetService(_ context.Context, req *pb.GetServiceRequest) (*pb.GetServiceResponse, error) {
	id := req.GetServiceId()
	info, ok := s.toolManager.GetServiceInfo(id)
	if !ok {
		return nil, status.Error(codes.NotFound, "service not found")
	}
	if info.Config == nil {
		return nil, status.Error(codes.Internal, "service config not found")
	}

	var state *pb.ServiceState
	if s.health != nil {
		if s, ok := s.health.GetState(id); ok {
			state = s
		}
	}

	if state == nil {
		status := pb.ServiceStatus_SERVICE_STATUS_UNKNOWN
		state = &pb.ServiceState{
			Config: info.Config,
			Status: &status,
		}
	}

	return &pb.GetServiceResponse{
		Service:      info.Config,
		ServiceState: state,
	}, nil
}

// ListTools returns all registered tools.
func (s *Server) ListTools(_ context.Context, _ *pb.ListToolsRequest) (*pb.ListToolsResponse, error) {
	tools := s.toolManager.ListTools()
	var responseTools []*mcprouterv1.Tool
	for _, t := range tools {
		responseTools = append(responseTools, t.Tool())
	}
	return &pb.ListToolsResponse{Tools: responseTools}, nil
}

// GetTool returns a specific tool by name.
func (s *Server) GetTool(_ context.Context, req *pb.GetToolRequest) (*pb.GetToolResponse, error) {
	t, ok := s.toolManager.GetTool(req.GetToolName())
	if !ok {
		return nil, status.Error(codes.NotFound, "tool not found")
	}
	return &pb.GetToolResponse{Tool: t.Tool()}, nil
}
