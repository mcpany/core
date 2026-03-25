// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

// RegistrationServer implements the gRPC server for service registration.
//
// Summary: Handles gRPC requests for registering and managing upstream services.
//
// Side Effects:
//   - Publishes messages to the event bus.
// NewRegistrationServer creates a new RegistrationServer initialized with the event bus and auth manager.
//
// Summary: Initializes a new RegistrationServer instance.
//
// Parameters:
//   - bus: *bus.Provider. The event bus used for communication with workers.
//   - authManager: *auth.Manager. Manager for handling authentication and OAuth flows.
//
// Returns:
//   - *RegistrationServer: A new instance of the RegistrationServer.
// ValidateService validates a service configuration by attempting to connect and discover tools.
//
// Summary: Validates the provided service configuration by connecting to the upstream service.
// Errors:
//   - None.
//
// Errors:
//   - triggers relevant error states on failure.
// Side Effects:
//   - updates relevant subsystem state or network conditions.
// Parameters:
//   - ctx: context.Context. The context for the request.
//   - req: *v1.ValidateServiceRequest. The validation request containing the service configuration.
//
// Returns:
//   - *v1.ValidateServiceResponse: The response containing validation results, discovered tools, and resources.
//   - error: An error if the validation request itself is invalid (e.g. missing config).
//
// Side Effects:
//   - Temporarily creates an upstream connection and then closes it.
// Errors:
//   - triggers relevant error states on failure.
package mcpserver

package mcpserver

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

// RegisterService handles a gRPC request to register a new upstream service.
//
// Summary: Asynchronously registers a new upstream service via the event bus.
//
// Parameters:
//   - ctx: context.Context. The context for the gRPC call.
//   - req: *v1.RegisterServiceRequest. The request containing the configuration of the service to be registered.
//
// Returns:
//   - *v1.RegisterServiceResponse: The response containing the registration status, service key, and discovered tools.
//   - error: An error if the registration fails, times out, or arguments are invalid.
//
// Side Effects:
//   - Publishes a registration request to the event bus.
//   - Waits for a response on a dedicated result channel.
// Errors:
//   - triggers relevant error states on failure.
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
// InitiateOAuth2Flow initiates an OAuth2 flow for a service or credential.
//
// Summary: Initiates the OAuth2 flow by generating an authorization URL.
//
// Parameters:
//   - ctx: context.Context. The context for the gRPC call.
//   - req: *v1.InitiateOAuth2FlowRequest. The request containing OAuth2 flow details.
//
// Returns:
//   - *v1.InitiateOAuth2FlowResponse: The response containing the authorization URL and state.
//   - error: An error if validation fails or flow initiation errors.
//
// Throws/Errors:
// Side Effects:
//   - updates relevant subsystem state or network conditions.
//   - codes.InvalidArgument: If required parameters are missing.
//   - codes.Unauthenticated: If the user is not authenticated.
//   - codes.Internal: If an internal error occurs.
// Side Effects:
//   - updates relevant subsystem state or network conditions.
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
// GetService retrieves a service by its name.
//
// Summary: Retrieves the configuration of a registered service.
//
// Parameters:
//   - ctx: context.Context. The context for the gRPC call.
//   - req: *v1.GetServiceRequest. The request containing the service name.
//
// Returns:
//   - *v1.GetServiceResponse: The response containing the service configuration.
//   - error: An error if the service is not found or other error.
// Errors:
//   - triggers relevant error states on failure.
// Side Effects:
//   - updates relevant subsystem state or network conditions.
//
// Throws/Errors:
// Side Effects:
//   - updates relevant subsystem state or network conditions.
//   - codes.InvalidArgument: If the service name is missing.
//   - codes.NotFound: If the service is not found.
//   - codes.DeadlineExceeded: If the request times out.
// Side Effects:
//   - updates relevant subsystem state or network conditions.
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
// ListServices lists all registered services by querying the service registry via the event bus.
//
// Summary: Lists all currently registered services.
//
// Parameters:
//   - ctx: context.Context. The context for the gRPC call.
//   - req: *v1.ListServicesRequest. The request object (empty for now).
//
// Returns:
//   - *v1.ListServicesResponse: The response containing a list of registered services.
//   - error: An error if the operation fails or times out.
//
// Throws/Errors:
//   - codes.DeadlineExceeded: If the request times out.
//   - codes.Internal: If an internal error occurs.
// Side Effects:
//   - updates relevant subsystem state or network conditions.
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
