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
type Server struct {
	pb.UnimplementedAdminServiceServer
	cache            *middleware.CachingMiddleware
	toolManager      tool.ManagerInterface
	serviceRegistry  serviceregistry.ServiceRegistryInterface
	storage          storage.Storage
	discoveryManager *discovery.Manager
	auditMiddleware  *middleware.AuditMiddleware
}

// NewServer creates a new Admin Server. cache manages the caching layer. toolManager is the toolManager. serviceRegistry is the registry of upstream services. storage provides the persistence layer. discoveryManager manages auto-discovery. auditMiddleware provides access to audit logs. Returns the result.
//
// Parameters:
//   - cache (*middleware.CachingMiddleware): The cache parameter.
//   - toolManager (tool.ManagerInterface): The toolManager parameter.
//   - serviceRegistry (serviceregistry.ServiceRegistryInterface): The serviceRegistry parameter.
//   - storage (storage.Storage): The storage parameter.
//   - discoveryManager (*discovery.Manager): The discoveryManager parameter.
//   - auditMiddleware (*middleware.AuditMiddleware): The auditMiddleware parameter.
//
// Returns:
//   - *Server: The resulting *Server.
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

// ClearCache flushes the semantic cache, removing all stored responses.
//
// Parameters:
//   - ctx (context.Context): The request context used for execution timeout and tracing.
//   - _ (*pb.ClearCacheRequest): The empty request message.
//
// Returns:
//   - *pb.ClearCacheResponse: An empty response message indicating success.
//   - error: Returns a gRPC status error if the cache operation fails.
//
// Errors:
//   - Returns an "Unavailable" error if the cache middleware is not initialized or unreachable.
//
// Side Effects:
//   - Modifies global state by removing all entries from the semantic cache.
func (s *Server) ClearCache(ctx context.Context, _ *pb.ClearCacheRequest) (*pb.ClearCacheResponse, error) {
	if s.cache == nil {
		return nil, status.Error(codes.FailedPrecondition, "caching is not enabled")
	}
	if err := s.cache.Clear(ctx); err != nil {
		return nil, err
	}
	return &pb.ClearCacheResponse{}, nil
}

// ListServices retrieves a list of all upstream services currently managed by the Service Registry.
//
// Parameters:
//   - _ (context.Context): The unused request context.
//   - _ (*pb.ListServicesRequest): The empty request message.
//
// Returns:
//   - *pb.ListServicesResponse: A message containing an array of registered services and their current status.
//   - error: Returns an error if the registry cannot be queried.
//
// Errors:
//   - Returns an error if internal components fail to retrieve service definitions.
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

// GetService retrieves a specific upstream service's configuration and runtime details by its unique identifier.
//
// Parameters:
//   - _ (context.Context): The unused request context.
//   - req (*pb.GetServiceRequest): The request message containing the target service ID.
//
// Returns:
//   - *pb.GetServiceResponse: A message containing the service's details.
//   - error: Returns an error if the requested service does not exist.
//
// Errors:
//   - Returns a "NotFound" error if the service ID is not registered in the system.
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

// ListTools retrieves a comprehensive catalog of all executable tools available across all registered services.
//
// Parameters:
//   - _ (context.Context): The unused request context.
//   - _ (*pb.ListToolsRequest): The empty request message.
//
// Returns:
//   - *pb.ListToolsResponse: A message containing a list of all accessible tools.
//   - error: Returns an error if the tool catalog cannot be accessed.
//
// Errors:
//   - Returns an error if internal components fail to aggregate tool definitions.
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

// GetTool fetches the schema and execution definition for a specific tool by its registered name.
//
// Parameters:
//   - _ (context.Context): The unused request context.
//   - req (*pb.GetToolRequest): The request message containing the tool's name.
//
// Returns:
//   - *pb.GetToolResponse: A message containing the detailed tool definition.
//   - error: Returns an error if the tool is not found.
//
// Errors:
//   - Returns a "NotFound" error if the tool name does not exist in the active catalog.
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

// CreateUser provisions a new user account in the system database and handles secure password hashing.
//
// Parameters:
//   - ctx (context.Context): The request context used for database timeouts.
//   - req (*pb.CreateUserRequest): The request object containing the new user's profile and authentication details.
//
// Returns:
//   - *pb.CreateUserResponse: A message containing the created user record (with secrets stripped).
//   - error: Returns a gRPC status error on validation or storage failures.
//
// Errors:
//   - Returns an "InvalidArgument" error if the user object is missing.
//   - Returns an "Internal" error if password hashing fails or the database write encounters an issue.
//
// Side Effects:
//   - Writes a new user record to the persistent database.
//   - Modifies the request's password hash in memory during processing.
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

// GetUser retrieves an existing user's profile information by their unique ID, ensuring secrets are not exposed.
//
// Parameters:
//   - ctx (context.Context): The request context used for database timeouts.
//   - req (*pb.GetUserRequest): The request object containing the target user ID.
//
// Returns:
//   - *pb.GetUserResponse: A message containing the sanitized user details.
//   - error: Returns a gRPC status error if the user is missing or the database fails.
//
// Errors:
//   - Returns a "NotFound" error if the user ID does not match any existing records.
//   - Returns an "Internal" error if the database read operation fails.
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

// ListUsers retrieves all registered user accounts from the database and strips sensitive authentication data before returning.
//
// Parameters:
//   - ctx (context.Context): The request context used for database timeouts.
//   - _ (*pb.ListUsersRequest): The empty request message.
//
// Returns:
//   - *pb.ListUsersResponse: A message containing an array of sanitized user profiles.
//   - error: Returns a gRPC status error if the database query fails.
//
// Errors:
//   - Returns an "Internal" error if the system cannot retrieve the users list from storage.
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

// UpdateUser modifies an existing user's profile and securely updates their password hash if provided.
//
// Parameters:
//   - ctx (context.Context): The request context used for database timeouts.
//   - req (*pb.UpdateUserRequest): The request object containing the updated user details.
//
// Returns:
//   - *pb.UpdateUserResponse: A message containing the updated, sanitized user profile.
//   - error: Returns a gRPC status error if validation, hashing, or storage fails.
//
// Errors:
//   - Returns an "InvalidArgument" error if the user object is missing.
//   - Returns an "Internal" error if password hashing or the database update fails.
//
// Side Effects:
//   - Modifies an existing user record in the persistent database.
//   - Modifies the request's password hash in memory during processing.
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

// DeleteUser permanently removes a user account from the system database by their ID.
//
// Parameters:
//   - ctx (context.Context): The request context used for database timeouts.
//   - req (*pb.DeleteUserRequest): The request message containing the ID of the user to delete.
//
// Returns:
//   - *pb.DeleteUserResponse: An empty message indicating successful deletion.
//   - error: Returns a gRPC status error if the deletion fails.
//
// Errors:
//   - Returns an "Internal" error if the database encounters an issue while attempting to delete the record.
//
// Side Effects:
//   - Permanently deletes a user record from the persistent database.
func (s *Server) DeleteUser(ctx context.Context, req *pb.DeleteUserRequest) (*pb.DeleteUserResponse, error) {
	if err := s.storage.DeleteUser(ctx, req.GetUserId()); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to delete user: %v", err)
	}
	return &pb.DeleteUserResponse{}, nil
}

// GetDiscoveryStatus queries all configured auto-discovery providers and aggregates their current operational statuses.
//
// Parameters:
//   - _ (context.Context): The request context used for internal cancellation.
//   - _ (*pb.GetDiscoveryStatusRequest): The empty request message.
//
// Returns:
//   - *pb.GetDiscoveryStatusResponse: A message containing an array of provider statuses and discovery counts.
//   - error: Returns a gRPC status error if the discovery manager is unreachable.
//
// Errors:
//   - Returns an "Internal" error if status aggregation fails unexpectedly.
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

// ListAuditLogs fetches paginated and filtered audit log entries from the configured audit middleware storage.
//
// Parameters:
//   - ctx (context.Context): The request context used for executing the log query.
//   - req (*pb.ListAuditLogsRequest): The request object containing time bounds, filters, and pagination parameters.
//
// Returns:
//   - *pb.ListAuditLogsResponse: A message containing the requested array of formatted audit log entries.
//   - error: Returns a gRPC status error if the query cannot be executed.
//
// Errors:
//   - Returns a "FailedPrecondition" error if the audit middleware is not currently enabled.
//   - Returns an "InvalidArgument" error if the provided time boundaries are malformed.
//   - Returns an "Internal" error if the underlying log read operation fails.
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
