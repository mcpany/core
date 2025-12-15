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
	"strings"

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
//
// Returns a validated `McpAnyServerConfig` or an error if loading or validation
// fails.
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

	if validationErrors := Validate(fileConfig, bt); len(validationErrors) > 0 {
		var errorMessages []string
		for _, e := range validationErrors {
			log.Error("Config validation error", "service", e.ServiceName, "error", e.Err)
			errorMessages = append(errorMessages, fmt.Sprintf("service '%s': %s", e.ServiceName, e.Err.Error()))
		}
		return nil, fmt.Errorf("config validation failed: %s", strings.Join(errorMessages, "; "))
	}

	if len(fileConfig.GetUpstreamServices()) > 0 {
		log.Info("Successfully processed config file", "services", len(fileConfig.GetUpstreamServices()))
	}
	return fileConfig, nil
}
