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
	RegisterService(ctx context.Context, serviceConfig *config.UpstreamServiceConfig) (string, []*config.ToolDefinition, error)
}

// ServiceRegistry is responsible for managing the lifecycle of upstream services.
// It orchestrates the creation of upstream instances via a factory and registers
// their associated tools, prompts, and resources with the respective managers.
type ServiceRegistry struct {
	mu              sync.RWMutex
	serviceConfigs  map[string]*config.UpstreamServiceConfig
	factory         factory.Factory
	toolManager     tool.ToolManagerInterface
	promptManager   prompt.PromptManagerInterface
	resourceManager resource.ResourceManagerInterface
	authManager     *auth.AuthManager
}

// New creates a new ServiceRegistry instance.
//
// factory is the factory used to create upstream service instances.
// toolManager is the manager for registering discovered tools.
// promptManager is the manager for registering discovered prompts.
// resourceManager is the manager for registering discovered resources.
// authManager is the manager for registering service-specific authenticators.
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
// the factory to create an upstream instance and then calls its Register method
// to discover and register its capabilities (tools, prompts, resources).
//
// ctx is the context for the registration process.
// serviceConfig is the configuration for the service to be registered.
// It returns the unique service key, a slice of discovered tool definitions,
// and an error if the registration fails.
func (r *ServiceRegistry) RegisterService(ctx context.Context, serviceConfig *config.UpstreamServiceConfig) (string, []*config.ToolDefinition, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	u, err := r.factory.NewUpstream(serviceConfig)
	if err != nil {
		return "", nil, fmt.Errorf("failed to create upstream for service %s: %w", serviceConfig.GetName(), err)
	}

	serviceKey, discoveredTools, err := u.Register(ctx, serviceConfig, r.toolManager, r.promptManager, r.resourceManager, false)
	if err != nil {
		return "", nil, err
	}

	r.serviceConfigs[serviceKey] = serviceConfig

	if authConfig := serviceConfig.GetAuthentication(); authConfig != nil {
		if apiKeyConfig := authConfig.GetApiKey(); apiKeyConfig != nil {
			authenticator := auth.NewAPIKeyAuthenticator(apiKeyConfig)
			r.authManager.AddAuthenticator(serviceKey, authenticator)
		}
	}

	return serviceKey, discoveredTools, nil
}

// GetServiceConfig returns the configuration for a given service key.
//
// serviceKey is the unique identifier for the service.
// It returns the service configuration and a boolean indicating whether the
// service was found.
func (r *ServiceRegistry) GetServiceConfig(serviceKey string) (*config.UpstreamServiceConfig, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	serviceConfig, ok := r.serviceConfigs[serviceKey]
	return serviceConfig, ok
}
