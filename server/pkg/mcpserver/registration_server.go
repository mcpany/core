// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

// Package mcpserver implements the MCP server functionality.
package mcpserver

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	v1 "github.com/mcpany/core/proto/api/v1"
	"github.com/mcpany/core/server/pkg/auth"
	"github.com/mcpany/core/server/pkg/bus"
	"github.com/mcpany/core/server/pkg/config"
	"github.com/mcpany/core/server/pkg/logging"
	"github.com/mcpany/core/server/pkg/pool"
	"github.com/mcpany/core/server/pkg/upstream/factory"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Summary: RegistrationServer implements the gRPC server for service registration. Handles gRPC requests for registering and managing upstream services.
//
// Parameters:
//   - None.
//
// Returns:
//   - None.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
type RegistrationServer struct {
	v1.UnimplementedRegistrationServiceServer
	bus         *bus.Provider
	authManager *auth.Manager
}

// Summary: NewRegistrationServerHook is a test hook for overriding the creation of a RegistrationServer. Test hook to override RegistrationServer creation.
//
// Parameters:
//   - None.
//
// Returns:
//   - None.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
var NewRegistrationServerHook func(bus interface{}, authManager interface{}) (*RegistrationServer, error)

// Summary: NewRegistrationServer creates a new RegistrationServer initialized with the event bus and auth manager. Initializes a new RegistrationServer instance.
//
// Parameters:
//   - bus (*bus.Provider): The bus parameter.
//   - authManager (*auth.Manager): The authManager parameter.
//
// Returns:
//   - *RegistrationServer: The resulting *RegistrationServer.
//   - error: An error if the operation fails.
//
// Errors:
//   - Returns an error if the operation fails or is invalid.
//
// Side Effects:
//   - None.
func NewRegistrationServer(bus *bus.Provider, authManager *auth.Manager) (*RegistrationServer, error) {
	if NewRegistrationServerHook != nil {
		// The type assertion is safe because this is a test-only hook.
		return NewRegistrationServerHook(bus, authManager)
	}
	if bus == nil {
		return nil, fmt.Errorf("bus is nil")
	}
	return &RegistrationServer{bus: bus, authManager: authManager}, nil
}

// Summary: ValidateService validates a service configuration by attempting to connect and discover tools. Validates the provided service configuration by connecting to the upstream service.
//
// Parameters:
//   - ctx (context.Context): The ctx parameter.
//   - req (*v1.ValidateServiceRequest): The req parameter.
//
// Returns:
//   - *v1.ValidateServiceResponse: The resulting *v1.ValidateServiceResponse.
//   - error: An error if the operation fails.
//
// Errors:
//   - Returns an error if the operation fails or is invalid.
//
// Side Effects:
//   - None.
func (s *RegistrationServer) ValidateService(ctx context.Context, req *v1.ValidateServiceRequest) (*v1.ValidateServiceResponse, error) {
	if req.GetConfig() == nil {
		return nil, status.Errorf(codes.InvalidArgument, "config is required")
	}

	// Validate config syntax
	if err := config.ValidateOrError(ctx, req.GetConfig()); err != nil {
		return v1.ValidateServiceResponse_builder{
			Valid:   false,
			Message: fmt.Sprintf("Invalid configuration: %v", err),
		}.Build(), nil
	}

	// Create temporary factory
	// We need a pool manager. Since we close the upstream immediately, a local pool manager is fine.
	poolManager := pool.NewManager()
	defer poolManager.CloseAll()

	// Use global settings for factory
	upstreamFactory := factory.NewUpstreamServiceFactory(poolManager, config.GlobalSettings().ToProto())

	u, err := upstreamFactory.NewUpstream(req.GetConfig())
	if err != nil {
		return v1.ValidateServiceResponse_builder{
			Valid:   false,
			Message: fmt.Sprintf("Failed to create upstream: %v", err),
		}.Build(), nil
	}
	defer func() {
		_ = u.Shutdown(ctx)
	}()

	// Use NoOp managers (except toolManager which needs to store ServiceInfo temporarily)
	toolManager := NewTemporaryToolManager()
	promptManager := &NoOpPromptManager{}
	resourceManager := &NoOpResourceManager{}

	_, tools, resources, err := u.Register(ctx, req.GetConfig(), toolManager, promptManager, resourceManager, false)
	if err != nil {
		return v1.ValidateServiceResponse_builder{
			Valid:   false,
			Message: fmt.Sprintf("Validation failed: %v", err),
		}.Build(), nil
	}

	return v1.ValidateServiceResponse_builder{
		Valid:               true,
		Message:             "Service configuration is valid and reachable.",
		DiscoveredTools:     tools,
		DiscoveredResources: resources,
	}.Build(), nil
}

// Summary: RegisterService handles a gRPC request to register a new upstream service. Asynchronously registers a new upstream service via the event bus.
//
// Parameters:
//   - ctx (context.Context): The ctx parameter.
//   - req (*v1.RegisterServiceRequest): The req parameter.
//
// Returns:
//   - *v1.RegisterServiceResponse: The resulting *v1.RegisterServiceResponse.
//   - error: An error if the operation fails.
//
// Errors:
//   - Returns an error if the operation fails or is invalid.
//
// Side Effects:
//   - None.
func (s *RegistrationServer) RegisterService(ctx context.Context, req *v1.RegisterServiceRequest) (*v1.RegisterServiceResponse, error) {
	if req.GetConfig() == nil {
		return nil, status.Errorf(codes.InvalidArgument, "config is required")
	}
	if req.GetConfig().GetName() == "" {
		return nil, status.Errorf(codes.InvalidArgument, "config.name is required")
	}

	if err := config.ValidateOrError(ctx, req.GetConfig()); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid config: %v", err)
	}

	correlationID := uuid.New().String()
	resultChan := make(chan *bus.ServiceRegistrationResult, 1)

	resultBus, _ := bus.GetBus[*bus.ServiceRegistrationResult](s.bus, "service_registration_results")
	unsubscribe := resultBus.SubscribeOnce(ctx, correlationID, func(result *bus.ServiceRegistrationResult) {
		resultChan <- result
	})
	defer unsubscribe()

	requestBus, _ := bus.GetBus[*bus.ServiceRegistrationRequest](s.bus, "service_registration_requests")
	regReq := &bus.ServiceRegistrationRequest{
		Config: req.GetConfig(),
	}
	regReq.SetCorrelationID(correlationID)
	_ = requestBus.Publish(ctx, "request", regReq)

	// Wait for the result, respecting the context's deadline
	select {
	case result := <-resultChan:
		if result.Error != nil {
			return nil, status.Errorf(codes.Internal, "failed to register service: %v", result.Error)
		}

		log := logging.GetLogger()
		log.InfoContext(ctx, "Service registered via bus",
			"service_name", req.GetConfig().GetName(),
			"service_key", result.ServiceKey,
			"discovered_tools_count", len(result.DiscoveredTools),
			"discovered_resources_count", len(result.DiscoveredResources),
		)
		if len(result.DiscoveredTools) > 0 {
			for _, tool := range result.DiscoveredTools {
				log.InfoContext(ctx, "Discovered tool", "tool", tool)
			}
		}
		if len(result.DiscoveredResources) > 0 {
			for _, resource := range result.DiscoveredResources {
				log.InfoContext(ctx, "Discovered resource", "resource", resource)
			}
		}

		msg := fmt.Sprintf("service %s registered successfully with key %s", req.GetConfig().GetName(), result.ServiceKey)
		resp := v1.RegisterServiceResponse_builder{}.Build()
		resp.SetMessage(msg)
		resp.SetDiscoveredTools(result.DiscoveredTools)
		resp.SetServiceKey(result.ServiceKey)
		resp.SetDiscoveredResources(result.DiscoveredResources)
		return resp, nil
	case <-ctx.Done():
		return nil, status.Errorf(codes.DeadlineExceeded, "context deadline exceeded while waiting for service registration")
	case <-time.After(300 * time.Second): // Add a safety timeout
		return nil, status.Errorf(codes.DeadlineExceeded, "timed out waiting for service registration result")
	}
}

// Summary: UnregisterService is not yet implemented. Handles the unregistration of a service (Not Implemented).
//
// Parameters:
//   - _ (context.Context): The _ parameter.
//   - _ (*v1.UnregisterServiceRequest): The _ parameter.
//
// Returns:
//   - *v1.UnregisterServiceResponse: The resulting *v1.UnregisterServiceResponse.
//   - error: An error if the operation fails.
//
// Errors:
//   - Returns an error if the operation fails or is invalid.
//
// Side Effects:
//   - None.
func (s *RegistrationServer) UnregisterService(_ context.Context, _ *v1.UnregisterServiceRequest) (*v1.UnregisterServiceResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UnregisterService not implemented")
}

// Summary: InitiateOAuth2Flow initiates an OAuth2 flow for a service or credential. Initiates the OAuth2 flow by generating an authorization URL.
//
// Parameters:
//   - ctx (context.Context): The ctx parameter.
//   - req (*v1.InitiateOAuth2FlowRequest): The req parameter.
//
// Returns:
//   - *v1.InitiateOAuth2FlowResponse: The resulting *v1.InitiateOAuth2FlowResponse.
//   - error: An error if the operation fails.
//
// Errors:
//   - Returns an error if the operation fails or is invalid.
//
// Side Effects:
//   - None.
func (s *RegistrationServer) InitiateOAuth2Flow(ctx context.Context, req *v1.InitiateOAuth2FlowRequest) (*v1.InitiateOAuth2FlowResponse, error) {
	if req.GetServiceId() == "" && req.GetCredentialId() == "" {
		return nil, status.Error(codes.InvalidArgument, "either service_id or credential_id is required")
	}
	if req.GetRedirectUrl() == "" {
		return nil, status.Error(codes.InvalidArgument, "redirect_url is required")
	}

	uid, ok := auth.UserFromContext(ctx)
	if !ok {
		// Ideally, we require authentication to initiate oauth.
		// If the context doesn't have it (e.g. CLI without user context?), what do we do?
		// We might default to a system or verify if the request is allowed.
		// For now, assume context has it or fail?
		// gRPC interceptors should populate it if auth is enabled.
		return nil, status.Error(codes.Unauthenticated, "user not authenticated")
	}

	url, state, err := s.authManager.InitiateOAuth(ctx, uid, req.GetServiceId(), req.GetCredentialId(), req.GetRedirectUrl())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to initiate oauth: %v", err)
	}

	return v1.InitiateOAuth2FlowResponse_builder{
		AuthorizationUrl: url,
		State:            state,
	}.Build(), nil
}

// Summary: RegisterTools is not yet implemented. Registers tools for a service (Not Implemented).
//
// Parameters:
//   - _ (context.Context): The _ parameter.
//   - _ (*v1.RegisterToolsRequest): The _ parameter.
//
// Returns:
//   - *v1.RegisterToolsResponse: The resulting *v1.RegisterToolsResponse.
//   - error: An error if the operation fails.
//
// Errors:
//   - Returns an error if the operation fails or is invalid.
//
// Side Effects:
//   - None.
func (s *RegistrationServer) RegisterTools(_ context.Context, _ *v1.RegisterToolsRequest) (*v1.RegisterToolsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method RegisterTools not implemented")
}

// Summary: GetServiceStatus is not yet implemented. Retrieves the status of a service (Not Implemented).
//
// Parameters:
//   - _ (context.Context): The _ parameter.
//   - _ (*v1.GetServiceStatusRequest): The _ parameter.
//
// Returns:
//   - *v1.GetServiceStatusResponse: The resulting *v1.GetServiceStatusResponse.
//   - error: An error if the operation fails.
//
// Errors:
//   - Returns an error if the operation fails or is invalid.
//
// Side Effects:
//   - None.
func (s *RegistrationServer) GetServiceStatus(_ context.Context, _ *v1.GetServiceStatusRequest) (*v1.GetServiceStatusResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetServiceStatus not implemented")
}

// Summary: GetService retrieves a service by its name. Retrieves the configuration of a registered service.
//
// Parameters:
//   - ctx (context.Context): The ctx parameter.
//   - req (*v1.GetServiceRequest): The req parameter.
//
// Returns:
//   - *v1.GetServiceResponse: The resulting *v1.GetServiceResponse.
//   - error: An error if the operation fails.
//
// Errors:
//   - Returns an error if the operation fails or is invalid.
//
// Side Effects:
//   - None.
func (s *RegistrationServer) GetService(ctx context.Context, req *v1.GetServiceRequest) (*v1.GetServiceResponse, error) {
	if req.GetServiceName() == "" {
		return nil, status.Errorf(codes.InvalidArgument, "service_name is required")
	}

	correlationID := uuid.New().String()
	resultChan := make(chan *bus.ServiceGetResult, 1)

	resultBus, _ := bus.GetBus[*bus.ServiceGetResult](s.bus, "service_get_results")
	unsubscribe := resultBus.SubscribeOnce(ctx, correlationID, func(result *bus.ServiceGetResult) {
		resultChan <- result
	})
	defer unsubscribe()

	requestBus, _ := bus.GetBus[*bus.ServiceGetRequest](s.bus, "service_get_requests")
	getReq := &bus.ServiceGetRequest{
		ServiceName: req.GetServiceName(),
	}
	getReq.SetCorrelationID(correlationID)
	_ = requestBus.Publish(ctx, "request", getReq)

	select {
	case result := <-resultChan:
		if result.Error != nil {
			return nil, status.Errorf(codes.NotFound, "failed to get service: %v", result.Error)
		}
		if result.Service == nil {
			return nil, status.Errorf(codes.NotFound, "service not found")
		}
		resp := v1.GetServiceResponse_builder{}.Build()
		resp.SetService(result.Service)
		return resp, nil
	case <-ctx.Done():
		return nil, status.Errorf(codes.DeadlineExceeded, "context deadline exceeded while waiting for service")
	case <-time.After(30 * time.Second):
		return nil, status.Errorf(codes.DeadlineExceeded, "timed out waiting for service get result")
	}
}

func (s *RegistrationServer) mustEmbedUnimplementedRegistrationServiceServer() {} //nolint:unused

// Summary: ListServices lists all registered services by querying the service registry via the event bus. Lists all currently registered services.
//
// Parameters:
//   - ctx (context.Context): The ctx parameter.
//   - _ (*v1.ListServicesRequest): The _ parameter.
//
// Returns:
//   - *v1.ListServicesResponse: The resulting *v1.ListServicesResponse.
//   - error: An error if the operation fails.
//
// Errors:
//   - Returns an error if the operation fails or is invalid.
//
// Side Effects:
//   - None.
func (s *RegistrationServer) ListServices(ctx context.Context, _ *v1.ListServicesRequest) (*v1.ListServicesResponse, error) {
	correlationID := uuid.New().String()
	resultChan := make(chan *bus.ServiceListResult, 1)

	resultBus, _ := bus.GetBus[*bus.ServiceListResult](s.bus, "service_list_results")
	unsubscribe := resultBus.SubscribeOnce(ctx, correlationID, func(result *bus.ServiceListResult) {
		resultChan <- result
	})
	defer unsubscribe()

	requestBus, _ := bus.GetBus[*bus.ServiceListRequest](s.bus, "service_list_requests")
	listReq := &bus.ServiceListRequest{}
	listReq.SetCorrelationID(correlationID)
	_ = requestBus.Publish(ctx, "request", listReq)

	select {
	case result := <-resultChan:
		if result.Error != nil {
			return nil, status.Errorf(codes.Internal, "failed to list services: %v", result.Error)
		}
		resp := v1.ListServicesResponse_builder{}.Build()
		resp.SetServices(result.Services)
		return resp, nil
	case <-ctx.Done():
		return nil, status.Errorf(codes.DeadlineExceeded, "context deadline exceeded while waiting for service list")
	case <-time.After(30 * time.Second):
		return nil, status.Errorf(codes.DeadlineExceeded, "timed out waiting for service list result")
	}
}
