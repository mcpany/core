// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

// Package serviceregistry provides service registry functionality.
package serviceregistry

import (
	"context"
	"fmt"
	"sync"
	"time"

	config "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/auth"
	"github.com/mcpany/core/server/pkg/logging"
	"github.com/mcpany/core/server/pkg/prompt"
	"github.com/mcpany/core/server/pkg/resource"
	"github.com/mcpany/core/server/pkg/tool"
	"github.com/mcpany/core/server/pkg/upstream"
	"github.com/mcpany/core/server/pkg/upstream/factory"
	"github.com/mcpany/core/server/pkg/util"
	"google.golang.org/protobuf/proto"
)

// ServiceRegistryInterface defines the interface for a service registry.
//
// It provides methods for registering, unregistering, and inspecting upstream services,
// acting as the central management point for all external service connections.
type ServiceRegistryInterface interface { //nolint:revive
	// RegisterService registers a new upstream service based on the provided configuration.
	//
	// It initializes the upstream connection, performs capabilities discovery, and registers tools/resources
	// with the respective managers.
	//
	// Parameters:
	//   - ctx: The context for the registration process.
	//   - serviceConfig: The configuration for the service to be registered.
	//
	// Returns:
	//   - string: The unique service key generated for the registered service.
	//   - []*config.ToolDefinition: A list of tools discovered during registration.
	//   - []*config.ResourceDefinition: A list of resources discovered during registration.
	//   - error: An error if the registration fails (e.g., config error, connection failure).
	RegisterService(ctx context.Context, serviceConfig *config.UpstreamServiceConfig) (string, []*config.ToolDefinition, []*config.ResourceDefinition, error)

	// UnregisterService removes a service from the registry and shuts down its upstream connection.
	//
	// Parameters:
	//   - ctx: The context for the unregistration process.
	//   - serviceName: The name of the service to remove.
	//
	// Returns:
	//   - error: An error if the service was not found or if shutdown failed.
	UnregisterService(ctx context.Context, serviceName string) error

	// GetAllServices returns a list of all registered services.
	//
	// Returns:
	//   - []*config.UpstreamServiceConfig: A slice of configuration objects for all registered services.
	//   - error: An error if the retrieval fails (unlikely in current implementation).
	GetAllServices() ([]*config.UpstreamServiceConfig, error)

	// GetServiceInfo retrieves the metadata for a service by its ID.
	//
	// Parameters:
	//   - serviceID: The unique identifier of the service.
	//
	// Returns:
	//   - *tool.ServiceInfo: The service metadata info if found.
	//   - bool: True if the service exists, false otherwise.
	GetServiceInfo(serviceID string) (*tool.ServiceInfo, bool)

	// GetServiceConfig returns the configuration for a given service key.
	//
	// Parameters:
	//   - serviceID: The unique identifier of the service.
	//
	// Returns:
	//   - *config.UpstreamServiceConfig: The service configuration if found.
	//   - bool: True if the service exists, false otherwise.
	GetServiceConfig(serviceID string) (*config.UpstreamServiceConfig, bool)

	// GetServiceError returns the registration error for a service, if any.
	//
	// Parameters:
	//   - serviceID: The unique identifier of the service.
	//
	// Returns:
	//   - string: The error message associated with the service, or empty string.
	//   - bool: True if an error exists, false otherwise.
	GetServiceError(serviceID string) (string, bool)
}

// ServiceRegistry is responsible for managing the lifecycle of upstream services.
//
// It orchestrates the creation of upstream service instances via a factory and registers their
// associated tools, prompts, and resources with the respective managers. It also handles the
// configuration of authentication for each service.
type ServiceRegistry struct {
	mu              sync.RWMutex
	serviceConfigs  map[string]*config.UpstreamServiceConfig
	serviceInfo     map[string]*tool.ServiceInfo
	serviceErrors   map[string]string
	healthErrors    map[string]string
	upstreams       map[string]upstream.Upstream
	factory         factory.Factory
	toolManager     tool.ManagerInterface
	promptManager   prompt.ManagerInterface
	resourceManager resource.ManagerInterface
	authManager     *auth.Manager
}

// New creates a new ServiceRegistry instance.
//
// It initializes the registry with the necessary managers and factory for creating and managing
// upstream services.
//
// Parameters:
//   - factory: The factory used to create upstream service instances.
//   - toolManager: The manager for registering discovered tools.
//   - promptManager: The manager for registering discovered prompts.
//   - resourceManager: The manager for registering discovered resources.
//   - authManager: The manager for registering service-specific authenticators.
//
// Returns:
//   - *ServiceRegistry: A new instance of ServiceRegistry.
func New(factory factory.Factory, toolManager tool.ManagerInterface, promptManager prompt.ManagerInterface, resourceManager resource.ManagerInterface, authManager *auth.Manager) *ServiceRegistry {
	return &ServiceRegistry{
		serviceConfigs:  make(map[string]*config.UpstreamServiceConfig),
		serviceInfo:     make(map[string]*tool.ServiceInfo),
		serviceErrors:   make(map[string]string),
		healthErrors:    make(map[string]string),
		upstreams:       make(map[string]upstream.Upstream),
		factory:         factory,
		toolManager:     toolManager,
		promptManager:   promptManager,
		resourceManager: resourceManager,
		authManager:     authManager,
	}
}

// RegisterService handles the registration of a new upstream service.
//
// It uses the factory to create an upstream instance, discovers its capabilities (tools, prompts, resources),
// and registers them with the appropriate managers. It also sets up any required authenticators for the service.
// If a service with the same name is already registered, the registration will fail.
//
// Parameters:
//   - ctx: The context for the registration process.
//   - serviceConfig: The configuration for the service to be registered.
//
// Returns:
//   - string: The unique service key.
//   - []*config.ToolDefinition: A slice of discovered tool definitions.
//   - []*config.ResourceDefinition: A slice of discovered resource definitions.
//   - error: An error if the registration fails.
func (r *ServiceRegistry) RegisterService(ctx context.Context, serviceConfig *config.UpstreamServiceConfig) (string, []*config.ToolDefinition, []*config.ResourceDefinition, error) {
	r.mu.Lock()

	serviceID, err := util.SanitizeServiceName(serviceConfig.GetName())
	if err != nil {
		r.mu.Unlock()
		return "", nil, nil, fmt.Errorf("failed to generate service key: %w", err)
	}
	if _, ok := r.serviceConfigs[serviceID]; ok {
		// If the service is already registered, check if it's currently active (in upstreams).
		// If it's NOT in upstreams, it means it failed previously, so we allow re-registration (retry).
		if _, isActive := r.upstreams[serviceID]; isActive {
			r.mu.Unlock()
			return "", nil, nil, fmt.Errorf("service with name %q already registered", serviceConfig.GetName())
		}
		// Proceed to overwrite the config and try again
	}

	// Register the config and clear error
	r.serviceConfigs[serviceID] = serviceConfig
	delete(r.serviceErrors, serviceID)

	u, err := r.factory.NewUpstream(serviceConfig)
	if err != nil {
		errMsg := fmt.Sprintf("failed to create upstream for service %s: %v", serviceConfig.GetName(), err)
		r.serviceErrors[serviceID] = errMsg
		r.mu.Unlock()
		return "", nil, nil, fmt.Errorf("%s", errMsg)
	}

	// Store the upstream immediately so we can close it if needed, although it might be partially initialized
	r.upstreams[serviceID] = u
	r.mu.Unlock()

	// Perform registration without holding the lock to avoid blocking other services
	_, discoveredTools, discoveredResources, err := u.Register(ctx, serviceConfig, r.toolManager, r.promptManager, r.resourceManager, false)

	r.mu.Lock()
	defer r.mu.Unlock()

	// Check if the service is still registered (it might have been removed while we were registering)
	if _, ok := r.serviceConfigs[serviceID]; !ok {
		// The service was removed concurrently. We need to clean up what we just created.
		// Note: UnregisterService would have called Shutdown, but it might have run before u.Register completed.
		// So we should try to clean up again.
		_ = u.Shutdown(ctx)
		r.toolManager.ClearToolsForService(serviceID)
		r.promptManager.ClearPromptsForService(serviceID)
		r.resourceManager.ClearResourcesForService(serviceID)
		r.authManager.RemoveAuthenticator(serviceID)

		// If Register failed anyway, return that error, but wrap it to indicate what happened.
		if err != nil {
			return "", nil, nil, fmt.Errorf("service %q was unregistered during registration failure: %w", serviceConfig.GetName(), err)
		}
		return "", nil, nil, fmt.Errorf("service %q was unregistered during registration", serviceConfig.GetName())
	}

	if err != nil {
		r.serviceErrors[serviceID] = err.Error()
		// We keep the service in serviceConfigs so that we know it exists (preventing config drift),
		// but we remove it from upstreams so that we don't try to use it.
		if u, ok := r.upstreams[serviceID]; ok {
			// We try to shutdown, though it might be partially initialized
			_ = u.Shutdown(ctx)
			delete(r.upstreams, serviceID)
		}
		return "", nil, nil, err
	}

	// Perform initial health check
	if checker, ok := u.(upstream.HealthChecker); ok {
		// Use a short timeout for health checks
		checkCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
		if hErr := checker.CheckHealth(checkCtx); hErr != nil {
			r.healthErrors[serviceID] = hErr.Error()
		} else {
			delete(r.healthErrors, serviceID)
		}
		cancel()
	}

	if authConfig := serviceConfig.GetAuthentication(); authConfig != nil {
		if apiKeyConfig := authConfig.GetApiKey(); apiKeyConfig != nil {
			authenticator := auth.NewAPIKeyAuthenticator(apiKeyConfig)
			if err := r.authManager.AddAuthenticator(serviceID, authenticator); err != nil {
				return "", nil, nil, fmt.Errorf("failed to add api key authenticator: %w", err)
			}
		}
		if oauth2Config := authConfig.GetOauth2(); oauth2Config != nil {
			config := &auth.OAuth2Config{
				IssuerURL: oauth2Config.GetIssuerUrl(),
				Audience:  oauth2Config.GetAudience(),
			}
			if err := r.authManager.AddOAuth2Authenticator(ctx, serviceID, config); err != nil {
				return "", nil, nil, fmt.Errorf("failed to add oauth2 authenticator: %w", err)
			}
		}
	}

	return serviceID, discoveredTools, discoveredResources, nil
}

// AddServiceInfo stores metadata about a service, indexed by its ID.
//
// Parameters:
//   - serviceID: The unique identifier for the service.
//   - info: The struct containing the service's metadata.
func (r *ServiceRegistry) AddServiceInfo(serviceID string, info *tool.ServiceInfo) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.serviceInfo[serviceID] = info
}

// GetServiceInfo retrieves the metadata for a service by its ID.
//
// Parameters:
//   - serviceID: The unique identifier for the service.
//
// Returns:
//   - *tool.ServiceInfo: The service info if found.
//   - bool: True if the service exists, false otherwise.
func (r *ServiceRegistry) GetServiceInfo(serviceID string) (*tool.ServiceInfo, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	info, ok := r.serviceInfo[serviceID]
	if !ok {
		return nil, false
	}

	// Clone ServiceInfo to avoid modifying internal state
	clonedInfo := *info // Shallow copy of struct
	if info.Config != nil {
		clonedConfig := proto.Clone(info.Config).(*config.UpstreamServiceConfig)
		util.StripSecretsFromService(clonedConfig)
		r.injectRuntimeInfo(clonedConfig)
		clonedInfo.Config = clonedConfig
	}
	return &clonedInfo, true
}

// GetServiceConfig returns the configuration for a given service key.
//
// Parameters:
//   - serviceID: The unique identifier for the service.
//
// Returns:
//   - *config.UpstreamServiceConfig: The service configuration if found.
//   - bool: True if the service exists, false otherwise.
func (r *ServiceRegistry) GetServiceConfig(serviceID string) (*config.UpstreamServiceConfig, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	serviceConfig, ok := r.serviceConfigs[serviceID]
	if !ok {
		return nil, false
	}
	cloned := proto.Clone(serviceConfig).(*config.UpstreamServiceConfig)
	util.StripSecretsFromService(cloned)
	r.injectRuntimeInfo(cloned)
	return cloned, true
}

// UnregisterService removes a service from the registry.
//
// It sanitizes the service name, shuts down the upstream connection, and clears all associated data
// from the various managers.
//
// Parameters:
//   - ctx: The context for the request.
//   - serviceName: The name of the service to unregister.
//
// Returns:
//   - error: An error if the service is not found or if shutdown fails.
func (r *ServiceRegistry) UnregisterService(ctx context.Context, serviceName string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	serviceID, err := util.SanitizeServiceName(serviceName)
	if err != nil {
		return fmt.Errorf("failed to sanitize service name %q: %w", serviceName, err)
	}

	if _, ok := r.serviceConfigs[serviceID]; !ok {
		return fmt.Errorf("service %q (id: %s) not found", serviceName, serviceID)
	}

	var shutdownErr error
	if u, ok := r.upstreams[serviceID]; ok {
		logging.GetLogger().Info("Shutting down upstream", "service", serviceName)
		if err := u.Shutdown(ctx); err != nil {
			shutdownErr = fmt.Errorf("failed to shutdown upstream for service %s: %w", serviceName, err)
		}
		logging.GetLogger().Info("Upstream shutdown complete", "service", serviceName)
		delete(r.upstreams, serviceID)
	}

	delete(r.serviceConfigs, serviceID)
	delete(r.serviceInfo, serviceID)
	delete(r.serviceErrors, serviceID)
	r.toolManager.ClearToolsForService(serviceID)
	r.promptManager.ClearPromptsForService(serviceID)
	r.resourceManager.ClearResourcesForService(serviceID)
	r.authManager.RemoveAuthenticator(serviceID)
	return shutdownErr
}

// GetServiceError returns the registration error for a service, if any.
//
// It prioritizes registration errors over health check errors.
//
// Parameters:
//   - serviceID: The unique identifier for the service.
//
// Returns:
//   - string: The error message.
//   - bool: True if an error exists, false otherwise.
func (r *ServiceRegistry) GetServiceError(serviceID string) (string, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if err, ok := r.serviceErrors[serviceID]; ok {
		return err, true
	}
	err, ok := r.healthErrors[serviceID]
	return err, ok
}

// StartHealthChecks starts a background loop to periodically check the health of registered upstream services.
//
// Parameters:
//   - ctx: The context to control the lifecycle of the health check loop.
//   - interval: The interval between health checks.
func (r *ServiceRegistry) StartHealthChecks(ctx context.Context, interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				r.checkAllHealth(ctx)
			}
		}
	}()
}

func (r *ServiceRegistry) checkAllHealth(ctx context.Context) {
	r.mu.RLock()
	// Copy upstreams to avoid holding lock during network calls
	targets := make(map[string]upstream.Upstream)
	for id, u := range r.upstreams {
		targets[id] = u
	}
	r.mu.RUnlock()

	for id, u := range targets {
		var errStr string
		if checker, ok := u.(upstream.HealthChecker); ok {
			// Use a short timeout for health checks
			checkCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
			if err := checker.CheckHealth(checkCtx); err != nil {
				errStr = err.Error()
			}
			cancel()
		}

		r.mu.Lock()
		if errStr != "" {
			r.healthErrors[id] = errStr
		} else {
			delete(r.healthErrors, id)
		}
		r.mu.Unlock()
	}
}

// Close gracefully shuts down all registered services.
//
// Parameters:
//   - ctx: The context for the shutdown operation.
//
// Returns:
//   - error: An error if the shutdown of any service fails.
func (r *ServiceRegistry) Close(ctx context.Context) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	var errs []error
	for serviceName, u := range r.upstreams {
		if err := u.Shutdown(ctx); err != nil {
			errs = append(errs, fmt.Errorf("failed to shutdown service %s: %w", serviceName, err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("failed to shutdown some services: %v", errs)
	}
	return nil
}

// GetAllServices returns a list of all registered services.
//
// Returns:
//   - []*config.UpstreamServiceConfig: A slice of configuration objects for all registered services.
//   - error: An error if the retrieval fails.
func (r *ServiceRegistry) GetAllServices() ([]*config.UpstreamServiceConfig, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	services := make([]*config.UpstreamServiceConfig, 0, len(r.serviceConfigs))
	for _, service := range r.serviceConfigs {
		cloned := proto.Clone(service).(*config.UpstreamServiceConfig)
		util.StripSecretsFromService(cloned)
		r.injectRuntimeInfo(cloned)
		services = append(services, cloned)
	}
	return services, nil
}

// injectRuntimeInfo populates runtime information (error status, tool count) into the config.
// Caller MUST hold r.mu lock (Read or Write).
func (r *ServiceRegistry) injectRuntimeInfo(config *config.UpstreamServiceConfig) {
	if config == nil {
		return
	}
	key := config.GetSanitizedName()
	if key == "" {
		// Fallback if sanitized name is not set (should not happen for registered services)
		key, _ = util.SanitizeServiceName(config.GetName())
	}

	// Error: prioritize explicit service errors (registration failure), then health check errors
	if err, ok := r.serviceErrors[key]; ok {
		config.SetLastError(err)
	} else if err, ok := r.healthErrors[key]; ok {
		config.SetLastError(err)
	}

	// Tool Count
	// r.toolManager is thread-safe (xsync.Map based) so calling ListTools is safe.
	// However, ListTools acquires its own locks.
	tools := r.toolManager.ListTools()
	count := 0
	for _, t := range tools {
		if t.Tool().GetServiceId() == key {
			count++
		}
	}
	//nolint:gosec // Tool count is unlikely to exceed int32 max
	config.SetToolCount(int32(count))
}
