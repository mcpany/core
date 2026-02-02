// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

// Package config provides configuration management for MCP Any.
package config

import (
	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/util"
)

// StripSecretsFromService removes sensitive information from the service configuration.
// It specifically targets plain text secrets in UpstreamAuth and other locations.
//
// Summary: Removes plain text secrets from a service configuration.
//
// Parameters:
//   - svc: *configv1.UpstreamServiceConfig. The service configuration to strip.
//
// Returns:
//   - None.
//
// Errors/Throws:
//   - None.
//
// Side Effects:
//   - Modifies the svc object in place.
func StripSecretsFromService(svc *configv1.UpstreamServiceConfig) {
	util.StripSecretsFromService(svc)
}

// StripSecretsFromProfile removes sensitive information from the profile definition.
//
// Summary: Removes plain text secrets from a profile definition.
//
// Parameters:
//   - profile: *configv1.ProfileDefinition. The profile definition to strip.
//
// Returns:
//   - None.
//
// Errors/Throws:
//   - None.
//
// Side Effects:
//   - Modifies the profile object in place.
func StripSecretsFromProfile(profile *configv1.ProfileDefinition) {
	util.StripSecretsFromProfile(profile)
}

// StripSecretsFromCollection removes sensitive information from the service collection.
//
// Summary: Removes plain text secrets from a service collection.
//
// Parameters:
//   - collection: *configv1.Collection. The collection to strip.
//
// Returns:
//   - None.
//
// Errors/Throws:
//   - None.
//
// Side Effects:
//   - Modifies the collection object in place.
func StripSecretsFromCollection(collection *configv1.Collection) {
	util.StripSecretsFromCollection(collection)
}

// StripSecretsFromAuth removes sensitive values from the authentication config.
//
// Summary: Removes plain text secrets from an authentication configuration.
//
// Parameters:
//   - auth: *configv1.Authentication. The authentication config to strip.
//
// Returns:
//   - None.
//
// Errors/Throws:
//   - None.
//
// Side Effects:
//   - Modifies the auth object in place.
func StripSecretsFromAuth(auth *configv1.Authentication) {
	util.StripSecretsFromAuth(auth)
}

// HydrateSecretsInService populates the service configuration with resolved secret values.
//
// Summary: Hydrates secrets in a service configuration using provided values.
//
// Parameters:
//   - svc: *configv1.UpstreamServiceConfig. The service configuration to hydrate.
//   - secrets: map[string]*configv1.SecretValue. A map of resolved secrets.
//
// Returns:
//   - None.
//
// Errors/Throws:
//   - None.
//
// Side Effects:
//   - Modifies the svc object in place.
func HydrateSecretsInService(svc *configv1.UpstreamServiceConfig, secrets map[string]*configv1.SecretValue) {
	util.HydrateSecretsInService(svc, secrets)
}
