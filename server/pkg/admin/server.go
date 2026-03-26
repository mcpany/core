// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

// Package admin implements the admin server
package admin

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	pb "github.com/mcpany/core/proto/admin/v1"
	configv1 "github.com/mcpany/core/proto/config/v1"
	mcprouterv1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/mcpany/core/server/pkg/audit"
	"github.com/mcpany/core/server/pkg/config"
	"github.com/mcpany/core/server/pkg/discovery"
	"github.com/mcpany/core/server/pkg/middleware"
	"github.com/mcpany/core/server/pkg/serviceregistry"
	"github.com/mcpany/core/server/pkg/storage"
	"github.com/mcpany/core/server/pkg/tool"
	"github.com/mcpany/core/server/pkg/util/passhash"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)

// Server implements the AdminServiceServer interface.
//
// Summary: gRPC server implementation for the Admin service.
type Server struct {
	pb.UnimplementedAdminServiceServer
	cache            *middleware.CachingMiddleware
	toolManager      tool.ManagerInterface
	serviceRegistry  serviceregistry.ServiceRegistryInterface
	storage          storage.Storage
	discoveryManager *discovery.Manager
	auditMiddleware  *middleware.AuditMiddleware
}

// NewServer creates a new Admin Server.
//
// Summary: Initializes NewServer operation.
//
// Parameters:
//   - cache: *middleware.CachingMiddleware. The caching middleware instance.
//   - toolManager: tool.ManagerInterface. The tool manager instance.
//   - serviceRegistry: serviceregistry.ServiceRegistryInterface. The service registry.
//   - storage: storage.Storage. The storage backend.
//   - discoveryManager: *discovery.Manager. The discovery manager.
//   - auditMiddleware: *middleware.AuditMiddleware. The audit middleware.
//
// Returns:
//   - *Server: The initialized admin server.
func NewServer(
	cache *middleware.CachingMiddleware,
	toolManager tool.ManagerInterface,
	serviceRegistry serviceregistry.ServiceRegistryInterface,
	storage storage.Storage,
	discoveryManager *discovery.Manager,
	auditMiddleware *middleware.AuditMiddleware,
) *Server {
	return &Server{
		cache:            cache,
		toolManager:      toolManager,
		serviceRegistry:  serviceRegistry,
		storage:          storage,
		discoveryManager: discoveryManager,
		auditMiddleware:  auditMiddleware,
	}
}

// ClearCache clears the cache.
//
// Summary: Executes ClearCache operation.
//
// Parameters:
//   - ctx: context.Context. The request context.
//   - req: *pb.ClearCacheRequest. The clear cache request.
//
// Returns:
//   - *pb.ClearCacheResponse: The clear cache response.
//   - error: An error if the cache cannot be cleared.
//
// Errors:
//   - Returns an error if the operation fails.
//
// Side Effects:
//   - None.
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
// Summary: Executes ListServices operation.
//
// Parameters:
//   - ctx: context.Context. The request context.
//   - req: *pb.ListServicesRequest. The list services request.
//
// Returns:
//   - *pb.ListServicesResponse: The list services response.
//   - error: An error if retrieval fails.
//
// Errors:
//   - Returns an error if the operation fails.
//
// Side Effects:
//   - None.
func (s *Server) ListServices(_ context.Context, _ *pb.ListServicesRequest) (*pb.ListServicesResponse, error) {
	var services []*configv1.UpstreamServiceConfig
	var serviceStates []*pb.ServiceState

	if s.serviceRegistry != nil {
		configs, err := s.serviceRegistry.GetAllServices()
		if err != nil {
			return nil, status.Errorf(codes.Internal, "failed to list services: %v", err)
		}
		for _, cfg := range configs {
			safeCfg := proto.Clone(cfg).(*configv1.UpstreamServiceConfig)
			config.StripSecretsFromService(safeCfg)
			services = append(services, safeCfg)

			state := pb.ServiceState_builder{
				Config: safeCfg,
				Status: proto.String("OK"),
			}.Build()
			if errMsg, ok := s.serviceRegistry.GetServiceError(cfg.GetId()); ok {
				state.SetStatus("ERROR")
				state.SetError(errMsg)
			}
			serviceStates = append(serviceStates, state)
		}
	} else {
		// Fallback to toolManager if serviceRegistry is not set (e.g. tests)
		serviceInfos := s.toolManager.ListServices()
		for _, info := range serviceInfos {
			if info.Config != nil {
				safeCfg := proto.Clone(info.Config).(*configv1.UpstreamServiceConfig)
				config.StripSecretsFromService(safeCfg)
				services = append(services, safeCfg)
				serviceStates = append(serviceStates, pb.ServiceState_builder{
					Config: safeCfg,
					Status: proto.String("OK"),
				}.Build())
			}
		}
	}

	return pb.ListServicesResponse_builder{
		Services:      services,
		ServiceStates: serviceStates,
	}.Build(), nil
}

// GetService returns a specific service by ID.
//
// Summary: Retrieves GetService operation.
//
// Parameters:
//   - ctx: context.Context. The request context.
//   - req: *pb.GetServiceRequest. The get service request.
//
// Returns:
//   - *pb.GetServiceResponse: The get service response.
//   - error: An error if the service is not found.
//
// Errors:
//   - Returns an error if the operation fails.
//
// Side Effects:
//   - None.
func (s *Server) GetService(_ context.Context, req *pb.GetServiceRequest) (*pb.GetServiceResponse, error) {
	if s.serviceRegistry != nil {
		cfg, ok := s.serviceRegistry.GetServiceConfig(req.GetServiceId())
		if !ok {
			return nil, status.Error(codes.NotFound, "service not found")
		}
		safeCfg := proto.Clone(cfg).(*configv1.UpstreamServiceConfig)
		config.StripSecretsFromService(safeCfg)

		state := pb.ServiceState_builder{
			Config: safeCfg,
			Status: proto.String("OK"),
		}.Build()
		if errMsg, ok := s.serviceRegistry.GetServiceError(cfg.GetId()); ok {
			state.SetStatus("ERROR")
			state.SetError(errMsg)
		}
		return pb.GetServiceResponse_builder{
			Service:      safeCfg,
			ServiceState: state,
		}.Build(), nil
	}

	info, ok := s.toolManager.GetServiceInfo(req.GetServiceId())
	if !ok {
		return nil, status.Error(codes.NotFound, "service not found")
	}
	if info.Config == nil {
		return nil, status.Error(codes.Internal, "service config not found")
	}
	safeCfg := proto.Clone(info.Config).(*configv1.UpstreamServiceConfig)
	config.StripSecretsFromService(safeCfg)

	return pb.GetServiceResponse_builder{
		Service: safeCfg,
		ServiceState: pb.ServiceState_builder{
			Config: safeCfg,
			Status: proto.String("OK"),
		}.Build(),
	}.Build(), nil
}

// ListTools returns all registered tools.
//
// Summary: Executes ListTools operation.
//
// Parameters:
//   - ctx: context.Context. The request context.
//   - req: *pb.ListToolsRequest. The list tools request.
//
// Returns:
//   - *pb.ListToolsResponse: The list tools response.
//   - error: An error if retrieval fails.
//
// Errors:
//   - Returns an error if the operation fails.
//
// Side Effects:
//   - None.
func (s *Server) ListTools(_ context.Context, _ *pb.ListToolsRequest) (*pb.ListToolsResponse, error) {
	tools := s.toolManager.ListTools()
	responseTools := make([]*mcprouterv1.Tool, 0, len(tools))
	for _, t := range tools {
		responseTools = append(responseTools, t.Tool())
	}
	return pb.ListToolsResponse_builder{Tools: responseTools}.Build(), nil
}

// GetTool returns a specific tool by name.
//
// Summary: Retrieves GetTool operation.
//
// Parameters:
//   - ctx: context.Context. The request context.
//   - req: *pb.GetToolRequest. The get tool request.
//
// Returns:
//   - *pb.GetToolResponse: The get tool response.
//   - error: An error if the tool is not found.
//
// Errors:
//   - Returns an error if the operation fails.
//
// Side Effects:
//   - None.
func (s *Server) GetTool(_ context.Context, req *pb.GetToolRequest) (*pb.GetToolResponse, error) {
	t, ok := s.toolManager.GetTool(req.GetToolName())
	if !ok {
		return nil, status.Error(codes.NotFound, "tool not found")
	}
	return pb.GetToolResponse_builder{Tool: t.Tool()}.Build(), nil
}

// CreateUser creates a new user.
//
// Summary: Initializes CreateUser operation.
//
// Parameters:
//   - ctx: context.Context. The request context.
//   - req: *pb.CreateUserRequest. The create user request.
//
// Returns:
//   - *pb.CreateUserResponse: The create user response.
//   - error: An error if the user cannot be created.
//
// Errors:
//   - Returns an error if the operation fails.
//
// Side Effects:
//   - None.
func (s *Server) CreateUser(ctx context.Context, req *pb.CreateUserRequest) (*pb.CreateUserResponse, error) {
	if !req.HasUser() {
		return nil, status.Error(codes.InvalidArgument, "user is required")
	}
	// Hash password if needed
	if req.GetUser().HasAuthentication() {
		if basic := req.GetUser().GetAuthentication().GetBasicAuth(); basic != nil {
			if basic.GetPasswordHash() != "" && !strings.HasPrefix(basic.GetPasswordHash(), "$2") {
				hashed, err := passhash.Password(basic.GetPasswordHash())
				if err != nil {
					return nil, status.Errorf(codes.Internal, "failed to hash password: %v", err)
				}
				basic.SetPasswordHash(hashed)
			}
		}
	}
	if err := s.storage.CreateUser(ctx, req.GetUser()); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create user: %v", err)
	}

	safeUser := proto.Clone(req.GetUser()).(*configv1.User)
	config.StripSecretsFromAuth(safeUser.GetAuthentication())
	return pb.CreateUserResponse_builder{User: safeUser}.Build(), nil
}

// GetUser retrieves a user by ID.
//
// Summary: Retrieves GetUser operation.
//
// Parameters:
//   - ctx: context.Context. The request context.
//   - req: *pb.GetUserRequest. The get user request.
//
// Returns:
//   - *pb.GetUserResponse: The get user response.
//   - error: An error if the user is not found.
//
// Errors:
//   - Returns an error if the operation fails.
//
// Side Effects:
//   - None.
func (s *Server) GetUser(ctx context.Context, req *pb.GetUserRequest) (*pb.GetUserResponse, error) {
	user, err := s.storage.GetUser(ctx, req.GetUserId())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get user: %v", err)
	}
	if user == nil {
		return nil, status.Error(codes.NotFound, "user not found")
	}

	safeUser := proto.Clone(user).(*configv1.User)
	config.StripSecretsFromAuth(safeUser.GetAuthentication())
	return pb.GetUserResponse_builder{User: safeUser}.Build(), nil
}

// ListUsers lists all users.
//
// Summary: Executes ListUsers operation.
//
// Parameters:
//   - ctx: context.Context. The request context.
//   - req: *pb.ListUsersRequest. The list users request.
//
// Returns:
//   - *pb.ListUsersResponse: The list users response.
//   - error: An error if retrieval fails.
//
// Errors:
//   - Returns an error if the operation fails.
//
// Side Effects:
//   - None.
func (s *Server) ListUsers(ctx context.Context, _ *pb.ListUsersRequest) (*pb.ListUsersResponse, error) {
	users, err := s.storage.ListUsers(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list users: %v", err)
	}

	safeUsers := make([]*configv1.User, 0, len(users))
	for _, u := range users {
		safeUser := proto.Clone(u).(*configv1.User)
		config.StripSecretsFromAuth(safeUser.GetAuthentication())
		safeUsers = append(safeUsers, safeUser)
	}

	return pb.ListUsersResponse_builder{Users: safeUsers}.Build(), nil
}

// UpdateUser updates an existing user.
//
// Summary: Executes UpdateUser operation.
//
// Parameters:
//   - ctx: context.Context. The request context.
//   - req: *pb.UpdateUserRequest. The update user request.
//
// Returns:
//   - *pb.UpdateUserResponse: The update user response.
//   - error: An error if the user cannot be updated.
//
// Errors:
//   - Returns an error if the operation fails.
//
// Side Effects:
//   - None.
func (s *Server) UpdateUser(ctx context.Context, req *pb.UpdateUserRequest) (*pb.UpdateUserResponse, error) {
	if !req.HasUser() {
		return nil, status.Error(codes.InvalidArgument, "user is required")
	}
	// Hash password if needed
	if req.GetUser().HasAuthentication() {
		if basic := req.GetUser().GetAuthentication().GetBasicAuth(); basic != nil {
			if basic.GetPasswordHash() != "" && !strings.HasPrefix(basic.GetPasswordHash(), "$2") {
				hashed, err := passhash.Password(basic.GetPasswordHash())
				if err != nil {
					return nil, status.Errorf(codes.Internal, "failed to hash password: %v", err)
				}
				basic.SetPasswordHash(hashed)
			}
		}
	}
	if err := s.storage.UpdateUser(ctx, req.GetUser()); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update user: %v", err)
	}

	safeUser := proto.Clone(req.GetUser()).(*configv1.User)
	config.StripSecretsFromAuth(safeUser.GetAuthentication())
	return pb.UpdateUserResponse_builder{User: safeUser}.Build(), nil
}

// DeleteUser deletes a user by ID.
//
// Summary: Executes DeleteUser operation.
//
// Parameters:
//   - ctx: context.Context. The request context.
//   - req: *pb.DeleteUserRequest. The delete user request.
//
// Returns:
//   - *pb.DeleteUserResponse: The delete user response.
//   - error: An error if the user cannot be deleted.
//
// Errors:
//   - Returns an error if the operation fails.
//
// Side Effects:
//   - None.
func (s *Server) DeleteUser(ctx context.Context, req *pb.DeleteUserRequest) (*pb.DeleteUserResponse, error) {
	if err := s.storage.DeleteUser(ctx, req.GetUserId()); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to delete user: %v", err)
	}
	return &pb.DeleteUserResponse{}, nil
}

// GetDiscoveryStatus returns the status of auto-discovery providers.
//
// Summary: Retrieves GetDiscoveryStatus operation.
//
// Parameters:
//   - ctx: context.Context. The request context.
//   - req: *pb.GetDiscoveryStatusRequest. The get discovery status request.
//
// Returns:
//   - *pb.GetDiscoveryStatusResponse: The get discovery status response.
//   - error: An error if the operation fails.
//
// Errors:
//   - Returns an error if the operation fails.
//
// Side Effects:
//   - None.
func (s *Server) GetDiscoveryStatus(_ context.Context, _ *pb.GetDiscoveryStatusRequest) (*pb.GetDiscoveryStatusResponse, error) {
	if s.discoveryManager == nil {
		return &pb.GetDiscoveryStatusResponse{}, nil
	}

	statuses := s.discoveryManager.GetStatuses()
	pbStatuses := make([]*pb.DiscoveryProviderStatus, 0, len(statuses))

	for _, st := range statuses {
		//nolint:gosec // Discovered count fits in int32
		pbStatuses = append(pbStatuses, pb.DiscoveryProviderStatus_builder{
			Name:            proto.String(st.Name),
			Status:          proto.String(st.Status),
			LastError:       proto.String(st.LastError),
			LastRunAt:       proto.String(st.LastRunAt.Format("2006-01-02T15:04:05Z07:00")),
			DiscoveredCount: proto.Int32(int32(st.DiscoveredCount)),
		}.Build())
	}

	return pb.GetDiscoveryStatusResponse_builder{Providers: pbStatuses}.Build(), nil
}

// ListAuditLogs returns audit logs matching the filter.
//
// Summary: Executes ListAuditLogs operation.
//
// Parameters:
//   - ctx: context.Context. The request context.
//   - req: *pb.ListAuditLogsRequest. The list audit logs request.
//
// Returns:
//   - *pb.ListAuditLogsResponse: The list audit logs response.
//   - error: An error if the operation fails.
//
// Errors:
//   - Returns an error if the operation fails.
//
// Side Effects:
//   - None.
func (s *Server) ListAuditLogs(ctx context.Context, req *pb.ListAuditLogsRequest) (*pb.ListAuditLogsResponse, error) {
	if s.auditMiddleware == nil {
		return nil, status.Error(codes.FailedPrecondition, "audit logging is not enabled")
	}

	var startTime, endTime *time.Time
	if req.GetStartTime() != "" {
		t, err := time.Parse(time.RFC3339, req.GetStartTime())
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "invalid start_time: %v", err)
		}
		startTime = &t
	}
	if req.GetEndTime() != "" {
		t, err := time.Parse(time.RFC3339, req.GetEndTime())
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "invalid end_time: %v", err)
		}
		endTime = &t
	}

	filter := audit.Filter{
		StartTime: startTime,
		EndTime:   endTime,
		ToolName:  req.GetToolName(),
		UserID:    req.GetUserId(),
		ProfileID: req.GetProfileId(),
		Limit:     int(req.GetLimit()),
		Offset:    int(req.GetOffset()),
	}

	entries, err := s.auditMiddleware.Read(ctx, filter)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to read audit logs: %v", err)
	}

	pbEntries := make([]*pb.AuditLogEntry, 0, len(entries))
	for _, e := range entries {
		var argsStr, resultStr string
		if len(e.Arguments) > 0 {
			argsStr = string(e.Arguments)
		}
		if e.Result != nil {
			if b, err := json.Marshal(e.Result); err == nil {
				resultStr = string(b)
			} else {
				// Fallback if marshalling fails (unlikely if it came from JSON)
				resultStr = fmt.Sprintf("%v", e.Result)
			}
		}
		pbEntries = append(pbEntries, pb.AuditLogEntry_builder{
			Timestamp:  proto.String(e.Timestamp.Format(time.RFC3339)),
			ToolName:   proto.String(e.ToolName),
			UserId:     proto.String(e.UserID),
			ProfileId:  proto.String(e.ProfileID),
			TraceId:    proto.String(e.TraceID),
			SpanId:     proto.String(e.SpanID),
			ParentId:   proto.String(e.ParentID),
			Arguments:  proto.String(argsStr),
			Result:     proto.String(resultStr),
			Error:      proto.String(e.Error),
			Duration:   proto.String(e.Duration),
			DurationMs: proto.Int64(e.DurationMs),
		}.Build())
	}
	return pb.ListAuditLogsResponse_builder{Entries: pbEntries}.Build(), nil
}
