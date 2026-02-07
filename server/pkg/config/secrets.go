// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

// Package config provides configuration management for MCP Any.
package config

import (
	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/util"
)

// StripSecretsFromService removes sensitive information from the service configuration.
//
// Summary: Sanitizes the UpstreamServiceConfig by redacting secrets to prevent leakage in logs or UI.
//
// Parameters:
//   - svc: *configv1.UpstreamServiceConfig. The service configuration object to sanitize.
//
// Side Effects:
//   - Modifies the provided service configuration in-place.
func StripSecretsFromService(svc *configv1.UpstreamServiceConfig) {
	util.StripSecretsFromService(svc)
}

// StripSecretsFromProfile removes sensitive information from the profile definition.
//
// Summary: Sanitizes the ProfileDefinition by redacting secrets.
//
// Parameters:
//   - profile: *configv1.ProfileDefinition. The profile definition to sanitize.
//
// Side Effects:
//   - Modifies the provided profile definition in-place.
func StripSecretsFromProfile(profile *configv1.ProfileDefinition) {
	util.StripSecretsFromProfile(profile)
}

// StripSecretsFromCollection removes sensitive information from the service collection.
//
// Summary: Sanitizes the Collection configuration by redacting secrets.
//
// Parameters:
//   - collection: *configv1.Collection. The collection configuration to sanitize.
//
// Side Effects:
//   - Modifies the provided collection configuration in-place.
func StripSecretsFromCollection(collection *configv1.Collection) {
	util.StripSecretsFromCollection(collection)
}

// StripSecretsFromAuth removes sensitive values from the authentication config.
//
// Summary: Sanitizes the Authentication configuration by redacting secrets (API keys, tokens, passwords).
//
// Parameters:
//   - auth: *configv1.Authentication. The authentication configuration to sanitize.
//
// Side Effects:
//   - Modifies the provided authentication configuration in-place.
func StripSecretsFromAuth(auth *configv1.Authentication) {
	util.StripSecretsFromAuth(auth)
}

// HydrateSecretsInService populates the service configuration with resolved secret values.
//
// Summary: Injects actual secret values into the service configuration from the provided secrets map.
// This is typically used when preparing a service for execution.
//
// Parameters:
//   - svc: *configv1.UpstreamServiceConfig. The service configuration to hydrate.
//   - secrets: map[string]*configv1.SecretValue. A map of resolved secrets to use for hydration.
//
// Side Effects:
//   - Modifies the provided service configuration in-place by setting secret values.
func HydrateSecretsInService(svc *configv1.UpstreamServiceConfig, secrets map[string]*configv1.SecretValue) {
	util.HydrateSecretsInService(svc, secrets)
}
