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
	"errors"
	"fmt"
	"regexp"
	"sync"

	"github.com/mcpxy/core/pkg/auth"
	"github.com/mcpxy/core/pkg/prompt"
	"github.com/mcpxy/core/pkg/resource"
	"github.com/mcpxy/core/pkg/tool"
	"github.com/mcpxy/core/pkg/upstream/factory"
	config "github.com/mcpxy/core/proto/config/v1"
)

var (
	// serviceNameRegex is the regular expression for validating service names.
	// It allows word characters (a-z, A-Z, 0-9, _), and hyphens.
	serviceNameRegex = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
)

// ServiceRegistryInterface defines the interface for a service registry.
// It provides a method for registering new upstream services.
type ServiceRegistryInterface interface {
	// RegisterService registers a new upstream service based on the provided
	// configuration. It returns the generated service ID, a list of any tools
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
	globalConfig    *config.GlobalSettings
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
func New(factory factory.Factory, toolManager tool.ToolManagerInterface, promptManager prompt.PromptManagerInterface, resourceManager resource.ResourceManagerInterface, authManager *auth.AuthManager, globalConfig *config.GlobalSettings) *ServiceRegistry {
	return &ServiceRegistry{
		serviceConfigs:  make(map[string]*config.UpstreamServiceConfig),
		factory:         factory,
		toolManager:     toolManager,
		promptManager:   promptManager,
		resourceManager: resourceManager,
		authManager:     authManager,
		globalConfig:    globalConfig,
	}
}

// validateServiceName checks if the service name is valid.
func (r *ServiceRegistry) validateServiceName(name string) error {
	if len(name) > 62 {
		return errors.New("service name length exceeds 62 characters")
	}
	if !serviceNameRegex.MatchString(name) {
		return errors.New("service name contains invalid characters")
	}
	return nil
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
// Returns the unique service ID, a slice of discovered tool definitions, and
// an error if the registration fails.
func (r *ServiceRegistry) RegisterService(ctx context.Context, serviceConfig *config.UpstreamServiceConfig) (string, []*config.ToolDefinition, []*config.ResourceDefinition, error) {
	if err := r.validateServiceName(serviceConfig.GetName()); err != nil {
		return "", nil, nil, fmt.Errorf("invalid service name %q: %w", serviceConfig.GetName(), err)
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	strategy := r.globalConfig.GetServiceNameStrategy()
	if r.isServiceNameRegistered(serviceConfig.GetName()) {
		switch strategy {
		case config.GlobalSettings_STRICT:
			return "", nil, nil, fmt.Errorf("service with name %q already registered", serviceConfig.GetName())
		case config.GlobalSettings_MERGE, config.GlobalSettings_MERGE_HASH, config.GlobalSettings_MERGE_IGNORE:
			// Allow registration, but the upstream will handle tool/resource conflicts.
			break
		default: // Default to MERGE for now
			break
		}
	}

	u, err := r.factory.NewUpstream(serviceConfig)
	if err != nil {
		return "", nil, nil, fmt.Errorf("failed to create upstream for service %s: %w", serviceConfig.GetName(), err)
	}

	serviceID, discoveredTools, discoveredResources, err := u.Register(ctx, serviceConfig, r.toolManager, r.promptManager, r.resourceManager, false, strategy)
	if err != nil {
		return "", nil, nil, err
	}

	r.serviceConfigs[serviceID] = serviceConfig

	if authConfig := serviceConfig.GetAuthentication(); authConfig != nil {
		if apiKeyConfig := authConfig.GetApiKey(); apiKeyConfig != nil {
			authenticator := auth.NewAPIKeyAuthenticator(apiKeyConfig)
			r.authManager.AddAuthenticator(serviceID, authenticator)
		}
		if oauth2Config := authConfig.GetOauth2(); oauth2Config != nil {
			config := &auth.OAuth2Config{
				IssuerURL:    oauth2Config.GetIssuerUrl(),
				Audience:     oauth2Config.GetAudience(),
			}
			if err := r.authManager.AddOAuth2Authenticator(ctx, serviceID, config); err != nil {
				return "", nil, nil, fmt.Errorf("failed to add oauth2 authenticator: %w", err)
			}
		}
	}

	return serviceID, discoveredTools, discoveredResources, nil
}

// GetServiceConfig returns the configuration for a given service ID.
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

func (r *ServiceRegistry) isServiceNameRegistered(name string) bool {
	for _, cfg := range r.serviceConfigs {
		if cfg.GetName() == name {
			return true
		}
	}
	return false
}
