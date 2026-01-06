// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

// Package admin implements the admin server
package admin

import (
	"context"

	"github.com/mcpany/core/server/pkg/middleware"
	"github.com/mcpany/core/server/pkg/serviceregistry"
	"github.com/mcpany/core/server/pkg/storage"
	"github.com/mcpany/core/server/pkg/tool"
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
	storage         storage.Storage
}

// NewServer creates a new Admin Server.
func NewServer(cache *middleware.CachingMiddleware, toolManager tool.ManagerInterface, serviceRegistry serviceregistry.ServiceRegistryInterface, storage storage.Storage) *Server {
	return &Server{
		cache:           cache,
		toolManager:     toolManager,
		serviceRegistry: serviceRegistry,
		storage:         storage,
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

// RegisterService registers a new upstream service.
func (s *Server) RegisterService(ctx context.Context, req *pb.RegisterServiceRequest) (*pb.RegisterServiceResponse, error) {
	if s.serviceRegistry == nil {
		return nil, status.Error(codes.FailedPrecondition, "service registry is not available")
	}

	config := req.GetServiceConfig()
	if config == nil {
		return nil, status.Error(codes.InvalidArgument, "service config is required")
	}
	if config.GetName() == "" {
		return nil, status.Error(codes.InvalidArgument, "service name is required")
	}

	serviceID, _, _, err := s.serviceRegistry.RegisterService(ctx, config)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to register service: %v", err)
	}

	// Persist the service configuration if storage is available
	if s.storage != nil {
		// Ensure the ID matches what was generated/used by the registry
		config.Id = &serviceID
		if err := s.storage.SaveService(ctx, config); err != nil {
			// Log error but don't fail the request as the service is running in memory
			// Ideally we should rollback, but for now we just warn.
			// logging.GetLogger().Error("Failed to persist service config", "error", err)
			return nil, status.Errorf(codes.Internal, "service registered but failed to persist: %v", err)
		}
	}

	return &pb.RegisterServiceResponse{ServiceId: &serviceID}, nil
}

// UnregisterService unregisters an existing upstream service.
func (s *Server) UnregisterService(ctx context.Context, req *pb.UnregisterServiceRequest) (*pb.UnregisterServiceResponse, error) {
	if s.serviceRegistry == nil {
		return nil, status.Error(codes.FailedPrecondition, "service registry is not available")
	}

	serviceName := req.GetServiceName()
	if serviceName == "" {
		return nil, status.Error(codes.InvalidArgument, "service name is required")
	}

	// Unregister from memory
	if err := s.serviceRegistry.UnregisterService(ctx, serviceName); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to unregister service: %v", err)
	}

	// Remove from persistence if storage is available
	if s.storage != nil {
		if err := s.storage.DeleteService(ctx, serviceName); err != nil {
			// Log warning
			return nil, status.Errorf(codes.Internal, "service unregistered but failed to delete from storage: %v", err)
		}
	}

	return &pb.UnregisterServiceResponse{}, nil
}
