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
// Summary: Redacts sensitive fields (like API keys) from an UpstreamServiceConfig object to prevent accidental exposure (e.g., in logs or UI).
//
// Parameters:
//   - svc (*configv1.UpstreamServiceConfig): The service configuration to be sanitized.
//
// Returns:
//   None.
//
// Side Effects:
//   - Modifies the `svc` object in-place by replacing sensitive values with redaction markers.
func StripSecretsFromService(svc *configv1.UpstreamServiceConfig) {
	util.StripSecretsFromService(svc)
}

// StripSecretsFromProfile removes sensitive information from the profile definition.
//
// Summary: Redacts sensitive fields from a ProfileDefinition object.
//
// Parameters:
//   - profile (*configv1.ProfileDefinition): The profile definition to be sanitized.
//
// Returns:
//   None.
//
// Side Effects:
//   - Modifies the `profile` object in-place.
func StripSecretsFromProfile(profile *configv1.ProfileDefinition) {
	util.StripSecretsFromProfile(profile)
}

// StripSecretsFromCollection removes sensitive information from the service collection.
//
// Summary: Redacts sensitive fields from a Collection object.
//
// Parameters:
//   - collection (*configv1.Collection): The collection configuration to be sanitized.
//
// Returns:
//   None.
//
// Side Effects:
//   - Modifies the `collection` object in-place.
func StripSecretsFromCollection(collection *configv1.Collection) {
	util.StripSecretsFromCollection(collection)
}

// StripSecretsFromAuth removes sensitive values from the authentication config.
//
// Summary: Redacts sensitive fields from an Authentication object.
//
// Parameters:
//   - auth (*configv1.Authentication): The authentication configuration to be sanitized.
//
// Returns:
//   None.
//
// Side Effects:
//   - Modifies the `auth` object in-place.
func StripSecretsFromAuth(auth *configv1.Authentication) {
	util.StripSecretsFromAuth(auth)
}

// HydrateSecretsInService populates the service configuration with resolved secret values.
//
// Summary: Injects resolved secret values into a service configuration, replacing placeholders or references.
//
// Parameters:
//   - svc (*configv1.UpstreamServiceConfig): The service configuration to be hydrated.
//   - secrets (map[string]*configv1.SecretValue): A map of resolved secrets available for injection.
//
// Returns:
//   None.
//
// Side Effects:
//   - Modifies the `svc` object in-place by injecting actual secret values.
func HydrateSecretsInService(svc *configv1.UpstreamServiceConfig, secrets map[string]*configv1.SecretValue) {
	util.HydrateSecretsInService(svc, secrets)
}
