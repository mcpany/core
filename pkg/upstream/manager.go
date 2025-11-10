package upstream

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"sort"
	"strings"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/encoding/prototext"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"gopkg.in/yaml.v3"
)

// serviceWithPriority holds an UpstreamServiceConfig and its priority.
type serviceWithPriority struct {
	config   *configv1.UpstreamServiceConfig
	priority int32
}

// UpstreamServiceManager manages the loading and merging of upstream services.
type UpstreamServiceManager struct {
	services map[string]*serviceWithPriority
}

// NewUpstreamServiceManager creates a new UpstreamServiceManager.
func NewUpstreamServiceManager() *UpstreamServiceManager {
	return &UpstreamServiceManager{
		services: make(map[string]*serviceWithPriority),
	}
}

// Load processes the server config to load and merge all upstream services.
// It returns the final merged list of services.
func (m *UpstreamServiceManager) Load(config *configv1.McpxServerConfig) ([]*configv1.UpstreamServiceConfig, error) {
	// Load locally defined services with default priority 0.
	for _, serviceConfig := range config.GetUpstreamServices() {
		m.addService(serviceConfig, 0)
	}

	// Load services from remote collections.
	for _, collection := range config.GetUpstreamServiceCollections() {
		if err := m.loadFromCollection(collection); err != nil {
			log.Printf("Error loading collection '%s' from '%s': %v. Skipping.", collection.GetName(), collection.GetHttpUrl(), err)
			// Continue loading other collections even if one fails.
		}
	}

	return m.GetServices(), nil
}

// GetServices returns the final list of merged UpstreamServiceConfig.
func (m *UpstreamServiceManager) GetServices() []*configv1.UpstreamServiceConfig {
	var serviceList []*configv1.UpstreamServiceConfig
	for _, s := range m.services {
		serviceList = append(serviceList, s.config)
	}
	// Sort by name for deterministic output.
	sort.Slice(serviceList, func(i, j int) bool {
		return serviceList[i].GetName() < serviceList[j].GetName()
	})
	return serviceList
}

// addService adds a service to the manager, handling priority-based merging.
func (m *UpstreamServiceManager) addService(serviceConfig *configv1.UpstreamServiceConfig, collectionPriority int32) {
	if serviceConfig == nil || serviceConfig.GetName() == "" {
		log.Println("Skipping upstream service with nil config or empty name")
		return
	}

	// Determine the service's priority. The service's own priority field takes precedence.
	priority := collectionPriority
	if serviceConfig.HasPriority() {
		priority = serviceConfig.GetPriority()
	}

	existing, exists := m.services[serviceConfig.GetName()]
	if !exists {
		// Service doesn't exist, add it.
		m.services[serviceConfig.GetName()] = &serviceWithPriority{
			config:   serviceConfig,
			priority: priority,
		}
		log.Printf("Loaded service '%s' with priority %d", serviceConfig.GetName(), priority)
	} else {
		// Service exists, check priority.
		if priority < existing.priority {
			// New service has higher priority (lower number), replace existing.
			m.services[serviceConfig.GetName()] = &serviceWithPriority{
				config:   serviceConfig,
				priority: priority,
			}
			log.Printf("Replaced service '%s' with higher priority version (new: %d, old: %d)", serviceConfig.GetName(), priority, existing.priority)
		} else if priority == existing.priority {
			// Same priority, keep the first one loaded.
			log.Printf("Ignoring duplicate service '%s' with same priority %d", serviceConfig.GetName(), priority)
		} else {
			// Lower priority, ignore the new one.
			log.Printf("Ignoring duplicate service '%s' with lower priority version (new: %d, old: %d)", serviceConfig.GetName(), priority, existing.priority)
		}
	}
}

// loadFromCollection fetches and processes an UpstreamServiceCollection.
func (m *UpstreamServiceManager) loadFromCollection(collection *configv1.UpstreamServiceCollection) error {
	if collection == nil || collection.GetHttpUrl() == "" {
		return fmt.Errorf("collection is nil or http_url is empty")
	}

	log.Printf("Loading upstream service collection '%s' from %s", collection.GetName(), collection.GetHttpUrl())

	req, err := http.NewRequest("GET", collection.GetHttpUrl(), nil)
	if err != nil {
		return fmt.Errorf("failed to create request for %s: %w", collection.GetHttpUrl(), err)
	}

	if err := setAuth(req, collection.GetUpstreamAuthentication()); err != nil {
		return fmt.Errorf("failed to set authentication for %s: %w", collection.GetHttpUrl(), err)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to fetch collection from %s: %w", collection.GetHttpUrl(), err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("received non-OK status code %d from %s", resp.StatusCode, collection.GetHttpUrl())
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body from %s: %w", collection.GetHttpUrl(), err)
	}

	remoteServices, err := parseServices(body, resp.Header.Get("Content-Type"))
	if err != nil {
		return fmt.Errorf("failed to parse services from %s: %w", collection.GetHttpUrl(), err)
	}

	log.Printf("Successfully fetched and parsed %d services from collection '%s'", len(remoteServices), collection.GetName())
	for _, serviceConfig := range remoteServices {
		m.addService(serviceConfig, collection.GetPriority())
	}

	return nil
}

// setAuth sets the authentication headers on the request.
func setAuth(req *http.Request, auth *configv1.UpstreamAuthentication) error {
	if auth == nil {
		return nil
	}
	if apiKeyAuth := auth.GetApiKey(); apiKeyAuth != nil {
		req.Header.Set(apiKeyAuth.GetHeaderName(), apiKeyAuth.GetApiKey())
	} else if bearerTokenAuth := auth.GetBearerToken(); bearerTokenAuth != nil {
		req.Header.Set("Authorization", "Bearer "+bearerTokenAuth.GetToken())
	} else if basicAuth := auth.GetBasicAuth(); basicAuth != nil {
		req.SetBasicAuth(basicAuth.GetUsername(), basicAuth.GetPassword())
	}
	return nil
}

// parseServices parses the response body based on the content type.
func parseServices(body []byte, contentType string) ([]*configv1.UpstreamServiceConfig, error) {
	var remoteServices []*configv1.UpstreamServiceConfig

	unmarshal := func(rawService map[string]any) (*configv1.UpstreamServiceConfig, error) {
		jsonBytes, err := json.Marshal(rawService)
		if err != nil {
			return nil, fmt.Errorf("error marshaling to JSON: %w", err)
		}
		serviceConfig := &configv1.UpstreamServiceConfig{}
		if err := protojson.Unmarshal(jsonBytes, serviceConfig); err != nil {
			return nil, fmt.Errorf("error unmarshaling to proto: %w", err)
		}
		return serviceConfig, nil
	}

	switch {
	case strings.Contains(contentType, "application/x-prototext"):
		messages := strings.Split(string(body), "---")
		for _, msgStr := range messages {
			if strings.TrimSpace(msgStr) == "" {
				continue
			}
			serviceConfig := &configv1.UpstreamServiceConfig{}
			if err := prototext.Unmarshal([]byte(msgStr), serviceConfig); err != nil {
				log.Printf("Error unmarshaling prototext message: %v", err)
				continue
			}
			remoteServices = append(remoteServices, serviceConfig)
		}

	case strings.Contains(contentType, "application/json"):
		var rawServices []map[string]any
		if err := json.Unmarshal(body, &rawServices); err != nil {
			return nil, fmt.Errorf("failed to unmarshal json: %w", err)
		}
		for _, rawService := range rawServices {
			serviceConfig, err := unmarshal(rawService)
			if err != nil {
				log.Printf("Error processing raw service: %v", err)
				continue
			}
			remoteServices = append(remoteServices, serviceConfig)
		}

	default: // Assume YAML by default
		var rawServices []map[string]any
		if err := yaml.Unmarshal(body, &rawServices); err != nil {
			return nil, fmt.Errorf("failed to unmarshal yaml: %w", err)
		}
		for _, rawService := range rawServices {
			serviceConfig, err := unmarshal(rawService)
			if err != nil {
				log.Printf("Error processing raw service: %v", err)
				continue
			}
			remoteServices = append(remoteServices, serviceConfig)
		}
	}
	return remoteServices, nil
}
