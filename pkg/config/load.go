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
	"fmt"

	"github.com/mcpany/core/pkg/logging"
	configv1 "github.com/mcpany/core/proto/config/v1"
)

// LoadServices loads, validates, and processes the MCP-X server configuration
// from a given store. It orchestrates the reading of the configuration,
// validates its contents, and returns a sanitized configuration object.
//
// If the provided store is empty or contains no configuration files, a default,
// empty configuration is returned.
//
// Parameters:
//   - store: The configuration store from which to load the configuration.
//
// Returns a validated `McpxServerConfig` or an error if loading or validation
// fails.
func LoadServices(store Store, binaryType string) (*configv1.McpxServerConfig, error) {
	log := logging.GetLogger().With("component", "configLoader")

	fileConfig, err := store.Load()
	if err != nil {
		return nil, fmt.Errorf("failed to load config from store: %w", err)
	}

	if fileConfig == nil {
		log.Info("No configuration files found or all were empty, using default configuration.")
		fileConfig = &configv1.McpxServerConfig{}
	}

	var bt BinaryType
	if binaryType == "server" {
		bt = Server
	} else if binaryType == "worker" {
		bt = Worker
	}

	if validationErrors := Validate(fileConfig, bt); len(validationErrors) > 0 {
		for _, e := range validationErrors {
			log.Error("Config validation error", "service", e.ServiceName, "error", e.Err)
		}
		return nil, fmt.Errorf("config validation failed")
	}

	if len(fileConfig.GetUpstreamServices()) > 0 {
		log.Info("Successfully processed config file", "services", len(fileConfig.GetUpstreamServices()))
	}
	return fileConfig, nil
}
