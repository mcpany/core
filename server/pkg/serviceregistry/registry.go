// Package serviceregistry provides service registry functionality.

package serviceregistry

import (
	"context"
	"fmt"
	"sync"

	"github.com/mcpany/core/pkg/auth"
	"github.com/mcpany/core/pkg/prompt"
	"github.com/mcpany/core/pkg/resource"
	"github.com/mcpany/core/pkg/tool"
	"github.com/mcpany/core/pkg/upstream"
	"github.com/mcpany/core/pkg/upstream/factory"
	"github.com/mcpany/core/pkg/util"
	config "github.com/mcpany/core/proto/config/v1"
)

// ServiceRegistryInterface defines the interface for a service registry.
// It provides a method for registering new upstream services.
type ServiceRegistryInterface interface { //nolint:revive
	// RegisterService registers a new upstream service based on the provided
	// configuration. It returns the generated service key, a list of any tools
	// discovered during registration, and an error if the registration fails.
	RegisterService(ctx context.Context, serviceConfig *config.UpstreamServiceConfig) (string, []*config.ToolDefinition, []*config.ResourceDefinition, error)
	// UnregisterService removes a service from the registry.
	UnregisterService(ctx context.Context, serviceName string) error
	// GetAllServices returns a list of all registered services.
	GetAllServices() ([]*config.UpstreamServiceConfig, error)
	// GetServiceInfo retrieves the metadata for a service by its ID.
	GetServiceInfo(serviceID string) (*tool.ServiceInfo, bool)
	// GetServiceConfig returns the configuration for a given service key.
	GetServiceConfig(serviceID string) (*config.UpstreamServiceConfig, bool)
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
	defer r.mu.Unlock()

	serviceID, err := util.SanitizeServiceName(serviceConfig.GetName())
	if err != nil {
		return "", nil, nil, fmt.Errorf("failed to generate service key: %w", err)
	}
	if _, ok := r.serviceConfigs[serviceID]; ok {
		return "", nil, nil, fmt.Errorf("service with name %q already registered", serviceConfig.GetName())
	}

	u, err := r.factory.NewUpstream(serviceConfig)
	if err != nil {
		return "", nil, nil, fmt.Errorf("failed to create upstream for service %s: %w", serviceConfig.GetName(), err)
	}

	serviceID, discoveredTools, discoveredResources, err := u.Register(ctx, serviceConfig, r.toolManager, r.promptManager, r.resourceManager, false)
	if err != nil {
		return "", nil, nil, err
	}

	r.serviceConfigs[serviceID] = serviceConfig
	r.upstreams[serviceID] = u

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
func (r *ServiceRegistry) UnregisterService(ctx context.Context, serviceName string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.serviceConfigs[serviceName]; !ok {
		return fmt.Errorf("service %q not found", serviceName)
	}

	if u, ok := r.upstreams[serviceName]; ok {
		if err := u.Shutdown(ctx); err != nil {
			return fmt.Errorf("failed to shutdown upstream for service %s: %w", serviceName, err)
		}
		delete(r.upstreams, serviceName)
	}

	delete(r.serviceConfigs, serviceName)
	delete(r.serviceInfo, serviceName)
	r.toolManager.ClearToolsForService(serviceName)
	r.promptManager.ClearPromptsForService(serviceName)
	r.resourceManager.ClearResourcesForService(serviceName)
	r.authManager.RemoveAuthenticator(serviceName)
	return nil
}

// GetAllServices returns a list of all registered services.
func (r *ServiceRegistry) GetAllServices() ([]*config.UpstreamServiceConfig, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	services := make([]*config.UpstreamServiceConfig, 0, len(r.serviceConfigs))
	for _, service := range r.serviceConfigs {
		services = append(services, service)
	}
	return services, nil
}
