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
// Summary: Redacts sensitive fields in the service configuration.
//
// Parameters:
//   - svc: *configv1.UpstreamServiceConfig. The service configuration to sanitize.
//
// Side Effects:
//   - Modifies the input configuration in place.
func StripSecretsFromService(svc *configv1.UpstreamServiceConfig) {
	util.StripSecretsFromService(svc)
}

// StripSecretsFromProfile removes sensitive information from the profile definition.
//
// Summary: Redacts sensitive fields in the profile definition.
//
// Parameters:
//   - profile: *configv1.ProfileDefinition. The profile definition to sanitize.
//
// Side Effects:
//   - Modifies the input profile in place.
func StripSecretsFromProfile(profile *configv1.ProfileDefinition) {
	util.StripSecretsFromProfile(profile)
}

// StripSecretsFromCollection removes sensitive information from the service collection.
//
// Summary: Redacts sensitive fields in the collection.
//
// Parameters:
//   - collection: *configv1.Collection. The collection to sanitize.
//
// Side Effects:
//   - Modifies the input collection in place.
func StripSecretsFromCollection(collection *configv1.Collection) {
	util.StripSecretsFromCollection(collection)
}

// StripSecretsFromAuth removes sensitive values from the authentication config.
//
// Summary: Redacts sensitive fields in the authentication configuration.
//
// Parameters:
//   - auth: *configv1.Authentication. The authentication configuration to sanitize.
//
// Side Effects:
//   - Modifies the input authentication config in place.
func StripSecretsFromAuth(auth *configv1.Authentication) {
	util.StripSecretsFromAuth(auth)
}

// HydrateSecretsInService populates the service configuration with resolved secret values.
//
// Summary: Injects resolved secrets into the service configuration.
//
// Parameters:
//   - svc: *configv1.UpstreamServiceConfig. The service configuration to hydrate.
//   - secrets: map[string]*configv1.SecretValue. A map of resolved secrets.
//
// Side Effects:
//   - Modifies the input service configuration in place.
func HydrateSecretsInService(svc *configv1.UpstreamServiceConfig, secrets map[string]*configv1.SecretValue) {
	util.HydrateSecretsInService(svc, secrets)
}
