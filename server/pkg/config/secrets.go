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
// Summary: Removes sensitive information from service configuration.
//
// Parameters:
//   - svc (*configv1.UpstreamServiceConfig): The upstream service configuration.
func StripSecretsFromService(svc *configv1.UpstreamServiceConfig) {
	util.StripSecretsFromService(svc)
}

// StripSecretsFromProfile removes sensitive information from the profile definition.
//
// Summary: Removes sensitive information from profile definition.
//
// Parameters:
//   - profile (*configv1.ProfileDefinition): The profile definition.
func StripSecretsFromProfile(profile *configv1.ProfileDefinition) {
	util.StripSecretsFromProfile(profile)
}

// StripSecretsFromCollection removes sensitive information from the service collection.
//
// Summary: Removes sensitive information from service collection.
//
// Parameters:
//   - collection (*configv1.Collection): The service collection.
func StripSecretsFromCollection(collection *configv1.Collection) {
	util.StripSecretsFromCollection(collection)
}

// StripSecretsFromAuth removes sensitive values from the authentication config.
//
// Summary: Removes sensitive values from authentication config.
//
// Parameters:
//   - auth (*configv1.Authentication): The authentication configuration.
func StripSecretsFromAuth(auth *configv1.Authentication) {
	util.StripSecretsFromAuth(auth)
}

// HydrateSecretsInService populates the service configuration with resolved secret values.
//
// Summary: Populates service configuration with resolved secret values.
//
// Parameters:
//   - svc (*configv1.UpstreamServiceConfig): The upstream service configuration.
//   - secrets (map[string]*configv1.SecretValue): The resolved secret values.
func HydrateSecretsInService(svc *configv1.UpstreamServiceConfig, secrets map[string]*configv1.SecretValue) {
	util.HydrateSecretsInService(svc, secrets)
}
