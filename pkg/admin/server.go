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
	"google.golang.org/protobuf/proto"
)

// Server implements the AdminServiceServer interface.
type Server struct {
	pb.UnimplementedAdminServiceServer
	cache         *middleware.CachingMiddleware
	toolManager   tool.ManagerInterface
	healthManager *health.Manager
}

// NewServer creates a new Admin Server.
func NewServer(cache *middleware.CachingMiddleware, toolManager tool.ManagerInterface, healthManager *health.Manager) *Server {
	return &Server{
		cache:         cache,
		toolManager:   toolManager,
		healthManager: healthManager,
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
	var serviceStates []*pb.ServiceState
	var services []*configv1.UpstreamServiceConfig

	for _, info := range serviceInfos {
		if info.Config != nil {
			services = append(services, info.Config)

			sStatus := pb.ServiceStatus_SERVICE_STATUS_UNKNOWN
			state := &pb.ServiceState{
				Config: info.Config,
				Status: &sStatus,
			}

			if s.healthManager != nil {
				// Use SanitizedName as key, consistent with how it's registered
				if h := s.healthManager.GetStatus(info.Config.GetSanitizedName()); h != nil {
					hStatus := h.Status
					state.Status = &hStatus
					state.LastError = proto.String(h.LastError)
					if !h.LastCheckTime.IsZero() {
						state.LastCheckTime = proto.Int64(h.LastCheckTime.Unix())
					}
				}
			}
			serviceStates = append(serviceStates, state)
		}
	}
	return &pb.ListServicesResponse{Services: services, ServiceStates: serviceStates}, nil
}

// GetService returns a specific service by ID.
func (s *Server) GetService(_ context.Context, req *pb.GetServiceRequest) (*pb.GetServiceResponse, error) {
	info, ok := s.toolManager.GetServiceInfo(req.GetServiceId())
	if !ok {
		return nil, status.Error(codes.NotFound, "service not found")
	}
	if info.Config == nil {
		return nil, status.Error(codes.Internal, "service config not found")
	}

	sStatus := pb.ServiceStatus_SERVICE_STATUS_UNKNOWN
	state := &pb.ServiceState{
		Config: info.Config,
		Status: &sStatus,
	}

	if s.healthManager != nil {
		if h := s.healthManager.GetStatus(info.Config.GetSanitizedName()); h != nil {
			hStatus := h.Status
			state.Status = &hStatus
			state.LastError = proto.String(h.LastError)
			if !h.LastCheckTime.IsZero() {
				state.LastCheckTime = proto.Int64(h.LastCheckTime.Unix())
			}
		}
	}

	return &pb.GetServiceResponse{Service: info.Config, ServiceState: state}, nil
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
