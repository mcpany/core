/*
 * Copyright 2025 Author(s) of MCP-XY
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package serviceregistry

import (
	"context"
	"fmt"
	"sync"

	"github.com/mcpxy/core/pkg/auth"
	"github.com/mcpxy/core/pkg/prompt"
	"github.com/mcpxy/core/pkg/resource"
	"github.com/mcpxy/core/pkg/tool"
	"github.com/mcpxy/core/pkg/upstream/factory"
	config "github.com/mcpxy/core/proto/config/v1"
)

// ServiceRegistryInterface defines the interface for a service registry.
// It provides a method for registering new upstream services.
type ServiceRegistryInterface interface {
	// RegisterService registers a new upstream service based on the provided
	// configuration. It returns the generated service key, a list of any tools
	// discovered during registration, and an error if the registration fails.
	RegisterService(ctx context.Context, serviceConfig *config.UpstreamServiceConfig) (string, []*config.ToolDefinition, []*config.ResourceDefinition, error)
}

// ServiceRegistry is responsible for managing the lifecycle of upstream
// services. It orchestrates the creation of upstream service instances via a
// factory and registers their associated tools, prompts, and resources with the
// respective managers. It also handles the configuration of authentication for
// each service.
type ServiceRegistry struct {
	mu              sync.RWMutex
	serviceConfigs  map[string]*config.UpstreamServiceConfig
	factory         factory.Factory
	toolManager     tool.ToolManagerInterface
	promptManager   prompt.PromptManagerInterface
	resourceManager resource.ResourceManagerInterface
	authManager     *auth.AuthManager
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
func New(factory factory.Factory, toolManager tool.ToolManagerInterface, promptManager prompt.PromptManagerInterface, resourceManager resource.ResourceManagerInterface, authManager *auth.AuthManager) *ServiceRegistry {
	return &ServiceRegistry{
		serviceConfigs:  make(map[string]*config.UpstreamServiceConfig),
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

	u, err := r.factory.NewUpstream(serviceConfig)
	if err != nil {
		return "", nil, nil, fmt.Errorf("failed to create upstream for service %s: %w", serviceConfig.GetName(), err)
	}

	serviceKey, discoveredTools, discoveredResources, err := u.Register(ctx, serviceConfig, r.toolManager, r.promptManager, r.resourceManager, false)
	if err != nil {
		return "", nil, nil, err
	}

	if _, ok := r.serviceConfigs[serviceKey]; ok {
		r.toolManager.ClearToolsForService(serviceKey) // Clean up the just-registered tools
		return "", nil, nil, fmt.Errorf("service with name %q already registered", serviceConfig.GetName())
	}

	r.serviceConfigs[serviceKey] = serviceConfig

	if authConfig := serviceConfig.GetAuthentication(); authConfig != nil {
		if apiKeyConfig := authConfig.GetApiKey(); apiKeyConfig != nil {
			authenticator := auth.NewAPIKeyAuthenticator(apiKeyConfig)
			r.authManager.AddAuthenticator(serviceKey, authenticator)
		}
		if oauth2Config := authConfig.GetOauth2(); oauth2Config != nil {
			config := &auth.OAuth2Config{
				IssuerURL:    oauth2Config.GetIssuerUrl(),
				Audience:     oauth2Config.GetAudience(),
			}
			if err := r.authManager.AddOAuth2Authenticator(ctx, serviceKey, config); err != nil {
				return "", nil, nil, fmt.Errorf("failed to add oauth2 authenticator: %w", err)
			}
		}
	}

	return serviceKey, discoveredTools, discoveredResources, nil
}

// GetServiceConfig returns the configuration for a given service key.
//
// Parameters:
//   - serviceKey: The unique identifier for the service.
//
// Returns the service configuration and a boolean indicating whether the service
// was found.
func (r *ServiceRegistry) GetServiceConfig(serviceKey string) (*config.UpstreamServiceConfig, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	serviceConfig, ok := r.serviceConfigs[serviceKey]
	return serviceConfig, ok
}
