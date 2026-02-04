// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package util //nolint:revive,nolintlint // Package name 'util' is common in this codebase

import (
	configv1 "github.com/mcpany/core/proto/config/v1"
)

// StripSecretsFromService removes sensitive information from the service configuration.
//
// Summary: Removes sensitive information from the service configuration.
//
// It specifically targets plain text secrets in UpstreamAuth and other locations.
//
// Parameters:
//   - svc: The upstream service configuration to strip secrets from.
func StripSecretsFromService(svc *configv1.UpstreamServiceConfig) {
	if svc == nil {
		return
	}
	if svc.GetUpstreamAuth() != nil {
		StripSecretsFromAuth(svc.GetUpstreamAuth())
	}
	if svc.GetAuthentication() != nil {
		StripSecretsFromAuth(svc.GetAuthentication())
	}

	// Service specific config
	// Service specific config
	if s := svc.GetCommandLineService(); s != nil {
		stripSecretsFromCommandLineService(s)
	} else if s := svc.GetHttpService(); s != nil {
		stripSecretsFromHTTPService(s)
	} else if s := svc.GetMcpService(); s != nil {
		stripSecretsFromMcpService(s)
	} else if s := svc.GetFilesystemService(); s != nil {
		stripSecretsFromFilesystemService(s)
	} else if s := svc.GetVectorService(); s != nil {
		stripSecretsFromVectorService(s)
	} else if s := svc.GetWebsocketService(); s != nil {
		stripSecretsFromWebsocketService(s)
	} else if s := svc.GetWebrtcService(); s != nil {
		stripSecretsFromWebrtcService(s)
	} else if s := svc.GetGrpcService(); s != nil {
		stripSecretsFromGrpcService(s)
	} else if s := svc.GetOpenapiService(); s != nil {
		stripSecretsFromOpenapiService(s)
	}
	// GraphQL and SQL services have no secrets to strip currently

	// Hooks
	for _, hook := range svc.GetPreCallHooks() {
		stripSecretsFromHook(hook)
	}
	for _, hook := range svc.GetPostCallHooks() {
		stripSecretsFromHook(hook)
	}

	// Cache
	if svc.GetCache() != nil {
		stripSecretsFromCacheConfig(svc.GetCache())
	}
}

// StripSecretsFromProfile removes sensitive information from the profile definition.
//
// Summary: Removes sensitive information from the profile definition.
//
// Parameters:
//   - profile: The profile definition to strip secrets from.
func StripSecretsFromProfile(profile *configv1.ProfileDefinition) {
	if profile == nil {
		return
	}
	for _, secret := range profile.GetSecrets() {
		scrubSecretValue(secret)
	}
}

// StripSecretsFromCollection removes sensitive information from the service collection.
//
// Summary: Removes sensitive information from the service collection.
//
// Parameters:
//   - collection: The service collection to strip secrets from.
func StripSecretsFromCollection(collection *configv1.Collection) {
	if collection == nil {
		return
	}
	for _, svc := range collection.GetServices() {
		StripSecretsFromService(svc)
	}
}

// StripSecretsFromAuth removes sensitive values from the authentication config.
//
// Summary: Removes sensitive values from the authentication config.
//
// Parameters:
//   - auth: The authentication configuration to strip secrets from.
func StripSecretsFromAuth(auth *configv1.Authentication) {
	if auth == nil {
		return
	}

	if apiKey := auth.GetApiKey(); apiKey != nil {
		scrubSecretValue(apiKey.GetValue())
		if apiKey.GetVerificationValue() != "" {
			// verification_value is just a string field in Opaque object (not unexported?).
			// Wait, verify if we can set fields directly on Opaque object?
			// NO. Opaque objects have unexported fields.
			// We cannot set apiKey.VerificationValue = proto.String("")
			// We MUST use a Setter or Builder.
			// DOES THE OPAQUE API HAVE SETTERS?
			// Usually yes, SetXxx.
			apiKey.SetVerificationValue("")
		}
	}
	if bearer := auth.GetBearerToken(); bearer != nil {
		scrubSecretValue(bearer.GetToken())
	}
	if basic := auth.GetBasicAuth(); basic != nil {
		scrubSecretValue(basic.GetPassword())
		// Username is usually safe
	}
	if oauth := auth.GetOauth2(); oauth != nil {
		scrubSecretValue(oauth.GetClientSecret())
		scrubSecretValue(oauth.GetClientId())
	}
	// Add other auth types as needed
}

func stripSecretsFromCommandLineService(s *configv1.CommandLineUpstreamService) {
	if s == nil {
		return
	}
	stripSecretsFromSecretMap(s.GetEnv())
	for _, call := range s.GetCalls() {
		stripSecretsFromCommandLineCall(call)
	}
}

func stripSecretsFromHTTPService(s *configv1.HttpUpstreamService) {
	if s == nil {
		return
	}
	for _, call := range s.GetCalls() {
		stripSecretsFromHTTPCall(call)
	}
}

func stripSecretsFromMcpService(s *configv1.McpUpstreamService) {
	if s == nil {
		return
	}
	if conn := s.GetStdioConnection(); conn != nil {
		stripSecretsFromSecretMap(conn.GetEnv())
	} else if conn := s.GetBundleConnection(); conn != nil {
		stripSecretsFromSecretMap(conn.GetEnv())
	}
	for _, call := range s.GetCalls() {
		stripSecretsFromMcpCall(call)
	}
}

func stripSecretsFromFilesystemService(s *configv1.FilesystemUpstreamService) {
	if s == nil {
		return
	}
	if fs := s.GetS3(); fs != nil { // GetS3() ? The oneof is filesystem_type.
		// Need to check names of getters for FilesystemService oneof.
		// Assuming named based on field names: s3, sftp, etc.
		if fs.GetSecretAccessKey() != "" {
			fs.SetSecretAccessKey("")
		}
		if fs.GetSessionToken() != "" {
			fs.SetSessionToken("")
		}
	} else if fs := s.GetSftp(); fs != nil {
		if fs.GetPassword() != "" {
			fs.SetPassword("")
		}
	}
}

func stripSecretsFromVectorService(s *configv1.VectorUpstreamService) {
	if s == nil {
		return
	}
	// Oneof vector_db_type
	if db := s.GetPinecone(); db != nil {
		if db.GetApiKey() != "" {
			db.SetApiKey("")
		}
	} else if db := s.GetMilvus(); db != nil {
		if db.GetApiKey() != "" {
			db.SetApiKey("")
		}
		if db.GetPassword() != "" {
			db.SetPassword("")
		}
	}
}

func stripSecretsFromWebsocketService(s *configv1.WebsocketUpstreamService) {
	if s == nil {
		return
	}
	for _, call := range s.GetCalls() {
		stripSecretsFromWebsocketCall(call)
	}
}

func stripSecretsFromWebrtcService(s *configv1.WebrtcUpstreamService) {
	if s == nil {
		return
	}
	for _, call := range s.GetCalls() {
		stripSecretsFromWebrtcCall(call)
	}
}

func stripSecretsFromGrpcService(_ *configv1.GrpcUpstreamService) {
	// gRPC calls don't have explicit parameter mapping with secrets currently defined in proto.
	// If they do, add logic here.
}

func stripSecretsFromOpenapiService(_ *configv1.OpenapiUpstreamService) {
	// OpenAPI calls use generic structures, check if they have secret mappings.
	// Current definition OpenAPICallDefinition doesn't have parameter mappings like HTTP.
}

func stripSecretsFromHook(h *configv1.CallHook) {
	if h == nil {
		return
	}
	if wh := h.GetWebhook(); wh != nil {
		// WebhookSecret is a string, clear it.
		wh.SetWebhookSecret("")
	}
}

func stripSecretsFromCacheConfig(c *configv1.CacheConfig) {
	// Deprecated ApiKey
	scrubSecretValue(c.GetSemanticConfig().GetApiKey())

	// Provider specific configs
	if openai := c.GetSemanticConfig().GetOpenai(); openai != nil {
		scrubSecretValue(openai.GetApiKey())
	}
	// Add other providers if they have secrets
}

func stripSecretsFromCommandLineCall(c *configv1.CommandLineCallDefinition) {
	if c == nil {
		return
	}
	for _, param := range c.GetParameters() {
		scrubSecretValue(param.GetSecret())
	}
}

func stripSecretsFromHTTPCall(c *configv1.HttpCallDefinition) {
	if c == nil {
		return
	}
	for _, param := range c.GetParameters() {
		scrubSecretValue(param.GetSecret())
	}
}

func stripSecretsFromWebsocketCall(c *configv1.WebsocketCallDefinition) {
	if c == nil {
		return
	}
	for _, param := range c.GetParameters() {
		scrubSecretValue(param.GetSecret())
	}
}

func stripSecretsFromWebrtcCall(c *configv1.WebrtcCallDefinition) {
	if c == nil {
		return
	}
	for _, param := range c.GetParameters() {
		scrubSecretValue(param.GetSecret())
	}
}

func stripSecretsFromMcpCall(_ *configv1.MCPCallDefinition) {
	// MCPCallDefinition doesn't seem to have explicit parameter mappings with secrets in the proto definition I read.
	// It uses input_schema and transformers.
}

func stripSecretsFromSecretMap(m map[string]*configv1.SecretValue) {
	for _, sv := range m {
		scrubSecretValue(sv)
	}
}

func scrubSecretValue(sv *configv1.SecretValue) {
	if sv == nil {
		return
	}
	// If it is a PLAIN value, we must remove it.
	// Opaque API: Value is a oneof.
	if sv.HasPlainText() {
		sv.ClearValue() // Scrub it.
	}
}

// HydrateSecretsInService populates the service configuration with resolved secret values.
//
// Summary: Populates the service configuration with resolved secret values.
//
// Parameters:
//   - svc: The upstream service configuration to hydrate secrets into.
//   - secrets: A map of resolved secret values.
func HydrateSecretsInService(svc *configv1.UpstreamServiceConfig, secrets map[string]*configv1.SecretValue) {
	if svc == nil || len(secrets) == 0 {
		return
	}

	if auth := svc.GetUpstreamAuth(); auth != nil {
		hydrateSecretsInAuth(auth, secrets)
	}

	// Hydrate other places if needed (e.g. Env vars in command line service)
	// Hydrate other places if needed (e.g. Env vars in command line service)
	if s := svc.GetCommandLineService(); s != nil {
		hydrateSecretsInEnv(s.GetEnv(), secrets)
		if ce := s.GetContainerEnvironment(); ce != nil {
			hydrateSecretsInEnv(ce.GetEnv(), secrets)
		}
	} else if s := svc.GetMcpService(); s != nil {
		if conn := s.GetStdioConnection(); conn != nil {
			hydrateSecretsInEnv(conn.GetEnv(), secrets)
		} else if conn := s.GetBundleConnection(); conn != nil {
			hydrateSecretsInEnv(conn.GetEnv(), secrets)
		}
	} else if s := svc.GetHttpService(); s != nil {
		hydrateSecretsInHTTPService(s, secrets)
	} else if s := svc.GetWebsocketService(); s != nil {
		hydrateSecretsInWebsocketService(s, secrets)
	} else if s := svc.GetWebrtcService(); s != nil {
		hydrateSecretsInWebrtcService(s, secrets)
	}
}

func hydrateSecretsInHTTPService(s *configv1.HttpUpstreamService, secrets map[string]*configv1.SecretValue) {
	if s == nil {
		return
	}
	for _, call := range s.GetCalls() {
		if call == nil {
			continue
		}
		for _, param := range call.GetParameters() {
			hydrateSecretValue(param.GetSecret(), secrets)
		}
	}
}

func hydrateSecretsInWebsocketService(s *configv1.WebsocketUpstreamService, secrets map[string]*configv1.SecretValue) {
	if s == nil {
		return
	}
	for _, call := range s.GetCalls() {
		if call == nil {
			continue
		}
		for _, param := range call.GetParameters() {
			hydrateSecretValue(param.GetSecret(), secrets)
		}
	}
}

func hydrateSecretsInWebrtcService(s *configv1.WebrtcUpstreamService, secrets map[string]*configv1.SecretValue) {
	if s == nil {
		return
	}
	for _, call := range s.GetCalls() {
		if call == nil {
			continue
		}
		for _, param := range call.GetParameters() {
			hydrateSecretValue(param.GetSecret(), secrets)
		}
	}
}

func hydrateSecretsInEnv(env map[string]*configv1.SecretValue, secrets map[string]*configv1.SecretValue) {
	for _, val := range env {
		hydrateSecretValue(val, secrets)
	}
}

func hydrateSecretsInAuth(auth *configv1.Authentication, secrets map[string]*configv1.SecretValue) {
	if apiKey := auth.GetApiKey(); apiKey != nil {
		hydrateSecretValue(apiKey.GetValue(), secrets)
	}
	if bearer := auth.GetBearerToken(); bearer != nil {
		hydrateSecretValue(bearer.GetToken(), secrets)
	}
	if basic := auth.GetBasicAuth(); basic != nil {
		hydrateSecretValue(basic.GetPassword(), secrets)
	}
	if oauth := auth.GetOauth2(); oauth != nil {
		hydrateSecretValue(oauth.GetClientId(), secrets)
		hydrateSecretValue(oauth.GetClientSecret(), secrets)
	}
}

func hydrateSecretValue(sv *configv1.SecretValue, secrets map[string]*configv1.SecretValue) {
	if sv == nil {
		return
	}
	// Check if it's an environment variable reference
	// Opaque: Check GetEnvironmentVariable().
	if key := sv.GetEnvironmentVariable(); key != "" {
		if secret, exists := secrets[key]; exists {
			// Replace with the secret from profile
			// We need to set the value.
			// secret.GetValue() is oneof.
			// Opaque API does not allow direct assignment of oneof field.
			// We MUST check what type secret has, and set that on sv.
			// Or if we can clone the whole SecretValue and assume wrapper compatibility?
			// sv is *SecretValue.
			// proto.Merge(sv, secret) ?
			// But we only want to copy the Value oneof.

			// If secret has PlainText, set PlainText.
			if secret.HasPlainText() {
				sv.SetPlainText(secret.GetPlainText())
			}
			// If other types are supported in SecretValue (usually just PlainText or EnvVar).
		}
	}
}
