// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

// Package config provides configuration management for MCP Any.
package config

import (
	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/util"
)

// StripSecretsFromService removes sensitive information from the service configuration.
//
// Summary: Redacts sensitive fields (like API keys) from a service configuration.
// It specifically targets plain text secrets in UpstreamAuth and other locations
// to ensure they are not logged or exposed in API responses.
//
// Parameters:
//   - svc: *configv1.UpstreamServiceConfig. The service configuration to sanitize.
//
// Side Effects:
//   - Modifies the input service configuration object in place.
func StripSecretsFromService(svc *configv1.UpstreamServiceConfig) {
	util.StripSecretsFromService(svc)
}

// StripSecretsFromProfile removes sensitive information from the profile definition.
//
// Summary: Redacts sensitive fields (like API keys) from a profile definition.
//
// Parameters:
//   - profile: *configv1.ProfileDefinition. The profile definition to sanitize.
//
// Side Effects:
//   - Modifies the input profile object in place.
func StripSecretsFromProfile(profile *configv1.ProfileDefinition) {
	util.StripSecretsFromProfile(profile)
}

// StripSecretsFromCollection removes sensitive information from the service collection.
//
// Summary: Redacts sensitive fields from a service collection configuration.
//
// Parameters:
//   - collection: *configv1.Collection. The collection configuration to sanitize.
//
// Side Effects:
//   - Modifies the input collection object in place.
func StripSecretsFromCollection(collection *configv1.Collection) {
	util.StripSecretsFromCollection(collection)
}

// StripSecretsFromAuth removes sensitive values from the authentication config.
//
// Summary: Redacts sensitive fields (API keys, tokens, passwords) from an authentication configuration.
//
// Parameters:
//   - auth: *configv1.Authentication. The authentication configuration to sanitize.
//
// Side Effects:
//   - Modifies the input authentication object in place.
func StripSecretsFromAuth(auth *configv1.Authentication) {
	util.StripSecretsFromAuth(auth)
}

// HydrateSecretsInService populates the service configuration with resolved secret values.
//
// Summary: Injects resolved secret values into a service configuration.
// This is typically used when loading a configuration to ensure that services have
// access to the necessary credentials.
//
// Parameters:
//   - svc: *configv1.UpstreamServiceConfig. The service configuration to hydrate.
//   - secrets: map[string]*configv1.SecretValue. A map of resolved secrets.
//
// Side Effects:
//   - Modifies the input service configuration object in place.
func HydrateSecretsInService(svc *configv1.UpstreamServiceConfig, secrets map[string]*configv1.SecretValue) {
	util.HydrateSecretsInService(svc, secrets)
}
