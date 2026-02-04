// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"context"
	"errors"
	"fmt"
	"strings"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/logging"
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
//
// LoadServices loads, validates, and processes the MCP Any server configuration.
// It acts as a resilient loader that filters out invalid services to allow the server to start
// even with partial configuration failures.
func LoadServices(ctx context.Context, store Store, binaryType string) (*configv1.McpAnyServerConfig, error) {
	log := logging.GetLogger().With("component", "configLoader")

	fileConfig, err := LoadResolvedConfig(ctx, store)
	if err != nil {
		var ae *ActionableError
		if errors.As(err, &ae) {
			var sb strings.Builder
			sb.WriteString("\n❌ Configuration Loading Failed:\n")
			sb.WriteString(fmt.Sprintf("    Error: %v\n", ae.Err))
			sb.WriteString(fmt.Sprintf("    💡 Fix: %s\n", ae.Suggestion))
			return nil, fmt.Errorf("%s", sb.String())
		}
		return nil, err
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

	validationErrors := Validate(ctx, fileConfig, bt)
	if len(validationErrors) > 0 {
		// Map errors to services for logging
		serviceErrors := make(map[string]string)
		for _, e := range validationErrors {
			log.Error("Config validation error", "service", e.ServiceName, "error", e.Err)
			serviceErrors[e.ServiceName] = e.Err.Error()
		}

		// Check global settings errors.
		if _, ok := serviceErrors["global_settings"]; ok {
			return nil, fmt.Errorf("global settings validation failed, check logs for details")
		}

		// Update: For "Friction Fighter" mission, we now fail strictly on startup if any service is invalid.
		// This prevents "silent failures" where a service is ignored but the server starts.
		// We can consider making this configurable later if "resilience" is needed for partial outages.
		if len(validationErrors) > 0 {
			var sb strings.Builder
			sb.WriteString("\n❌ Configuration Validation Failed:\n")

			for _, e := range validationErrors {
				sb.WriteString(fmt.Sprintf("\n  Service: %q\n", e.ServiceName))

				if ae, ok := e.Err.(*ActionableError); ok {
					sb.WriteString(fmt.Sprintf("    Error: %v\n", ae.Err))
					sb.WriteString(fmt.Sprintf("    💡 Fix: %s\n", ae.Suggestion))
				} else {
					sb.WriteString(fmt.Sprintf("    Error: %v\n", e.Err))
				}
			}
			return nil, fmt.Errorf("%s", sb.String())
		}
	}

	if len(fileConfig.GetUpstreamServices()) > 0 {
		log.Info("Successfully processed config file", "services", len(fileConfig.GetUpstreamServices()))
	}
	return fileConfig, nil
}

// LoadResolvedConfig loads key resolved configuration (merging services, setting defaults)
// without performing strict validation or filtering. This is useful for tools that need
// to inspect the configuration (like validate or doc) regardless of validity.
func LoadResolvedConfig(ctx context.Context, store Store) (*configv1.McpAnyServerConfig, error) {
	log := logging.GetLogger().With("component", "configLoader")

	fileConfig, err := store.Load(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to load config from store: %w", err)
	}

	if fileConfig == nil {
		if store.HasConfigSources() {
			return nil, fmt.Errorf("configuration sources provided but loaded configuration is empty. Check if the sources are empty or invalid")
		}
		log.Info("No configuration files found or all were empty, using default configuration.")
		fileConfig = configv1.McpAnyServerConfig_builder{}.Build()
	}

	// Use profiles from config if available, otherwise fall back to global settings
	profiles := fileConfig.GetGlobalSettings().GetProfiles()
	if len(profiles) == 0 {
		profiles = GlobalSettings().Profiles()
	}

	manager := NewUpstreamServiceManager(profiles)
	services, err := manager.LoadAndMergeServices(ctx, fileConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to load and merge services: %w", err)
	}
	fileConfig.SetUpstreamServices(services)

	// If no users are configured, create a default user that has access to all profiles
	if len(fileConfig.GetUsers()) == 0 {
		log.Info("No users configured, creating default user")
		allProfileIDs := make(map[string]bool)

		// Collect from resolved active profiles
		for _, p := range profiles {
			allProfileIDs[p] = true
		}

		// Ensure at least "default"
		allProfileIDs["default"] = true

		var profileIDs []string
		for id := range allProfileIDs {
			profileIDs = append(profileIDs, id)
		}

		apiKey := GlobalSettings().APIKey()
		defaultUser := configv1.User_builder{
			Id:         proto.String("default"),
			ProfileIds: profileIDs,
		}.Build()
		if apiKey != "" {
			headerLoc := configv1.APIKeyAuth_HEADER
			defaultUser.SetAuthentication(configv1.Authentication_builder{
				ApiKey: configv1.APIKeyAuth_builder{
					ParamName:         proto.String("X-API-Key"),
					VerificationValue: proto.String(apiKey),
					In:                &headerLoc,
				}.Build(),
			}.Build())
		}
		fileConfig.SetUsers([]*configv1.User{defaultUser})
	}

	return fileConfig, nil
}
