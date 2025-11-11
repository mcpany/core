/*
 * Copyright 2025 Author(s) of MCP Any
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

package config

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"sort"

	"github.com/mcpany/core/pkg/util"
	"github.com/mcpany/core/pkg/logging"
	configv1 "github.com/mcpany/core/proto/config/v1"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/encoding/prototext"
	"sigs.k8s.io/yaml"

	"github.com/Masterminds/semver/v3"
)

// UpstreamServiceManager manages the loading and merging of upstream services.
type UpstreamServiceManager struct {
	log          *slog.Logger
	services     map[string]*configv1.UpstreamServiceConfig
	servicePriorities map[string]int32
}

// NewUpstreamServiceManager creates a new UpstreamServiceManager.
func NewUpstreamServiceManager() *UpstreamServiceManager {
	return &UpstreamServiceManager{
		log:          logging.GetLogger().With("component", "UpstreamServiceManager"),
		services:     make(map[string]*configv1.UpstreamServiceConfig),
		servicePriorities: make(map[string]int32),
	}
}

// LoadAndMergeServices loads all upstream services from the given configuration,
// including local and remote collections, and merges them.
func (m *UpstreamServiceManager) LoadAndMergeServices(ctx context.Context, config *configv1.McpAnyServerConfig) ([]*configv1.UpstreamServiceConfig, error) {
	// Load local services with default priority 0
	for _, service := range config.GetUpstreamServices() {
		m.addService(service, 0)
	}

	// Load and merge remote service collections
	for _, collection := range config.GetUpstreamServiceCollections() {
		if err := m.loadAndMergeCollection(ctx, collection); err != nil {
			m.log.Warn("Failed to load upstream service collection", "name", collection.GetName(), "url", collection.GetHttpUrl(), "error", err)
			// Continue loading other collections even if one fails
		}
	}

	// Return the final list of services
	var services []*configv1.UpstreamServiceConfig
	for _, service := range m.services {
		services = append(services, service)
	}
	sort.Slice(services, func(i, j int) bool {
		return services[i].GetName() < services[j].GetName()
	})
	return services, nil
}

func (m *UpstreamServiceManager) loadAndMergeCollection(ctx context.Context, collection *configv1.UpstreamServiceCollection) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, collection.GetHttpUrl(), nil)
	if err != nil {
		return fmt.Errorf("failed to create http request: %w", err)
	}

	if auth := collection.GetAuthentication(); auth != nil {
		if err := m.applyAuthentication(req, auth); err != nil {
			return fmt.Errorf("failed to apply authentication for collection %s: %w", collection.GetName(), err)
		}
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to fetch collection from url %s: %w", collection.GetHttpUrl(), err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to fetch collection from url %s: status code %d", collection.GetHttpUrl(), resp.StatusCode)
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
		m.addService(service, priority)
	}

	m.log.Info("Successfully loaded and merged upstream service collection", "name", collection.GetName(), "url", collection.GetHttpUrl(), "services_loaded", len(services))
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
		apiKeyValue, err := util.ResolveSecret(apiKey.GetApiKey())
		if err != nil {
			return err
		}
		req.Header.Set(apiKey.GetHeaderName(), apiKeyValue)
	} else if bearerToken := auth.GetBearerToken(); bearerToken != nil {
		tokenValue, err := util.ResolveSecret(bearerToken.GetToken())
		if err != nil {
			return err
		}
		req.Header.Set("Authorization", "Bearer "+tokenValue)
	} else if basicAuth := auth.GetBasicAuth(); basicAuth != nil {
		passwordValue, err := util.ResolveSecret(basicAuth.GetPassword())
		if err != nil {
			return err
		}
		req.SetBasicAuth(basicAuth.GetUsername(), passwordValue)
	}

	return nil
}

func (m *UpstreamServiceManager) addService(service *configv1.UpstreamServiceConfig, priority int32) {
	if service == nil {
		return
	}
	serviceName := service.GetName()
	if existingPriority, exists := m.servicePriorities[serviceName]; exists {
		if priority < existingPriority {
			// New service has higher priority, replace the old one
			m.services[serviceName] = service
			m.servicePriorities[serviceName] = priority
			m.log.Info("Replaced service due to higher priority", "service_name", serviceName, "old_priority", existingPriority, "new_priority", priority)
		} else if priority == existingPriority {
			// Same priority, keep the one loaded first
			m.log.Info("Ignoring service with same priority, keeping the first one loaded", "service_name", serviceName, "priority", priority)
		} else {
			// lower priority, do nothing
			m.log.Info("Ignoring service due to lower priority", "service_name", serviceName, "existing_priority", existingPriority, "new_priority", priority)
		}
	} else {
		// New service, add it
		m.services[serviceName] = service
		m.servicePriorities[serviceName] = priority
		m.log.Info("Added new service", "service_name", serviceName, "priority", priority)
	}
}
