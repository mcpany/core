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

	"github.com/mcpany/core/pkg/logging"
	"github.com/mcpany/core/pkg/util"
	configv1 "github.com/mcpany/core/proto/config/v1"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/encoding/prototext"
	"google.golang.org/protobuf/proto"
	"sigs.k8s.io/yaml"

	"github.com/Masterminds/semver/v3"
)

// UpstreamServiceManager manages the lifecycle and configuration of upstream services.
// It handles loading, validating, and merging service configurations from various sources,
// including local files and remote URLs (e.g., GitHub).
type UpstreamServiceManager struct {
	log               *slog.Logger
	services          map[string]*configv1.UpstreamServiceConfig
	servicePriorities map[string]int32
	httpClient        *http.Client
	newGitHub         func(ctx context.Context, rawURL string) (*GitHub, error)
	enabledProfiles   []string
}

// NewUpstreamServiceManager creates a new instance of UpstreamServiceManager.
// It initializes the manager with the specified enabled profiles and default settings.
//
// Parameters:
//   enabledProfiles: A list of profile names that are active. Services must match one of these profiles to be loaded.
//
// Returns:
//   A pointer to a fully initialized UpstreamServiceManager.
func NewUpstreamServiceManager(enabledProfiles []string) *UpstreamServiceManager {
	if len(enabledProfiles) == 0 {
		enabledProfiles = []string{"default"}
	}
	return &UpstreamServiceManager{
		log:               logging.GetLogger().With("component", "UpstreamServiceManager"),
		services:          make(map[string]*configv1.UpstreamServiceConfig),
		servicePriorities: make(map[string]int32),
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
// It processes both locally defined services and remote service collections, merging them
// based on their priority and name.
//
// Parameters:
//   ctx: The context for the operation.
//   config: The main server configuration containing service definitions and collection references.
//
// Returns:
//   A slice of pointers to UpstreamServiceConfig objects that represent the final set of loaded services.
//   An error if any critical failure occurs during loading or merging.
func (m *UpstreamServiceManager) LoadAndMergeServices(ctx context.Context, config *configv1.McpAnyServerConfig) ([]*configv1.UpstreamServiceConfig, error) {
	// Load local services with default priority 0
	for _, service := range config.GetUpstreamServices() {
		if err := m.addService(service, 0); err != nil {
			return nil, err
		}
	}

	// Load and merge remote service collections
	for _, collection := range config.GetUpstreamServiceCollections() {
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

func (m *UpstreamServiceManager) loadAndMergeCollection(ctx context.Context, collection *configv1.UpstreamServiceCollection) error {
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
				newCollection := configv1.UpstreamServiceCollection_builder{
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

func (m *UpstreamServiceManager) loadFromURL(ctx context.Context, url string, collection *configv1.UpstreamServiceCollection) error {
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
		var serviceList configv1.UpstreamServiceCollectionShare
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
	}

	return m.unmarshalProtoJSON(jsonData, services)
}

func (m *UpstreamServiceManager) unmarshalProtoJSON(data []byte, services *[]*configv1.UpstreamServiceConfig) error {
	var serviceList configv1.UpstreamServiceCollectionShare
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

func (m *UpstreamServiceManager) applyAuthentication(req *http.Request, auth *configv1.UpstreamAuthentication) error {
	if auth == nil {
		return nil
	}

	if apiKey := auth.GetApiKey(); apiKey != nil {
		apiKeyValue, err := util.ResolveSecret(req.Context(), apiKey.GetApiKey())
		if err != nil {
			return err
		}
		req.Header.Set(apiKey.GetHeaderName(), apiKeyValue)
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

func (m *UpstreamServiceManager) addService(service *configv1.UpstreamServiceConfig, priority int32) error {
	if service == nil {
		return nil
	}

	// Filter by profile
	serviceProfiles := service.GetProfiles()
	if len(serviceProfiles) == 0 {
		// Default to "default" profile if none specified
		serviceProfiles = []*configv1.Profile{{Name: "default"}}
	}

	allowed := false
	for _, sp := range serviceProfiles {
		// Normalize profile: default ID to Name if missing
		if sp.GetId() == "" && sp.GetName() != "" {
			sp.Id = sp.GetName()
		}
		for _, ep := range m.enabledProfiles {
			if sp.GetName() == ep || sp.GetId() == ep {
				allowed = true
				break
			}
		}
		if allowed {
			break
		}
	}

	if !allowed {
		m.log.Debug("Skipping service due to profile mismatch", "service_name", service.GetName(), "service_profiles", serviceProfiles, "enabled_profiles", m.enabledProfiles)
		return nil
	}

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
			return fmt.Errorf("duplicate service name found: %s", serviceName)
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
