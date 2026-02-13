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
// Summary: Redacts secrets from a service configuration.
//
// It specifically targets plain text secrets in UpstreamAuth and other locations.
//
// Parameters:
//   - svc: *configv1.UpstreamServiceConfig. The service configuration to clean.
func StripSecretsFromService(svc *configv1.UpstreamServiceConfig) {
	util.StripSecretsFromService(svc)
}

// StripSecretsFromProfile removes sensitive information from the profile definition.
//
// Summary: Redacts secrets from a profile definition.
//
// Parameters:
//   - profile: *configv1.ProfileDefinition. The profile to clean.
func StripSecretsFromProfile(profile *configv1.ProfileDefinition) {
	util.StripSecretsFromProfile(profile)
}

// StripSecretsFromCollection removes sensitive information from the service collection.
//
// Summary: Redacts secrets from a collection.
//
// Parameters:
//   - collection: *configv1.Collection. The collection to clean.
func StripSecretsFromCollection(collection *configv1.Collection) {
	util.StripSecretsFromCollection(collection)
}

// StripSecretsFromAuth removes sensitive values from the authentication config.
//
// Summary: Redacts secrets from authentication configuration.
//
// Parameters:
//   - auth: *configv1.Authentication. The authentication config to clean.
func StripSecretsFromAuth(auth *configv1.Authentication) {
	util.StripSecretsFromAuth(auth)
}

// HydrateSecretsInService populates the service configuration with resolved secret values.
//
// Summary: Injects resolved secrets into a service configuration.
//
// Parameters:
//   - svc: *configv1.UpstreamServiceConfig. The service configuration to hydrate.
//   - secrets: map[string]*configv1.SecretValue. A map of secrets to inject.
func HydrateSecretsInService(svc *configv1.UpstreamServiceConfig, secrets map[string]*configv1.SecretValue) {
	util.HydrateSecretsInService(svc, secrets)
}
