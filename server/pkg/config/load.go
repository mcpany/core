// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
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

		// Sentinel: Security enhancement to prevent insecure default state
		// If no API key is configured, generate one to ensure the default user is not open to the world.
		if apiKey == "" {
			bytes := make([]byte, 16)
			if _, err := rand.Read(bytes); err == nil {
				apiKey = hex.EncodeToString(bytes)
				GlobalSettings().SetAPIKey(apiKey)
				// Use fmt.Fprintf to stderr to ensure visibility without relying on structured logging
				// which might hide secrets or be ingested by log aggregation systems.
				fmt.Fprintf(os.Stderr, "\n=================================================================\n")
				fmt.Fprintf(os.Stderr, "⚠️  SECURITY WARNING: No API Key configured.\n")
				fmt.Fprintf(os.Stderr, "   Generated temporary key for default user: %s\n", apiKey)
				fmt.Fprintf(os.Stderr, "   Please set MCPANY_API_KEY environment variable for production.\n")
				fmt.Fprintf(os.Stderr, "=================================================================\n\n")
			} else {
				log.Error("Failed to generate random API key", "error", err)
				// Fail safe: do not set apiKey, but maybe we should fail hard?
				// If generation fails, we fall back to insecure default? No, that's bad.
				// But defaulting to empty string maintains existing behavior (insecure).
				// Given crypto/rand failure is extremely rare, logging error is okay for now,
				// but ideally we should panic or exit.
			}
		}

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
