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
// Summary: removes sensitive information from the service configuration.
//
// Parameters:
//   - svc: *configv1.UpstreamServiceConfig. The svc.
//
// Returns:
//   None.
func StripSecretsFromService(svc *configv1.UpstreamServiceConfig) {
	util.StripSecretsFromService(svc)
}

// StripSecretsFromProfile removes sensitive information from the profile definition.
//
// Summary: removes sensitive information from the profile definition.
//
// Parameters:
//   - profile: *configv1.ProfileDefinition. The profile.
//
// Returns:
//   None.
func StripSecretsFromProfile(profile *configv1.ProfileDefinition) {
	util.StripSecretsFromProfile(profile)
}

// StripSecretsFromCollection removes sensitive information from the service collection.
//
// Summary: removes sensitive information from the service collection.
//
// Parameters:
//   - collection: *configv1.Collection. The collection.
//
// Returns:
//   None.
func StripSecretsFromCollection(collection *configv1.Collection) {
	util.StripSecretsFromCollection(collection)
}

// StripSecretsFromAuth removes sensitive values from the authentication config.
//
// Summary: removes sensitive values from the authentication config.
//
// Parameters:
//   - auth: *configv1.Authentication. The auth.
//
// Returns:
//   None.
func StripSecretsFromAuth(auth *configv1.Authentication) {
	util.StripSecretsFromAuth(auth)
}

// HydrateSecretsInService populates the service configuration with resolved secret values.
//
// Summary: populates the service configuration with resolved secret values.
//
// Parameters:
//   - svc: *configv1.UpstreamServiceConfig. The svc.
//   - secrets: map[string]*configv1.SecretValue. The secrets.
//
// Returns:
//   None.
func HydrateSecretsInService(svc *configv1.UpstreamServiceConfig, secrets map[string]*configv1.SecretValue) {
	util.HydrateSecretsInService(svc, secrets)
}
