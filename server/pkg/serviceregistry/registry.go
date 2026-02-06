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
// Summary: Interface for managing the lifecycle of upstream services.
type ServiceRegistryInterface interface { //nolint:revive
	// RegisterService registers a new upstream service based on the provided configuration.
	//
	// Parameters:
	//   ctx (context.Context): The registration context.
	//   serviceConfig (*config.UpstreamServiceConfig): The service configuration. Must not be nil.
	//
	// Returns:
	//   (string): The generated service ID.
	//   ([]*config.ToolDefinition): Discovered tools.
	//   ([]*config.ResourceDefinition): Discovered resources.
	//   (error): Error if registration fails.
	RegisterService(ctx context.Context, serviceConfig *config.UpstreamServiceConfig) (string, []*config.ToolDefinition, []*config.ResourceDefinition, error)

	// UnregisterService removes a service from the registry and shuts down its upstream connection.
	//
	// Parameters:
	//   ctx (context.Context): The context for the operation.
	//   serviceName (string): The name of the service to remove.
	//
	// Returns:
	//   (error): Error if the service is not found or shutdown fails.
	UnregisterService(ctx context.Context, serviceName string) error

	// GetAllServices returns a list of all registered services.
	//
	// Returns:
	//   ([]*config.UpstreamServiceConfig): List of service configurations.
	//   (error): Error if retrieval fails.
	GetAllServices() ([]*config.UpstreamServiceConfig, error)

	// GetServiceInfo retrieves the metadata for a service by its ID.
	//
	// Parameters:
	//   serviceID (string): The service ID.
	//
	// Returns:
	//   (*tool.ServiceInfo): The service metadata.
	//   (bool): True if found.
	GetServiceInfo(serviceID string) (*tool.ServiceInfo, bool)

	// GetServiceConfig returns the configuration for a given service key.
	//
	// Parameters:
	//   serviceID (string): The service ID.
	//
	// Returns:
	//   (*config.UpstreamServiceConfig): The service configuration.
	//   (bool): True if found.
	GetServiceConfig(serviceID string) (*config.UpstreamServiceConfig, bool)

	// GetServiceError returns the registration error for a service, if any.
	//
	// Parameters:
	//   serviceID (string): The service ID.
	//
	// Returns:
	//   (string): The error message.
	//   (bool): True if an error exists.
	GetServiceError(serviceID string) (string, bool)
}

// ServiceRegistry is responsible for managing the lifecycle of upstream services.
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
// Parameters:
//   factory (factory.Factory): The factory for creating upstreams.
//   toolManager (tool.ManagerInterface): The tool manager.
//   promptManager (prompt.ManagerInterface): The prompt manager.
//   resourceManager (resource.ManagerInterface): The resource manager.
//   authManager (*auth.Manager): The authentication manager.
//
// Returns:
//   (*ServiceRegistry): The initialized registry.
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
// Parameters:
//   ctx (context.Context): The registration context.
//   serviceConfig (*config.UpstreamServiceConfig): The service configuration. Must not be nil.
//
// Returns:
//   (string): The service ID.
//   ([]*config.ToolDefinition): Discovered tools.
//   ([]*config.ResourceDefinition): Discovered resources.
//   (error): Error if registration fails.
//
// Side Effects:
//   - Creates an upstream connection.
//   - Updates managers with discovered capabilities.
//   - Configures authentication.
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

	if err == nil {
		// Safety Check: Duplicate Tool Detection
		// Although we enforce namespacing, we warn if multiple services expose the same tool name
		// to avoid user confusion or potential shadowing if namespacing is bypassed or ambiguous.
		existingTools := r.toolManager.ListTools()
		for _, dt := range discoveredTools {
			for _, et := range existingTools {
				// Check if names match and they are from DIFFERENT services
				// et.Tool().GetServiceId() might be empty for internal tools, or the service ID we just generated.
				// We care if it's NOT the current serviceID.
				if dt.GetName() == et.Tool().GetName() && et.Tool().GetServiceId() != serviceID {
					logging.GetLogger().Warn("Duplicate tool name detected across services",
						"tool", dt.GetName(),
						"service_new", serviceConfig.GetName(),
						"service_existing", et.Tool().GetServiceId(),
						"warning", "Potential confusion or shadowing")
				}
			}
		}
	}

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
//   serviceID (string): The service ID.
//   info (*tool.ServiceInfo): The service metadata.
func (r *ServiceRegistry) AddServiceInfo(serviceID string, info *tool.ServiceInfo) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.serviceInfo[serviceID] = info
}

// GetServiceInfo retrieves the metadata for a service by its ID.
//
// Parameters:
//   serviceID (string): The service ID.
//
// Returns:
//   (*tool.ServiceInfo): The service metadata.
//   (bool): True if found.
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
//   serviceID (string): The service ID.
//
// Returns:
//   (*config.UpstreamServiceConfig): The service configuration.
//   (bool): True if found.
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
// Summary: Removes a service from the registry and shuts it down.
//
// Parameters:
//   ctx (context.Context): The context for the operation.
//   serviceName (string): The name of the service.
//
// Returns:
//   (error): Error if not found or shutdown fails.
//
// Side Effects:
//   - Stops the upstream connection.
//   - Clears data from managers.
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
// Parameters:
//   serviceID (string): The service ID.
//
// Returns:
//   (string): The error message.
//   (bool): True if an error exists.
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
//   ctx (context.Context): The context to cancel the loop.
//   interval (time.Duration): The check interval.
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

	// ⚡ BOLT: Parallelize health checks with concurrency limit
	// Randomized Selection from Top 5 High-Impact Targets
	var wg sync.WaitGroup
	sem := make(chan struct{}, 20) // Limit to 20 concurrent checks

	for id, u := range targets {
		wg.Add(1)
		go func(id string, u upstream.Upstream) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

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
		}(id, u)
	}
	wg.Wait()
}

// Close gracefully shuts down all registered services.
//
// Parameters:
//   ctx (context.Context): The shutdown context.
//
// Returns:
//   (error): Error if any shutdown fails.
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
// Summary: Lists all services with redacted secrets.
//
// Returns:
//   ([]*config.UpstreamServiceConfig): List of service configs.
//   (error): Error if retrieval fails.
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
	// ⚡ BOLT: Optimized from O(N*M) to O(N) by using indexed lookup instead of iterating all tools.
	// Randomized Selection from Top 5 High-Impact Targets
	// Note: Tools are indexed by Service ID (hash), which might differ from the registry key (Sanitized Name).
	toolKey := config.GetId()
	if toolKey == "" {
		toolKey = key
	}
	count := r.toolManager.GetToolCountForService(toolKey)
	//nolint:gosec // Tool count is unlikely to exceed int32 max
	config.SetToolCount(int32(count))
}
