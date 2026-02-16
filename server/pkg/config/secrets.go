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
// Parameters:
//   - svc: *configv1.UpstreamServiceConfig. The upstream service configuration to strip secrets from.
//
// Returns:
//   None.
//
// Side Effects:
//   - Modifies the provided service configuration in place.
func StripSecretsFromService(svc *configv1.UpstreamServiceConfig) {
	util.StripSecretsFromService(svc)
}

// StripSecretsFromProfile removes sensitive information from the profile definition.
//
// Summary: Redacts secrets from a profile definition.
//
// Parameters:
//   - profile: *configv1.ProfileDefinition. The profile definition.
//
// Returns:
//   None.
//
// Side Effects:
//   - Modifies the provided profile definition in place.
func StripSecretsFromProfile(profile *configv1.ProfileDefinition) {
	util.StripSecretsFromProfile(profile)
}

// StripSecretsFromCollection removes sensitive information from the service collection.
//
// Summary: Redacts secrets from a service collection.
//
// Parameters:
//   - collection: *configv1.Collection. The service collection.
//
// Returns:
//   None.
//
// Side Effects:
//   - Modifies the provided collection in place.
func StripSecretsFromCollection(collection *configv1.Collection) {
	util.StripSecretsFromCollection(collection)
}

// StripSecretsFromAuth removes sensitive values from the authentication config.
//
// Summary: Redacts secrets from an authentication configuration.
//
// Parameters:
//   - auth: *configv1.Authentication. The authentication configuration.
//
// Returns:
//   None.
//
// Side Effects:
//   - Modifies the provided authentication configuration in place.
func StripSecretsFromAuth(auth *configv1.Authentication) {
	util.StripSecretsFromAuth(auth)
}

// HydrateSecretsInService populates the service configuration with resolved secret values.
//
// Summary: Injects resolved secrets into a service configuration.
//
// Parameters:
//   - svc: *configv1.UpstreamServiceConfig. The service configuration.
//   - secrets: map[string]*configv1.SecretValue. The resolved secrets.
//
// Returns:
//   None.
//
// Side Effects:
//   - Modifies the provided service configuration in place.
func HydrateSecretsInService(svc *configv1.UpstreamServiceConfig, secrets map[string]*configv1.SecretValue) {
	util.HydrateSecretsInService(svc, secrets)
}
