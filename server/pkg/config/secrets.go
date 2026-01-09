// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package config

import (
	configv1 "github.com/mcpany/core/proto/config/v1"
	"google.golang.org/protobuf/proto"
)

// StripSecretsFromService removes sensitive information from the service configuration.
// It specifically targets plain text secrets in UpstreamAuth.
func StripSecretsFromService(svc *configv1.UpstreamServiceConfig) {
	if svc == nil {
		return
	}
	if svc.UpstreamAuth != nil {
		StripSecretsFromAuth(svc.UpstreamAuth)
	}
	// TODO: Check for other places where secrets might be embedded (e.g. headers in tool definitions?)
}

// StripSecretsFromProfile removes sensitive information from the profile definition.
func StripSecretsFromProfile(profile *configv1.ProfileDefinition) {
	if profile == nil {
		return
	}
	for _, secret := range profile.Secrets {
		scrubSecretValue(secret)
	}
}

// StripSecretsFromCollection removes sensitive information from the service collection.
func StripSecretsFromCollection(collection *configv1.UpstreamServiceCollectionShare) {
	if collection == nil {
		return
	}
	for _, svc := range collection.Services {
		StripSecretsFromService(svc)
	}
}

// StripSecretsFromAuth removes sensitive values from the authentication config.
func StripSecretsFromAuth(auth *configv1.Authentication) {
	if auth == nil {
		return
	}

	if apiKey := auth.GetApiKey(); apiKey != nil {
		scrubSecretValue(apiKey.Value)
	}
	if bearer := auth.GetBearerToken(); bearer != nil {
		scrubSecretValue(bearer.Token)
	}
	if basic := auth.GetBasicAuth(); basic != nil {
		scrubSecretValue(basic.Password)
		// Username is usually safe
	}
	if oauth := auth.GetOauth2(); oauth != nil {
		scrubSecretValue(oauth.ClientSecret)
		// ClientID is usually safe-ish, but maybe scrub if paranoid?
		// ClientSecret is definitely sensitive.
	}
	// Add other auth types as needed
}

func scrubSecretValue(sv *configv1.SecretValue) {
	if sv == nil {
		return
	}
	// If it is a PLAIN value, we must remove it.
	if _, ok := sv.Value.(*configv1.SecretValue_PlainText); ok {
		sv.Value = nil
	}
}

// HydrateSecretsInService populates the service configuration with resolved secret values.
func HydrateSecretsInService(svc *configv1.UpstreamServiceConfig, secrets map[string]*configv1.SecretValue) {
	if svc == nil || len(secrets) == 0 {
		return
	}

	if auth := svc.UpstreamAuth; auth != nil {
		hydrateSecretsInAuth(auth, secrets)
	}

	// Hydrate other places if needed (e.g. Env vars in command line service)
	// TODO: Add hydration for container environments and other fields.
}

func hydrateSecretsInAuth(auth *configv1.Authentication, secrets map[string]*configv1.SecretValue) {
	if apiKey := auth.GetApiKey(); apiKey != nil {
		hydrateSecretValue(apiKey.Value, secrets)
	}
	if bearer := auth.GetBearerToken(); bearer != nil {
		hydrateSecretValue(bearer.Token, secrets)
	}
	if basic := auth.GetBasicAuth(); basic != nil {
		hydrateSecretValue(basic.Password, secrets)
	}
	if oauth := auth.GetOauth2(); oauth != nil {
		hydrateSecretValue(oauth.ClientId, secrets)
		hydrateSecretValue(oauth.ClientSecret, secrets)
	}
}

func hydrateSecretValue(sv *configv1.SecretValue, secrets map[string]*configv1.SecretValue) {
	if sv == nil {
		return
	}
	// Check if it's an environment variable reference
	if envVar, ok := sv.Value.(*configv1.SecretValue_EnvironmentVariable); ok {
		key := envVar.EnvironmentVariable
		if secret, exists := secrets[key]; exists {
			// Replace with the secret from profile
			// We clone it to avoid shared state issues if we mutate later (though we shouldn't)
			sv.Value = proto.Clone(secret).(*configv1.SecretValue).Value
		}
	}
}
