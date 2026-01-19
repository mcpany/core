// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

// Package serviceregistry provides service registry functionality.
package serviceregistry

import (
	"context"
	"fmt"
	"sync"

	config "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/auth"
	"github.com/mcpany/core/server/pkg/logging"
	"github.com/mcpany/core/server/pkg/prompt"
	"github.com/mcpany/core/server/pkg/resource"
	"github.com/mcpany/core/server/pkg/tool"
	"github.com/mcpany/core/server/pkg/upstream"
	"github.com/mcpany/core/server/pkg/upstream/factory"
	"github.com/mcpany/core/server/pkg/util"
)

// ServiceRegistryInterface defines the interface for a service registry.
// It provides a method for registering new upstream services.
type ServiceRegistryInterface interface { //nolint:revive
	// RegisterService registers a new upstream service based on the provided
	// configuration. It returns the generated service key, a list of any tools
	// discovered during registration, and an error if the registration fails.
	RegisterService(ctx context.Context, serviceConfig *config.UpstreamServiceConfig) (string, []*config.ToolDefinition, []*config.ResourceDefinition, error)
	// UnregisterService removes a service from the registry.
	//
	// ctx is the context for the request.
	// serviceName is the serviceName.
	//
	// Returns an error if the operation fails.
	UnregisterService(ctx context.Context, serviceName string) error
	// GetAllServices returns a list of all registered services.
	//
	// Returns the result.
	// Returns an error if the operation fails.
	GetAllServices() ([]*config.UpstreamServiceConfig, error)
	// GetServiceInfo retrieves the metadata for a service by its ID.
	//
	// serviceID is the serviceID.
	//
	// Returns the result.
	// Returns true if successful.
	GetServiceInfo(serviceID string) (*tool.ServiceInfo, bool)
	// GetServiceConfig returns the configuration for a given service key.
	//
	// serviceID is the serviceID.
	//
	// Returns the result.
	// Returns true if successful.
	GetServiceConfig(serviceID string) (*config.UpstreamServiceConfig, bool)
	// GetServiceError returns the registration error for a service, if any.
	//
	// serviceID is the serviceID.
	//
	// Returns the result.
	// Returns true if successful.
	GetServiceError(serviceID string) (string, bool)
}

// ServiceRegistry is responsible for managing the lifecycle of upstream
// services. It orchestrates the creation of upstream service instances via a
// factory and registers their associated tools, prompts, and resources with the
// respective managers. It also handles the configuration of authentication for
// each service.
type ServiceRegistry struct {
	mu              sync.RWMutex
	serviceConfigs  map[string]*config.UpstreamServiceConfig
	serviceInfo     map[string]*tool.ServiceInfo
	serviceErrors   map[string]string
	upstreams       map[string]upstream.Upstream
	factory         factory.Factory
	toolManager     tool.ManagerInterface
	promptManager   prompt.ManagerInterface
	resourceManager resource.ManagerInterface
	authManager     *auth.Manager
}

// New creates a new ServiceRegistry instance, which is responsible for managing
// the lifecycle of upstream services.
//
// Parameters:
//   - factory: The factory used to create upstream service instances.
//   - toolManager: The manager for registering discovered tools.
//   - promptManager: The manager for registering discovered prompts.
//   - resourceManager: The manager for registering discovered resources.
//   - authManager: The manager for registering service-specific authenticators.
//
// Returns a new instance of `ServiceRegistry`.
func New(factory factory.Factory, toolManager tool.ManagerInterface, promptManager prompt.ManagerInterface, resourceManager resource.ManagerInterface, authManager *auth.Manager) *ServiceRegistry {
	return &ServiceRegistry{
		serviceConfigs:  make(map[string]*config.UpstreamServiceConfig),
		serviceInfo:     make(map[string]*tool.ServiceInfo),
		serviceErrors:   make(map[string]string),
		upstreams:       make(map[string]upstream.Upstream),
		factory:         factory,
		toolManager:     toolManager,
		promptManager:   promptManager,
		resourceManager: resourceManager,
		authManager:     authManager,
	}
}

// RegisterService handles the registration of a new upstream service. It uses
// the factory to create an upstream instance, discovers its capabilities (tools,
// prompts, resources), and registers them with the appropriate managers. It also
// sets up any required authenticators for the service.
//
// If a service with the same name is already registered, the registration will
// fail.
//
// Parameters:
//   - ctx: The context for the registration process.
//   - serviceConfig: The configuration for the service to be registered.
//
// Returns the unique service key, a slice of discovered tool definitions, and
// an error if the registration fails.
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
// serviceID is the unique identifier for the service.
// info is the ServiceInfo struct containing the service's metadata.
func (r *ServiceRegistry) AddServiceInfo(serviceID string, info *tool.ServiceInfo) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.serviceInfo[serviceID] = info
}

// GetServiceInfo retrieves the metadata for a service by its ID.
//
// serviceID is the unique identifier for the service.
// It returns the ServiceInfo and a boolean indicating whether the service was found.
func (r *ServiceRegistry) GetServiceInfo(serviceID string) (*tool.ServiceInfo, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	info, ok := r.serviceInfo[serviceID]
	return info, ok
}

// GetServiceConfig returns the configuration for a given service key.
//
// Parameters:
//   - serviceID: The unique identifier for the service.
//
// Returns the service configuration and a boolean indicating whether the service
// was found.
func (r *ServiceRegistry) GetServiceConfig(serviceID string) (*config.UpstreamServiceConfig, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	serviceConfig, ok := r.serviceConfigs[serviceID]
	return serviceConfig, ok
}

// UnregisterService removes a service from the registry.
//
// ctx is the context for the request.
// serviceName is the serviceName.
//
// Returns an error if the operation fails.
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
// serviceID is the serviceID.
//
// Returns the result.
// Returns true if successful.
func (r *ServiceRegistry) GetServiceError(serviceID string) (string, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	err, ok := r.serviceErrors[serviceID]
	return err, ok
}

// Close gracefully shuts down all registered services.
//
// ctx is the context for the request.
//
// Returns an error if the operation fails.
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
// Returns the result.
// Returns an error if the operation fails.
func (r *ServiceRegistry) GetAllServices() ([]*config.UpstreamServiceConfig, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	services := make([]*config.UpstreamServiceConfig, 0, len(r.serviceConfigs))
	for _, service := range r.serviceConfigs {
		services = append(services, service)
	}
	return services, nil
}
