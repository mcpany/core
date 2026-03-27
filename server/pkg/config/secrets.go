// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

// Package config provides configuration management for MCP Any.
package config

import (
	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/util"
)

// StripSecretsFromService provides stripsecretsfromservice functionality.
//
// Summary: StripSecretsFromService.
//
// Parameters.
//   - svc: The parameter.
//
// Returns.
//   - None.
func StripSecretsFromService(svc *configv1.UpstreamServiceConfig) {
	util.StripSecretsFromService(svc)
}

// StripSecretsFromProfile provides stripsecretsfromprofile functionality.
//
// Summary: StripSecretsFromProfile.
//
// Parameters.
//   - profile: The parameter.
//
// Returns.
//   - None.
func StripSecretsFromProfile(profile *configv1.ProfileDefinition) {
	util.StripSecretsFromProfile(profile)
}

// StripSecretsFromCollection provides stripsecretsfromcollection functionality.
//
// Summary: StripSecretsFromCollection.
//
// Parameters.
//   - collection: The parameter.
//
// Returns.
//   - None.
func StripSecretsFromCollection(collection *configv1.Collection) {
	util.StripSecretsFromCollection(collection)
}

// StripSecretsFromAuth provides stripsecretsfromauth functionality.
//
// Summary: StripSecretsFromAuth.
//
// Parameters.
//   - auth: The parameter.
//
// Returns.
//   - None.
func StripSecretsFromAuth(auth *configv1.Authentication) {
	util.StripSecretsFromAuth(auth)
}

// HydrateSecretsInService provides hydratesecretsinservice functionality.
//
// Summary: HydrateSecretsInService.
//
// Parameters.
//   - svc: The parameter.
//   - secrets: The parameter.
//
// Returns.
//   - None.
func HydrateSecretsInService(svc *configv1.UpstreamServiceConfig, secrets map[string]*configv1.SecretValue) {
	util.HydrateSecretsInService(svc, secrets)
}
