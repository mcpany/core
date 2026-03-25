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
// Summary: Implements the AdminServiceServer interface for administrative operations.
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
// Summary: Initializes a new Server instance with its required dependencies.
// Parameters:
//   - cache (*middleware.CachingMiddleware): The caching middleware for the server.
//   - toolManager (tool.ManagerInterface): The manager for tool-related operations.
//   - serviceRegistry (serviceregistry.ServiceRegistryInterface): The registry for upstream services.
//   - storage (storage.Storage): The persistence layer for the server.
//   - discoveryManager (*discovery.Manager): The manager for service discovery.
//   - auditMiddleware (*middleware.AuditMiddleware): The middleware for audit logging.
//
// Returns:
//   - *Server: The newly created Server instance.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
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

// ClearCache clears the server's cache.
// Summary: Executes the ClearCache operation on the caching middleware.
// Parameters:
//   - ctx (context.Context): The context for the request.
//   - _ (*pb.ClearCacheRequest): The request parameters (unused).
//
// Returns:
//   - *pb.ClearCacheResponse: An empty response on success.
//   - error: An error if caching is disabled or the clear operation fails.
//
// Errors:
//   - FailedPrecondition: If caching is not enabled.
//   - Internal: If the cache clear operation fails.
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
// Summary: Lists all upstream services managed by the server.
// Parameters:
//   - ctx (context.Context): The context for the request.
//   - _ (*pb.ListServicesRequest): The request parameters (unused).
//
// Returns:
//   - *pb.ListServicesResponse: The list of services and their current states.
//   - error: An error if the service listing fails.
//
// Errors:
//   - Internal: If the service registry fails to retrieve services.
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
// Summary: Retrieves the configuration and state of a specific upstream service.
// Parameters:
//   - ctx (context.Context): The context for the request.
//   - req (*pb.GetServiceRequest): The request containing the service ID.
//
// Returns:
//   - *pb.GetServiceResponse: The requested service configuration and state.
//   - error: An error if the service is not found or retrieval fails.
//
// Errors:
//   - NotFound: If the service ID does not exist.
//   - Internal: If the service configuration is missing for an existing service.
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
// Summary: Lists all tools available through the server.
// Parameters:
//   - ctx (context.Context): The context for the request.
//   - _ (*pb.ListToolsRequest): The request parameters (unused).
//
// Returns:
//   - *pb.ListToolsResponse: The list of registered tools.
//   - error: An error if the tool listing fails.
//
// Errors:
//   - None.
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
// Summary: Retrieves the details of a specific tool by its name.
// Parameters:
//   - ctx (context.Context): The context for the request.
//   - req (*pb.GetToolRequest): The request containing the tool name.
//
// Returns:
//   - *pb.GetToolResponse: The requested tool details.
//   - error: An error if the tool is not found.
//
// Errors:
//   - NotFound: If the tool name does not exist.
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

// CreateUser creates a new user in the system.
// Summary: Registers a new user and persists their credentials.
// Parameters:
//   - ctx (context.Context): The context for the request.
//   - req (*pb.CreateUserRequest): The request containing user details and credentials.
//
// Returns:
//   - *pb.CreateUserResponse: The newly created user details (secrets redacted).
//   - error: An error if the user cannot be created.
//
// Errors:
//   - InvalidArgument: If the user object is missing from the request.
//   - Internal: If password hashing or storage operations fail.
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

// GetUser retrieves a user by their ID.
// Summary: Fetches the details of a specific user.
// Parameters:
//   - ctx (context.Context): The context for the request.
//   - req (*pb.GetUserRequest): The request containing the user ID.
//
// Returns:
//   - *pb.GetUserResponse: The requested user details (secrets redacted).
//   - error: An error if the user is not found or retrieval fails.
//
// Errors:
//   - NotFound: If the user ID does not exist.
//   - Internal: If the storage operation fails.
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

// ListUsers lists all users in the system.
// Summary: Retrieves a list of all registered users.
// Parameters:
//   - ctx (context.Context): The context for the request.
//   - _ (*pb.ListUsersRequest): The request parameters (unused).
//
// Returns:
//   - *pb.ListUsersResponse: The list of users (secrets redacted).
//   - error: An error if listing users fails.
//
// Errors:
//   - Internal: If the storage operation fails.
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

// UpdateUser updates an existing user's details.
// Summary: Modifies the details and credentials of an existing user.
// Parameters:
//   - ctx (context.Context): The context for the request.
//   - req (*pb.UpdateUserRequest): The request containing updated user details.
//
// Returns:
//   - *pb.UpdateUserResponse: The updated user details (secrets redacted).
//   - error: An error if the user cannot be updated.
//
// Errors:
//   - InvalidArgument: If the user object is missing from the request.
//   - Internal: If password hashing or storage operations fail.
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

// DeleteUser removes a user from the system by ID.
// Summary: Deletes a specific user from storage.
// Parameters:
//   - ctx (context.Context): The context for the request.
//   - req (*pb.DeleteUserRequest): The request containing the user ID to delete.
//
// Returns:
//   - *pb.DeleteUserResponse: An empty response on success.
//   - error: An error if the user cannot be deleted.
//
// Errors:
//   - Internal: If the storage operation fails.
//
// Side Effects:
//   - None.
func (s *Server) DeleteUser(ctx context.Context, req *pb.DeleteUserRequest) (*pb.DeleteUserResponse, error) {
	if err := s.storage.DeleteUser(ctx, req.GetUserId()); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to delete user: %v", err)
	}
	return &pb.DeleteUserResponse{}, nil
}

// GetDiscoveryStatus returns the current status of all auto-discovery providers.
// Summary: Retrieves operational metrics and status for discovery providers.
// Parameters:
//   - ctx (context.Context): The context for the request.
//   - _ (*pb.GetDiscoveryStatusRequest): The request parameters (unused).
//
// Returns:
//   - *pb.GetDiscoveryStatusResponse: The status of all discovery providers.
//   - error: An error if retrieving the status fails.
//
// Errors:
//   - None.
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

// ListAuditLogs returns audit logs matching the specified filter criteria.
// Summary: Queries and returns historical audit log entries.
// Parameters:
//   - ctx (context.Context): The context for the request.
//   - req (*pb.ListAuditLogsRequest): The filter criteria including time range and identifiers.
//
// Returns:
//   - *pb.ListAuditLogsResponse: The matching audit log entries.
//   - error: An error if audit logging is disabled or the query fails.
//
// Errors:
//   - FailedPrecondition: If audit logging is not enabled.
//   - InvalidArgument: If start_time or end_time are invalid timestamps.
//   - Internal: If the storage operation fails.
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
