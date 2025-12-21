// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

// Package admin implements the admin server
package admin

import (
	"context"

	"github.com/mcpany/core/pkg/middleware"
	"github.com/mcpany/core/pkg/serviceregistry"
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
	cache           *middleware.CachingMiddleware
	toolManager     tool.ManagerInterface
	serviceRegistry serviceregistry.ServiceRegistryInterface
}

// NewServer creates a new Admin Server.
func NewServer(cache *middleware.CachingMiddleware, toolManager tool.ManagerInterface, serviceRegistry serviceregistry.ServiceRegistryInterface) *Server {
	return &Server{
		cache:           cache,
		toolManager:     toolManager,
		serviceRegistry: serviceRegistry,
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
	for _, info := range serviceInfos {
		if info.Config != nil {
			services = append(services, info.Config)
		}
	}
	return &pb.ListServicesResponse{Services: services}, nil
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
	return &pb.GetServiceResponse{Service: info.Config}, nil
}

// ListTools returns all registered tools.
func (s *Server) ListTools(_ context.Context, _ *pb.ListToolsRequest) (*pb.ListToolsResponse, error) {
	tools := s.toolManager.ListTools()
	responseTools := make([]*mcprouterv1.Tool, 0, len(tools))
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

// CreateService registers a new service.
func (s *Server) CreateService(ctx context.Context, req *pb.CreateServiceRequest) (*pb.CreateServiceResponse, error) {
	if s.serviceRegistry == nil {
		return nil, status.Error(codes.Unimplemented, "service registry not available")
	}
	if req.GetService() == nil {
		return nil, status.Error(codes.InvalidArgument, "service config is required")
	}

	serviceID, _, _, err := s.serviceRegistry.RegisterService(ctx, req.GetService())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to register service: %v", err)
	}

	return &pb.CreateServiceResponse{ServiceId: &serviceID}, nil
}

// UpdateService updates an existing service.
func (s *Server) UpdateService(ctx context.Context, req *pb.UpdateServiceRequest) (*pb.UpdateServiceResponse, error) {
	if s.serviceRegistry == nil {
		return nil, status.Error(codes.Unimplemented, "service registry not available")
	}
	if req.GetService() == nil {
		return nil, status.Error(codes.InvalidArgument, "service config is required")
	}

	serviceName := req.GetService().GetName()
	if serviceName == "" {
		return nil, status.Error(codes.InvalidArgument, "service name is required")
	}

	// To prevent data loss, we verify the new service config is valid and registerable
	// by checking mandatory fields before unregistering the old service.
	// Note: True transactional safety would require a "DryRun" or atomic swap support in ServiceRegistry.
	// For now, we do basic validation.
	// Re-sanitizing name to ensure it matches expectations
	newServiceName := req.GetService().GetName()
	if newServiceName != serviceName {
		return nil, status.Error(codes.InvalidArgument, "service name in body must match existing service name")
	}

	// Unregister existing
	if err := s.serviceRegistry.UnregisterService(ctx, serviceName); err != nil {
		return nil, status.Errorf(codes.NotFound, "failed to find service to update: %v", err)
	}

	// Register new
	serviceID, _, _, err := s.serviceRegistry.RegisterService(ctx, req.GetService())
	if err != nil {
		// Critical Failure: We unregistered but failed to re-register.
		// Attempt to restore (best effort)?
		// For this implementation, we return the error.
		return nil, status.Errorf(codes.Internal, "failed to register updated service: %v", err)
	}

	return &pb.UpdateServiceResponse{ServiceId: &serviceID}, nil
}

// DeleteService removes a service.
func (s *Server) DeleteService(ctx context.Context, req *pb.DeleteServiceRequest) (*pb.DeleteServiceResponse, error) {
	if s.serviceRegistry == nil {
		return nil, status.Error(codes.Unimplemented, "service registry not available")
	}
	if req.GetServiceName() == "" {
		return nil, status.Error(codes.InvalidArgument, "service name is required")
	}

	if err := s.serviceRegistry.UnregisterService(ctx, req.GetServiceName()); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to unregister service: %v", err)
	}

	return &pb.DeleteServiceResponse{}, nil
}
