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
// Summary: Server implements the AdminServiceServer interface.
//
// Summary: Server implements the AdminServiceServer interface.
type Server struct {
	pb.UnimplementedAdminServiceServer
	cache            *middleware.CachingMiddleware
	toolManager      tool.ManagerInterface
	serviceRegistry  serviceregistry.ServiceRegistryInterface
	storage          storage.Storage
	discoveryManager *discovery.Manager
	auditMiddleware  *middleware.AuditMiddleware
// NewServer creates a new Admin Server. cache manages the caching layer. toolManager is the toolManager. serviceRegistry is the registry of upstream services. storage provides the persistence layer. discoveryManager manages auto-discovery. auditMiddleware provides access to audit logs. Returns the result.
//
// Summary: NewServer creates a new Admin Server. cache manages the caching layer. toolManager is the toolManager. serviceRegistry is the registry of upstream services. storage provides the persistence layer. discoveryManager manages auto-discovery. auditMiddleware provides access to audit logs. Returns the result.
//
// Parameters:
//   - cache (*middleware.CachingMiddleware): The provided cache data.
//   - toolManager (tool.ManagerInterface): The provided toolmanager data.
//   - serviceRegistry (serviceregistry.ServiceRegistryInterface): The provided serviceregistry data.
//   - storage (storage.Storage): The provided storage data.
//   - discoveryManager (*discovery.Manager): The provided discoverymanager data.
//   - auditMiddleware (*middleware.AuditMiddleware): The provided auditmiddleware data.
//
// Returns:
//   - *Server: The resulting object or data structure.
//
// Errors:
//   - None.
//
// Side Effects:
//   - May modify internal state or perform external network calls.
//   - None.
//
// Side Effects:
//   - May modify internal state or perform external network calls.
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
// ClearCache clears the cache. ctx is the context for the request. _ is an unused parameter. Returns the response. Returns an error if the operation fails.
//
// Summary: ClearCache clears the cache. ctx is the context for the request. _ is an unused parameter. Returns the response. Returns an error if the operation fails.
//
// Parameters:
//   - ctx (context.Context): The cancellation and deadline context.
//   - _ (*pb.ClearCacheRequest): The provided _ data.
//
// Returns:
//   - *pb.ClearCacheResponse: The resulting object or data structure.
//   - error: An error if the execution fails, otherwise nil.
//
// Errors:
//   - Returns an error if the operation fails, invalid input is provided, or a downstream dependency fails.
//
// Side Effects:
//   - May modify internal state or perform external network calls.
//
// Errors:
//   - Returns an error if the operation fails, invalid input is provided, or a downstream dependency fails.
//
// Side Effects:
//   - May modify internal state or perform external network calls.
func (s *Server) ClearCache(ctx context.Context, _ *pb.ClearCacheRequest) (*pb.ClearCacheResponse, error) {
	if s.cache == nil {
		return nil, status.Error(codes.FailedPrecondition, "caching is not enabled")
	}
// ListServices returns all registered services. _ is an unused parameter. _ is an unused parameter. Returns the response. Returns an error if the operation fails.
//
// Summary: ListServices returns all registered services. _ is an unused parameter. _ is an unused parameter. Returns the response. Returns an error if the operation fails.
//
// Parameters:
//   - _ (context.Context): The provided _ data.
//   - _ (*pb.ListServicesRequest): The provided _ data.
//
// Returns:
//   - *pb.ListServicesResponse: The resulting object or data structure.
//   - error: An error if the execution fails, otherwise nil.
//
// Errors:
//   - Returns an error if the operation fails, invalid input is provided, or a downstream dependency fails.
//
// Side Effects:
//   - May modify internal state or perform external network calls.
//   - *pb.ListServicesResponse: The resulting object or data structure.
//   - error: An error if the execution fails, otherwise nil.
//
// Errors:
//   - Returns an error if the operation fails, invalid input is provided, or a downstream dependency fails.
//
// Side Effects:
//   - May modify internal state or perform external network calls.
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
// GetService returns a specific service by ID. _ is an unused parameter. req is the request object. Returns the response. Returns an error if the operation fails.
//
// Summary: GetService returns a specific service by ID. _ is an unused parameter. req is the request object. Returns the response. Returns an error if the operation fails.
//
// Parameters:
//   - _ (context.Context): The provided _ data.
//   - req (*pb.GetServiceRequest): The incoming request payload.
//
// Returns:
//   - *pb.GetServiceResponse: The resulting object or data structure.
//   - error: An error if the execution fails, otherwise nil.
//
// Errors:
//   - Returns an error if the operation fails, invalid input is provided, or a downstream dependency fails.
//
// Side Effects:
//   - May modify internal state or perform external network calls.
//
// Returns:
//   - *pb.GetServiceResponse: The resulting object or data structure.
//   - error: An error if the execution fails, otherwise nil.
//
// Errors:
//   - Returns an error if the operation fails, invalid input is provided, or a downstream dependency fails.
//
// Side Effects:
//   - May modify internal state or perform external network calls.
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
// ListTools returns all registered tools. _ is an unused parameter. _ is an unused parameter. Returns the response. Returns an error if the operation fails.
//
// Summary: ListTools returns all registered tools. _ is an unused parameter. _ is an unused parameter. Returns the response. Returns an error if the operation fails.
//
// Parameters:
//   - _ (context.Context): The provided _ data.
//   - _ (*pb.ListToolsRequest): The provided _ data.
//
// Returns:
//   - *pb.ListToolsResponse: The resulting object or data structure.
//   - error: An error if the execution fails, otherwise nil.
//
// Errors:
//   - Returns an error if the operation fails, invalid input is provided, or a downstream dependency fails.
//
// Side Effects:
//   - May modify internal state or perform external network calls.
//   - _ (context.Context): The provided _ data.
//   - _ (*pb.ListToolsRequest): The provided _ data.
//
// Returns:
//   - *pb.ListToolsResponse: The resulting object or data structure.
//   - error: An error if the execution fails, otherwise nil.
//
// Errors:
//   - Returns an error if the operation fails, invalid input is provided, or a downstream dependency fails.
// GetTool returns a specific tool by name. _ is an unused parameter. req is the request object. Returns the response. Returns an error if the operation fails.
//
// Summary: GetTool returns a specific tool by name. _ is an unused parameter. req is the request object. Returns the response. Returns an error if the operation fails.
//
// Parameters:
//   - _ (context.Context): The provided _ data.
//   - req (*pb.GetToolRequest): The incoming request payload.
//
// Returns:
//   - *pb.GetToolResponse: The resulting object or data structure.
//   - error: An error if the execution fails, otherwise nil.
//
// Errors:
//   - Returns an error if the operation fails, invalid input is provided, or a downstream dependency fails.
//
// Side Effects:
//   - May modify internal state or perform external network calls.
//
// Parameters:
//   - _ (context.Context): The provided _ data.
//   - req (*pb.GetToolRequest): The incoming request payload.
//
// Returns:
//   - *pb.GetToolResponse: The resulting object or data structure.
//   - error: An error if the execution fails, otherwise nil.
// CreateUser creates a new user. ctx is the context for the request. req is the request object. Returns the response. Returns an error if the operation fails.
//
// Summary: CreateUser creates a new user. ctx is the context for the request. req is the request object. Returns the response. Returns an error if the operation fails.
//
// Parameters:
//   - ctx (context.Context): The cancellation and deadline context.
//   - req (*pb.CreateUserRequest): The incoming request payload.
//
// Returns:
//   - *pb.CreateUserResponse: The resulting object or data structure.
//   - error: An error if the execution fails, otherwise nil.
//
// Errors:
//   - Returns an error if the operation fails, invalid input is provided, or a downstream dependency fails.
//
// Side Effects:
//   - May modify internal state or perform external network calls.
//
// Summary: CreateUser creates a new user. ctx is the context for the request. req is the request object. Returns the response. Returns an error if the operation fails.
//
// Parameters:
//   - ctx (context.Context): The cancellation and deadline context.
//   - req (*pb.CreateUserRequest): The incoming request payload.
//
// Returns:
//   - *pb.CreateUserResponse: The resulting object or data structure.
//   - error: An error if the execution fails, otherwise nil.
//
// Errors:
//   - Returns an error if the operation fails, invalid input is provided, or a downstream dependency fails.
//
// Side Effects:
//   - May modify internal state or perform external network calls.
func (s *Server) CreateUser(ctx context.Context, req *pb.CreateUserRequest) (*pb.CreateUserResponse, error) {
	if !req.HasUser() {
		return nil, status.Error(codes.InvalidArgument, "user is required")
	}
	// Hash password if needed
	if req.GetUser().HasAuthentication() {
		if basic := req.GetUser().GetAuthentication().GetBasicAuth(); basic != nil {
			if basic.GetPasswordHash() != "" && !strings.HasPrefix(basic.GetPasswordHash(), "$2") {
				hashed, err := passhash.Password(basic.GetPasswordHash())
// GetUser retrieves a user by ID. ctx is the context for the request. req is the request object. Returns the response. Returns an error if the operation fails.
//
// Summary: GetUser retrieves a user by ID. ctx is the context for the request. req is the request object. Returns the response. Returns an error if the operation fails.
//
// Parameters:
//   - ctx (context.Context): The cancellation and deadline context.
//   - req (*pb.GetUserRequest): The incoming request payload.
//
// Returns:
//   - *pb.GetUserResponse: The resulting object or data structure.
//   - error: An error if the execution fails, otherwise nil.
//
// Errors:
//   - Returns an error if the operation fails, invalid input is provided, or a downstream dependency fails.
//
// Side Effects:
//   - May modify internal state or perform external network calls.

// GetUser retrieves a user by ID. ctx is the context for the request. req is the request object. Returns the response. Returns an error if the operation fails.
//
// Summary: GetUser retrieves a user by ID. ctx is the context for the request. req is the request object. Returns the response. Returns an error if the operation fails.
//
// Parameters:
//   - ctx (context.Context): The cancellation and deadline context.
//   - req (*pb.GetUserRequest): The incoming request payload.
//
// Returns:
//   - *pb.GetUserResponse: The resulting object or data structure.
//   - error: An error if the execution fails, otherwise nil.
//
// Errors:
// ListUsers lists all users. ctx is the context for the request. _ is an unused parameter. Returns the response. Returns an error if the operation fails.
//
// Summary: ListUsers lists all users. ctx is the context for the request. _ is an unused parameter. Returns the response. Returns an error if the operation fails.
//
// Parameters:
//   - ctx (context.Context): The cancellation and deadline context.
//   - _ (*pb.ListUsersRequest): The provided _ data.
//
// Returns:
//   - *pb.ListUsersResponse: The resulting object or data structure.
//   - error: An error if the execution fails, otherwise nil.
//
// Errors:
//   - Returns an error if the operation fails, invalid input is provided, or a downstream dependency fails.
//
// Side Effects:
//   - May modify internal state or perform external network calls.
	return pb.GetUserResponse_builder{User: safeUser}.Build(), nil
}

// ListUsers lists all users. ctx is the context for the request. _ is an unused parameter. Returns the response. Returns an error if the operation fails.
//
// Summary: ListUsers lists all users. ctx is the context for the request. _ is an unused parameter. Returns the response. Returns an error if the operation fails.
//
// Parameters:
//   - ctx (context.Context): The cancellation and deadline context.
//   - _ (*pb.ListUsersRequest): The provided _ data.
//
// Returns:
//   - *pb.ListUsersResponse: The resulting object or data structure.
//   - error: An error if the execution fails, otherwise nil.
//
// Errors:
// UpdateUser updates an existing user. ctx is the context for the request. req is the request object. Returns the response. Returns an error if the operation fails.
//
// Summary: UpdateUser updates an existing user. ctx is the context for the request. req is the request object. Returns the response. Returns an error if the operation fails.
//
// Parameters:
//   - ctx (context.Context): The cancellation and deadline context.
//   - req (*pb.UpdateUserRequest): The incoming request payload.
//
// Returns:
//   - *pb.UpdateUserResponse: The resulting object or data structure.
//   - error: An error if the execution fails, otherwise nil.
//
// Errors:
//   - Returns an error if the operation fails, invalid input is provided, or a downstream dependency fails.
//
// Side Effects:
//   - May modify internal state or perform external network calls.
	}

	return pb.ListUsersResponse_builder{Users: safeUsers}.Build(), nil
}

// UpdateUser updates an existing user. ctx is the context for the request. req is the request object. Returns the response. Returns an error if the operation fails.
//
// Summary: UpdateUser updates an existing user. ctx is the context for the request. req is the request object. Returns the response. Returns an error if the operation fails.
//
// Parameters:
//   - ctx (context.Context): The cancellation and deadline context.
//   - req (*pb.UpdateUserRequest): The incoming request payload.
//
// Returns:
//   - *pb.UpdateUserResponse: The resulting object or data structure.
//   - error: An error if the execution fails, otherwise nil.
//
// Errors:
//   - Returns an error if the operation fails, invalid input is provided, or a downstream dependency fails.
//
// Side Effects:
//   - May modify internal state or perform external network calls.
func (s *Server) UpdateUser(ctx context.Context, req *pb.UpdateUserRequest) (*pb.UpdateUserResponse, error) {
	if !req.HasUser() {
		return nil, status.Error(codes.InvalidArgument, "user is required")
// DeleteUser deletes a user by ID. ctx is the context for the request. req is the request object. Returns the response. Returns an error if the operation fails.
//
// Summary: DeleteUser deletes a user by ID. ctx is the context for the request. req is the request object. Returns the response. Returns an error if the operation fails.
//
// Parameters:
//   - ctx (context.Context): The cancellation and deadline context.
//   - req (*pb.DeleteUserRequest): The incoming request payload.
//
// Returns:
//   - *pb.DeleteUserResponse: The resulting object or data structure.
//   - error: An error if the execution fails, otherwise nil.
//
// Errors:
//   - Returns an error if the operation fails, invalid input is provided, or a downstream dependency fails.
//
// Side Effects:
//   - May modify internal state or perform external network calls.
	}

	safeUser := proto.Clone(req.GetUser()).(*configv1.User)
	config.StripSecretsFromAuth(safeUser.GetAuthentication())
	return pb.UpdateUserResponse_builder{User: safeUser}.Build(), nil
}

// GetDiscoveryStatus returns the status of auto-discovery providers.
//
// Summary: GetDiscoveryStatus returns the status of auto-discovery providers.
//
// Parameters:
//   - _ (context.Context): The provided _ data.
//   - _ (*pb.GetDiscoveryStatusRequest): The provided _ data.
//
// Returns:
//   - *pb.GetDiscoveryStatusResponse: The resulting object or data structure.
//   - error: An error if the execution fails, otherwise nil.
//
// Errors:
//   - Returns an error if the operation fails, invalid input is provided, or a downstream dependency fails.
//
// Side Effects:
//   - May modify internal state or perform external network calls.
// Side Effects:
//   - May modify internal state or perform external network calls.
func (s *Server) DeleteUser(ctx context.Context, req *pb.DeleteUserRequest) (*pb.DeleteUserResponse, error) {
	if err := s.storage.DeleteUser(ctx, req.GetUserId()); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to delete user: %v", err)
	}
	return &pb.DeleteUserResponse{}, nil
}

// GetDiscoveryStatus returns the status of auto-discovery providers.
//
// Summary: GetDiscoveryStatus returns the status of auto-discovery providers.
//
// Parameters:
//   - _ (context.Context): The provided _ data.
//   - _ (*pb.GetDiscoveryStatusRequest): The provided _ data.
//
// Returns:
//   - *pb.GetDiscoveryStatusResponse: The resulting object or data structure.
//   - error: An error if the execution fails, otherwise nil.
//
// Errors:
// ListAuditLogs returns audit logs matching the filter.
//
// Summary: ListAuditLogs returns audit logs matching the filter.
//
// Parameters:
//   - ctx (context.Context): The cancellation and deadline context.
//   - req (*pb.ListAuditLogsRequest): The incoming request payload.
//
// Returns:
//   - *pb.ListAuditLogsResponse: The resulting object or data structure.
//   - error: An error if the execution fails, otherwise nil.
//
// Errors:
//   - Returns an error if the operation fails, invalid input is provided, or a downstream dependency fails.
//
// Side Effects:
//   - May modify internal state or perform external network calls.
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
// Summary: ListAuditLogs returns audit logs matching the filter.
//
// Parameters:
//   - ctx (context.Context): The cancellation and deadline context.
//   - req (*pb.ListAuditLogsRequest): The incoming request payload.
//
// Returns:
//   - *pb.ListAuditLogsResponse: The resulting object or data structure.
//   - error: An error if the execution fails, otherwise nil.
//
// Errors:
//   - Returns an error if the operation fails, invalid input is provided, or a downstream dependency fails.
//
// Side Effects:
//   - May modify internal state or perform external network calls.
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
