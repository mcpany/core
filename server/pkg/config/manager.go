// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"sort"
	"strings"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/logging"
	"github.com/mcpany/core/server/pkg/profile"
	"github.com/mcpany/core/server/pkg/util"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/encoding/prototext"
	"google.golang.org/protobuf/proto"
	"sigs.k8s.io/yaml"

	"github.com/Masterminds/semver/v3"
)

// MergeStrategyReplace indicates that the new configuration list should replace the existing one.
const MergeStrategyReplace = "replace"

// UpstreamServiceManager manages the lifecycle and configuration of upstream services.
//
// Summary: Handles loading, validating, and merging service configurations from various sources.
//
// Side Effects:
//   - Stores the final, merged UpstreamServiceConfig objects.
//   - Makes HTTP requests to fetch remote configurations.
type UpstreamServiceManager struct {
	log               *slog.Logger
	services          map[string]*configv1.UpstreamServiceConfig // Stores the final, merged UpstreamServiceConfig objects
	servicePriorities map[string]int32
	httpClient        *http.Client
	newGitHub         func(ctx context.Context, rawURL string) (*GitHub, error)
	enabledProfiles   []string
	// New fields for profile management
	profileServiceOverrides map[string]*configv1.ProfileServiceConfig // Stores overrides from profiles
	profileSecrets          map[string]*configv1.SecretValue          // Stores secrets resolved from profiles
}

// NewUpstreamServiceManager creates a new instance of UpstreamServiceManager.
//
// Summary: Initializes a new UpstreamServiceManager with the specified profiles.
//
// Parameters:
//   - enabledProfiles: []string. A list of profile names that are active. Services must match one of these profiles to be loaded.
//
// Returns:
//   - *UpstreamServiceManager: A pointer to a fully initialized UpstreamServiceManager.
func NewUpstreamServiceManager(enabledProfiles []string) *UpstreamServiceManager {
	if len(enabledProfiles) == 0 {
		enabledProfiles = []string{"default"}
	}
	return &UpstreamServiceManager{
		log:                     logging.GetLogger().With("component", "UpstreamServiceManager"),
		services:                make(map[string]*configv1.UpstreamServiceConfig),
		servicePriorities:       make(map[string]int32),
		profileServiceOverrides: make(map[string]*configv1.ProfileServiceConfig),
		profileSecrets:          make(map[string]*configv1.SecretValue),
		httpClient: &http.Client{
			Transport: &http.Transport{
				DialContext: util.SafeDialContext,
			},
		},
		newGitHub:       NewGitHub,
		enabledProfiles: enabledProfiles,
	}
}

// LoadAndMergeServices loads all upstream services from the provided configuration.
//
// Summary: Processes local and remote service configurations, merging them based on priority and name.
//
// Parameters:
//   - ctx: context.Context. The context for the operation.
//   - config: *configv1.McpAnyServerConfig. The main server configuration containing service definitions and collection references.
//
// Returns:
//   - []*configv1.UpstreamServiceConfig: A slice of merged service configurations.
//   - error: An error if any critical failure occurs during loading or merging.
//
// Side Effects:
//   - May clear existing services if a replace strategy is configured.
//   - Fetches remote collections via HTTP.
func (m *UpstreamServiceManager) LoadAndMergeServices(ctx context.Context, config *configv1.McpAnyServerConfig) ([]*configv1.UpstreamServiceConfig, error) {
	// Respect merge strategy
	if strategy := config.GetMergeStrategy(); strategy != nil {
		if strategy.GetUpstreamServiceList() == MergeStrategyReplace {
			m.log.Info("Merge strategy is 'replace', clearing existing services")
			m.services = make(map[string]*configv1.UpstreamServiceConfig)
			m.servicePriorities = make(map[string]int32)
		}
		if strategy.GetProfileList() == MergeStrategyReplace {
			m.log.Info("Merge strategy is 'replace', clearing existing profile overrides")
			m.profileServiceOverrides = make(map[string]*configv1.ProfileServiceConfig)
			m.profileSecrets = make(map[string]*configv1.SecretValue)
		}
	}

	// Initialize Profile Manager and resolve overrides
	pm := profile.NewManager(config.GetGlobalSettings().GetProfileDefinitions())
	for _, profileName := range m.enabledProfiles {
		configs, secrets, err := pm.ResolveProfile(profileName)
		if err != nil {
			m.log.Warn("Failed to resolve profile", "profile", profileName, "error", err)
			continue
		}
		// Merge resolved configs
		for serviceID, cfg := range configs {
			if existing, exists := m.profileServiceOverrides[serviceID]; exists {
				proto.Merge(existing, cfg)
			} else {
				m.profileServiceOverrides[serviceID] = proto.Clone(cfg).(*configv1.ProfileServiceConfig)
			}
		}
		// Merge resolved secrets
		for key, secret := range secrets {
			m.profileSecrets[key] = proto.Clone(secret).(*configv1.SecretValue)
		}
	}

	// Load local services with default priority 0
	for _, service := range config.GetUpstreamServices() {
		if err := m.addService(service, 0); err != nil {
			return nil, err
		}
	}

	// Load and merge remote service collections
	for _, collection := range config.GetCollections() {
		if err := m.loadAndMergeCollection(ctx, collection); err != nil {
			m.log.Warn("Failed to load upstream service collection", "name", collection.GetName(), "url", collection.GetHttpUrl(), "error", err)
			// Continue loading other collections even if one fails
		}
	}

	// Return the final list of services
	services := make([]*configv1.UpstreamServiceConfig, 0, len(m.services))
	for _, service := range m.services {
		services = append(services, service)
	}
	sort.Slice(services, func(i, j int) bool {
		return services[i].GetName() < services[j].GetName()
	})
	return services, nil
}

func (m *UpstreamServiceManager) loadAndMergeCollection(ctx context.Context, collection *configv1.Collection) error {
	// 1. Load inline services
	for _, service := range collection.GetServices() {
		priority := collection.GetPriority()
		if service.HasPriority() {
			priority = service.GetPriority()
		}
		if err := m.addService(service, priority); err != nil {
			m.log.Warn("Failed to add inline service from collection", "collection", collection.GetName(), "service", service.GetName(), "error", err)
		}
	}

	// 2. Load from URL if present
	if collection.GetHttpUrl() == "" {
		return nil
	}

	if isGitHubURL(collection.GetHttpUrl()) {
		g, err := m.newGitHub(ctx, collection.GetHttpUrl())
		if err != nil {
			return fmt.Errorf("failed to parse github url: %w", err)
		}

		if g.URLType == "blob" {
			return m.loadFromURL(ctx, g.ToRawContentURL(), collection)
		}

		contents, err := g.List(ctx, collection.GetAuthentication())
		if err != nil {
			return fmt.Errorf("failed to list github directory: %w", err)
		}

		for _, content := range contents {
			if content.Type == "dir" {
				newCollection := configv1.Collection_builder{
					Name:           proto.String(collection.GetName()),
					HttpUrl:        proto.String(content.HTMLURL),
					Priority:       proto.Int32(collection.GetPriority()),
					Authentication: collection.GetAuthentication(),
				}.Build()
				if err := m.loadAndMergeCollection(ctx, newCollection); err != nil {
					m.log.Warn("Failed to load from github url", "url", content.HTMLURL, "error", err)
				}
			} else if content.Type == "file" {
				fileName := content.Name
				if fileName == "" {
					fileName = content.HTMLURL
				}
				lowerName := strings.ToLower(fileName)
				if !strings.HasSuffix(lowerName, ".yaml") && !strings.HasSuffix(lowerName, ".yml") && !strings.HasSuffix(lowerName, ".json") {
					continue
				}
				if err := m.loadFromURL(ctx, content.DownloadURL, collection); err != nil {
					m.log.Warn("Failed to load from github url", "url", content.DownloadURL, "error", err)
				}
			}
		}
		return nil
	}

	return m.loadFromURL(ctx, collection.GetHttpUrl(), collection)
}

func (m *UpstreamServiceManager) loadFromURL(ctx context.Context, url string, collection *configv1.Collection) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("failed to create http request: %w", err)
	}

	if auth := collection.GetAuthentication(); auth != nil {
		if err := m.applyAuthentication(req, auth); err != nil {
			return fmt.Errorf("failed to apply authentication for collection %s: %w", collection.GetName(), err)
		}
	}

	resp, err := m.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to fetch collection from url %s: %w", url, err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to fetch collection from url %s: status code %d", url, resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	var services []*configv1.UpstreamServiceConfig
	contentType := resp.Header.Get("Content-Type")
	if err := m.unmarshalServices(body, &services, contentType); err != nil {
		return fmt.Errorf("failed to unmarshal services: %w", err)
	}

	for _, service := range services {
		priority := collection.GetPriority()
		if service.HasPriority() {
			priority = service.GetPriority()
		}
		if err := m.addService(service, priority); err != nil {
			m.log.Warn("Failed to add service from collection", "service", service.GetName(), "error", err)
		}
	}

	m.log.Info("Successfully loaded and merged upstream service collection", "name", collection.GetName(), "url", url, "services_loaded", len(services))
	return nil
}

func (m *UpstreamServiceManager) unmarshalServices(data []byte, services *[]*configv1.UpstreamServiceConfig, contentType string) error {
	var jsonData []byte
	var err error

	switch contentType {
	case "application/json":
		jsonData = data
	case "application/protobuf", "text/plain":
		var serviceList configv1.Collection
		if err := prototext.Unmarshal(data, &serviceList); err != nil {
			return fmt.Errorf("failed to unmarshal protobuf text: %w", err)
		}
		*services = serviceList.GetServices()
		return nil
	default:
		// Assume YAML
		jsonData, err = yaml.YAMLToJSON(data)
		if err != nil {
			return fmt.Errorf("failed to convert yaml to json: %w", err)
		}
		m.log.Info("Unmarshalled YAML to JSON", "json", string(util.RedactJSON(jsonData)))
	}

	return m.unmarshalProtoJSON(jsonData, services)
}

func (m *UpstreamServiceManager) unmarshalProtoJSON(data []byte, services *[]*configv1.UpstreamServiceConfig) error {
	var serviceList configv1.Collection
	if err := protojson.Unmarshal(data, &serviceList); err != nil {
		// If unmarshalling into a list fails, try unmarshalling as a single service
		var singleService configv1.UpstreamServiceConfig
		if err := protojson.Unmarshal(data, &singleService); err == nil {
			*services = append(*services, &singleService)
			return nil
		}
		return fmt.Errorf("failed to unmarshal services: %w", err)
	}

	if len(serviceList.GetServices()) == 0 {
		var singleService configv1.UpstreamServiceConfig
		if err := protojson.Unmarshal(data, &singleService); err == nil {
			if singleService.GetName() != "" {
				*services = append(*services, &singleService)
				return nil
			}
		}
	}

	// Validate semver
	if version := serviceList.GetVersion(); version != "" {
		if _, err := semver.NewVersion(version); err != nil {
			return fmt.Errorf("invalid semantic version in collection %s: %w", serviceList.GetName(), err)
		}
		m.log.Info("Collection version", "version", version, "name", serviceList.GetName())
	}

	*services = serviceList.GetServices()
	return nil
}

func (m *UpstreamServiceManager) applyAuthentication(req *http.Request, auth *configv1.Authentication) error {
	if auth == nil {
		return nil
	}

	if apiKey := auth.GetApiKey(); apiKey != nil {
		apiKeyValue, err := util.ResolveSecret(req.Context(), apiKey.GetValue())
		if err != nil {
			return err
		}
		req.Header.Set(apiKey.GetParamName(), apiKeyValue)
	} else if bearerToken := auth.GetBearerToken(); bearerToken != nil {
		tokenValue, err := util.ResolveSecret(req.Context(), bearerToken.GetToken())
		if err != nil {
			return err
		}
		req.Header.Set("Authorization", "Bearer "+tokenValue)
	} else if basicAuth := auth.GetBasicAuth(); basicAuth != nil {
		passwordValue, err := util.ResolveSecret(req.Context(), basicAuth.GetPassword())
		if err != nil {
			return err
		}
		req.SetBasicAuth(basicAuth.GetUsername(), passwordValue)
	}

	return nil
}

func (m *UpstreamServiceManager) addService(service *configv1.UpstreamServiceConfig, priority int32) error { //nolint:unparam
	if service == nil {
		return nil
	}

	// Profile filtering is now implicitly done by "enabled state".
	// By default, if a service is not mentioned in any profile config, is it enabled?
	// Or disabled?
	// If we want "centralized management", maybe default is disabled unless enabled in profile?
	// Check overrides first
	m.log.Info("Checking service disabled status", "service", service.GetName(), "disabled_field", service.GetDisable())
	isOverrideDisabled := service.GetDisable()
	var activeConfig *configv1.ProfileServiceConfig

	// We can match by Name or ID. Ideally ID.
	if cfg, ok := m.profileServiceOverrides[service.GetId()]; ok {
		activeConfig = cfg
	} else if cfg, ok := m.profileServiceOverrides[service.GetName()]; ok {
		activeConfig = cfg
	}

	if activeConfig != nil {
		if activeConfig.HasEnabled() {
			isOverrideDisabled = !activeConfig.GetEnabled()
		}
	}

	// If the service has a config error, we skip hydration and adding it to active service lists if it's not safe.
	// But we DO want it in the config list for the UI.
	// The `m.services` map is used to build the final config list.
	// We need to make sure we don't try to use this service for actual operations if it has errors.
	// Since UpstreamServiceManager is responsible for loading config, but downstream components (like mcpany-server)
	// iterate over this config to start services, we need to ensure downstream is safe OR we mark it disabled here.
	if service.HasConfigError() {
		m.log.Warn("Adding service with configuration error (marked disabled)", "service", service.GetName(), "error", service.GetConfigError())
		service.SetDisable(true) // Force disable so it doesn't get started
	} else {
		// Only hydrate secrets if config is valid to avoid potential issues
		HydrateSecretsInService(service, m.profileSecrets)
	}

	// Apply enabled/disabled override
	if isOverrideDisabled {
		m.log.Info("Service disabled by profile override, skipping", "service_name", service.GetName())
		return nil
	}
	// If explicitly enabled by profile, ensure it's added regardless of other conditions (e.g., if it was implicitly disabled by another profile rule, though current logic doesn't have that)
	// For now, if it's explicitly enabled, it just means it passes this check.

	serviceName := service.GetName()
	if existingPriority, exists := m.servicePriorities[serviceName]; exists {
		switch {
		case priority < existingPriority:
			// New service has higher priority, replace the old one
			m.services[serviceName] = service
			m.servicePriorities[serviceName] = priority
			m.log.Info("Replaced service due to higher priority", "service_name", serviceName, "old_priority", existingPriority, "new_priority", priority)
		case priority == existingPriority:
			// Same priority, this is a duplicate
			m.log.Warn("Duplicate service name found, skipping", "service_name", serviceName)
			return nil
		default:
			// lower priority, do nothing
			m.log.Info("Ignoring service due to lower priority", "service_name", serviceName, "existing_priority", existingPriority, "new_priority", priority)
		}
	} else {
		// New service, add it
		m.services[serviceName] = service
		m.servicePriorities[serviceName] = priority
		m.log.Info("Added new service", "service_name", serviceName, "priority", priority)
	}
	return nil
}
