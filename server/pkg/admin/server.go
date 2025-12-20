// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

// Package admin implements the admin server
package admin

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/mcpany/core/pkg/bus"
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
	cache       *middleware.CachingMiddleware
	toolManager tool.ManagerInterface
	bus         *bus.Provider
}

// NewServer creates a new Admin Server.
func NewServer(cache *middleware.CachingMiddleware, toolManager tool.ManagerInterface, bus *bus.Provider) *Server {
	return &Server{
		cache:       cache,
		toolManager: toolManager,
		bus:         bus,
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

// CreateService registers a new upstream service dynamically.
func (s *Server) CreateService(ctx context.Context, req *pb.CreateServiceRequest) (*pb.CreateServiceResponse, error) {
	if req.GetService() == nil {
		return nil, status.Error(codes.InvalidArgument, "service config is required")
	}
	if req.GetService().GetName() == "" {
		return nil, status.Error(codes.InvalidArgument, "service name is required")
	}

	_, err := s.handleRegistrationRequest(ctx, req.GetService())
	if err != nil {
		return nil, err
	}

	return &pb.CreateServiceResponse{Service: req.GetService()}, nil
}

// UpdateService updates an existing upstream service.
func (s *Server) UpdateService(ctx context.Context, req *pb.UpdateServiceRequest) (*pb.UpdateServiceResponse, error) {
	if req.GetService() == nil {
		return nil, status.Error(codes.InvalidArgument, "service config is required")
	}
	if req.GetService().GetName() == "" {
		return nil, status.Error(codes.InvalidArgument, "service name is required")
	}

	// Update is effectively a re-registration in our current architecture
	_, err := s.handleRegistrationRequest(ctx, req.GetService())
	if err != nil {
		return nil, err
	}

	return &pb.UpdateServiceResponse{Service: req.GetService()}, nil
}

// DeleteService unregisters an upstream service.
func (s *Server) DeleteService(ctx context.Context, req *pb.DeleteServiceRequest) (*pb.DeleteServiceResponse, error) {
	if req.GetServiceId() == "" {
		return nil, status.Error(codes.InvalidArgument, "service id is required")
	}

	config := &configv1.UpstreamServiceConfig{
		Name:    proto.String(req.GetServiceId()),
		Disable: proto.Bool(true),
	}

	_, err := s.handleRegistrationRequest(ctx, config)
	if err != nil {
		return nil, err
	}

	return &pb.DeleteServiceResponse{}, nil
}

func (s *Server) handleRegistrationRequest(ctx context.Context, config *configv1.UpstreamServiceConfig) (*bus.ServiceRegistrationResult, error) {
	correlationID := uuid.New().String()
	resultChan := make(chan *bus.ServiceRegistrationResult, 1)

	resultBus := bus.GetBus[*bus.ServiceRegistrationResult](s.bus, bus.ServiceRegistrationResultTopic)
	unsubscribe := resultBus.SubscribeOnce(
		ctx,
		correlationID,
		func(result *bus.ServiceRegistrationResult) {
			resultChan <- result
		},
	)
	defer unsubscribe()

	requestBus := bus.GetBus[*bus.ServiceRegistrationRequest](s.bus, bus.ServiceRegistrationRequestTopic)
	regReq := &bus.ServiceRegistrationRequest{
		Context: ctx,
		Config:  config,
	}
	regReq.SetCorrelationID(correlationID)

	if err := requestBus.Publish(ctx, "request", regReq); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to publish registration request: %v", err)
	}

	select {
	case result := <-resultChan:
		if result.Error != nil {
			return nil, status.Errorf(codes.Internal, "registration failed: %v", result.Error)
		}
		return result, nil
	case <-ctx.Done():
		return nil, status.Error(codes.DeadlineExceeded, "context deadline exceeded")
	case <-time.After(30 * time.Second):
		return nil, status.Error(codes.DeadlineExceeded, "timed out waiting for registration result")
	}
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
