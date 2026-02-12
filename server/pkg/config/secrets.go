// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

// Package config provides configuration management for MCP Any.
package config

import (
	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/util"
)

// StripSecretsFromService removes sensitive information from the service configuration.
// It specifically targets plain text secrets in UpstreamAuth and other locations
// to prevent them from being leaked in logs or API responses.
//
// Parameters:
//   - svc: *configv1.UpstreamServiceConfig. The service configuration to sanitize.
//
// Side Effects:
//   - Modifies the provided svc object in place.
func StripSecretsFromService(svc *configv1.UpstreamServiceConfig) {
	util.StripSecretsFromService(svc)
}

// StripSecretsFromProfile removes sensitive information from the profile definition.
// This ensures that API keys or other secrets embedded in profile definitions
// are redacted before being exposed.
//
// Parameters:
//   - profile: *configv1.ProfileDefinition. The profile definition to sanitize.
//
// Side Effects:
//   - Modifies the provided profile object in place.
func StripSecretsFromProfile(profile *configv1.ProfileDefinition) {
	util.StripSecretsFromProfile(profile)
}

// StripSecretsFromCollection removes sensitive information from the service collection.
// It targets authentication details within the collection configuration.
//
// Parameters:
//   - collection: *configv1.Collection. The collection configuration to sanitize.
//
// Side Effects:
//   - Modifies the provided collection object in place.
func StripSecretsFromCollection(collection *configv1.Collection) {
	util.StripSecretsFromCollection(collection)
}

// StripSecretsFromAuth removes sensitive values from the authentication configuration.
// It clears API keys, tokens, and passwords.
//
// Parameters:
//   - auth: *configv1.Authentication. The authentication configuration to sanitize.
//
// Side Effects:
//   - Modifies the provided auth object in place.
func StripSecretsFromAuth(auth *configv1.Authentication) {
	util.StripSecretsFromAuth(auth)
}

// HydrateSecretsInService populates the service configuration with resolved secret values.
// It replaces references to secrets (e.g., environment variables or vault paths)
// with their actual values for runtime use.
//
// Parameters:
//   - svc: *configv1.UpstreamServiceConfig. The service configuration to hydrate.
//   - secrets: map[string]*configv1.SecretValue. A map of resolved secrets where keys match the secret references.
//
// Side Effects:
//   - Modifies the provided svc object in place by injecting secret values.
func HydrateSecretsInService(svc *configv1.UpstreamServiceConfig, secrets map[string]*configv1.SecretValue) {
	util.HydrateSecretsInService(svc, secrets)
}
