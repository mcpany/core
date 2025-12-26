// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"context"
	"fmt"

	"github.com/mcpany/core/pkg/logging"
	configv1 "github.com/mcpany/core/proto/config/v1"
	"google.golang.org/protobuf/proto"
)

// LoadServices loads, validates, and processes the MCP Any server configuration
// from a given store. It orchestrates the reading of the configuration,
// validates its contents, and returns a sanitized configuration object.
//
// If the provided store is empty or contains no configuration files, a default,
// empty configuration is returned.
//
// Parameters:
//   - store: The configuration store from which to load the configuration.
//   - binaryType: The type of binary running the code (e.g., "server", "worker").
//
// Returns:
//   - A validated `McpAnyServerConfig` object.
//   - An error if loading or validation fails.
func LoadServices(store Store, binaryType string) (*configv1.McpAnyServerConfig, error) {
	log := logging.GetLogger().With("component", "configLoader")

	fileConfig, err := store.Load()
	if err != nil {
		return nil, fmt.Errorf("failed to load config from store: %w", err)
	}

	if fileConfig == nil {
		log.Info("No configuration files found or all were empty, using default configuration.")
		fileConfig = &configv1.McpAnyServerConfig{}
	}

	// Use profiles from config if available, otherwise fall back to global settings
	profiles := fileConfig.GetGlobalSettings().GetProfiles()
	if len(profiles) == 0 {
		profiles = GlobalSettings().Profiles()
	}

	manager := NewUpstreamServiceManager(profiles)
	services, err := manager.LoadAndMergeServices(context.Background(), fileConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to load and merge services: %w", err)
	}
	fileConfig.SetUpstreamServices(services)

	// If no users are configured, create a default user that has access to all profiles
	if len(fileConfig.GetUsers()) == 0 {
		log.Info("No users configured, creating default user")
		allProfileIDs := make(map[string]bool)
		for _, svc := range fileConfig.GetUpstreamServices() {
			for _, p := range svc.GetProfiles() {
				if p.GetId() != "" {
					allProfileIDs[p.GetId()] = true
				}
			}
		}
		var profileIDs []string
		for id := range allProfileIDs {
			profileIDs = append(profileIDs, id)
		}

		apiKey := GlobalSettings().APIKey()
		defaultUser := &configv1.User{
			Id:         proto.String("default"),
			ProfileIds: profileIDs,
		}
		if apiKey != "" {
			headerLoc := configv1.APIKeyAuth_HEADER
			defaultUser.Authentication = &configv1.AuthenticationConfig{
				AuthMethod: &configv1.AuthenticationConfig_ApiKey{
					ApiKey: &configv1.APIKeyAuth{
						ParamName: proto.String("X-API-Key"),
						KeyValue:  proto.String(apiKey),
						In:        &headerLoc,
					},
				},
			}
		}
		fileConfig.Users = []*configv1.User{defaultUser}
	}

	var bt BinaryType
	switch binaryType {
	case "server":
		bt = Server
	case "worker":
		bt = Worker
	default:
		log.Error("Unknown binary type", "binary_type", binaryType)
		return nil, fmt.Errorf("unknown binary type: %s", binaryType)
	}

	if validationErrors := Validate(context.Background(), fileConfig, bt); len(validationErrors) > 0 {
		invalidServices := make(map[string]bool)
		for _, e := range validationErrors {
			// Some validation errors might not have a service name (e.g. global settings)
			// In that case, we should probably fail.
			if e.ServiceName == "global_settings" || e.ServiceName == "" {
				log.Error("Critical config validation error", "error", e.Err)
				return nil, fmt.Errorf("critical config validation failed: %s", e.Err.Error())
			}

			log.Error("Config validation error, skipping service", "service", e.ServiceName, "error", e.Err)
			invalidServices[e.ServiceName] = true
		}

		// Filter out invalid services
		if len(invalidServices) > 0 {
			var validServices []*configv1.UpstreamServiceConfig
			for _, svc := range fileConfig.GetUpstreamServices() {
				if !invalidServices[svc.GetName()] {
					validServices = append(validServices, svc)
				}
			}
			fileConfig.UpstreamServices = validServices
		}
	}

	if len(fileConfig.GetUpstreamServices()) > 0 {
		log.Info("Successfully processed config file", "services", len(fileConfig.GetUpstreamServices()))
	}
	return fileConfig, nil
}
