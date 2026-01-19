// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

// Package config provides configuration management for MCP Any.
package config

import (
	configv1 "github.com/mcpany/core/proto/config/v1"
	"google.golang.org/protobuf/proto"
)

// StripSecretsFromService removes sensitive information from the service configuration.
// It specifically targets plain text secrets in UpstreamAuth and other locations.
func StripSecretsFromService(svc *configv1.UpstreamServiceConfig) {
	if svc == nil {
		return
	}
	if svc.UpstreamAuth != nil {
		StripSecretsFromAuth(svc.UpstreamAuth)
	}
	if svc.Authentication != nil {
		StripSecretsFromAuth(svc.Authentication)
	}

	// Service specific config
	switch s := svc.ServiceConfig.(type) {
	case *configv1.UpstreamServiceConfig_CommandLineService:
		stripSecretsFromCommandLineService(s.CommandLineService)
	case *configv1.UpstreamServiceConfig_HttpService:
		stripSecretsFromHTTPService(s.HttpService)
	case *configv1.UpstreamServiceConfig_McpService:
		stripSecretsFromMcpService(s.McpService)
	case *configv1.UpstreamServiceConfig_FilesystemService:
		stripSecretsFromFilesystemService(s.FilesystemService)
	case *configv1.UpstreamServiceConfig_VectorService:
		stripSecretsFromVectorService(s.VectorService)
	case *configv1.UpstreamServiceConfig_WebsocketService:
		stripSecretsFromWebsocketService(s.WebsocketService)
	case *configv1.UpstreamServiceConfig_WebrtcService:
		stripSecretsFromWebrtcService(s.WebrtcService)
	case *configv1.UpstreamServiceConfig_GrpcService:
		stripSecretsFromGrpcService(s.GrpcService)
	case *configv1.UpstreamServiceConfig_OpenapiService:
		stripSecretsFromOpenapiService(s.OpenapiService)
	case *configv1.UpstreamServiceConfig_GraphqlService:
		// No explicit secrets in GraphQL service definition yet, but checking calls might be good if added later.
	case *configv1.UpstreamServiceConfig_SqlService:
		// No explicit secrets in SQL service definition yet.
	}

	// Hooks
	for _, hook := range svc.PreCallHooks {
		stripSecretsFromHook(hook)
	}
	for _, hook := range svc.PostCallHooks {
		stripSecretsFromHook(hook)
	}

	// Cache
	if svc.Cache != nil {
		stripSecretsFromCacheConfig(svc.Cache)
	}
}

// StripSecretsFromProfile removes sensitive information from the profile definition.
//
// profile is the profile.
func StripSecretsFromProfile(profile *configv1.ProfileDefinition) {
	if profile == nil {
		return
	}
	for _, secret := range profile.Secrets {
		scrubSecretValue(secret)
	}
}

// StripSecretsFromCollection removes sensitive information from the service collection.
//
// collection is the collection.
func StripSecretsFromCollection(collection *configv1.Collection) {
	if collection == nil {
		return
	}
	for _, svc := range collection.Services {
		StripSecretsFromService(svc)
	}
}

// StripSecretsFromAuth removes sensitive values from the authentication config.
//
// auth is the auth.
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
		scrubSecretValue(oauth.ClientId)
	}
	// Add other auth types as needed
}

func stripSecretsFromCommandLineService(s *configv1.CommandLineUpstreamService) {
	if s == nil {
		return
	}
	stripSecretsFromSecretMap(s.Env)
	for _, call := range s.Calls {
		stripSecretsFromCommandLineCall(call)
	}
}

func stripSecretsFromHTTPService(s *configv1.HttpUpstreamService) {
	if s == nil {
		return
	}
	for _, call := range s.Calls {
		stripSecretsFromHTTPCall(call)
	}
}

func stripSecretsFromMcpService(s *configv1.McpUpstreamService) {
	if s == nil {
		return
	}
	switch conn := s.ConnectionType.(type) {
	case *configv1.McpUpstreamService_StdioConnection:
		stripSecretsFromSecretMap(conn.StdioConnection.Env)
	case *configv1.McpUpstreamService_BundleConnection:
		stripSecretsFromSecretMap(conn.BundleConnection.Env)
	}
	for _, call := range s.Calls {
		stripSecretsFromMcpCall(call)
	}
}

func stripSecretsFromFilesystemService(s *configv1.FilesystemUpstreamService) {
	if s == nil {
		return
	}
	switch fs := s.FilesystemType.(type) {
	case *configv1.FilesystemUpstreamService_S3:
		if fs.S3.SecretAccessKey != nil {
			fs.S3.SecretAccessKey = proto.String("")
		}
		if fs.S3.SessionToken != nil {
			fs.S3.SessionToken = proto.String("")
		}
	case *configv1.FilesystemUpstreamService_Sftp:
		if fs.Sftp.Password != nil {
			fs.Sftp.Password = proto.String("")
		}
	}
}

func stripSecretsFromVectorService(s *configv1.VectorUpstreamService) {
	if s == nil {
		return
	}
	switch db := s.VectorDbType.(type) {
	case *configv1.VectorUpstreamService_Pinecone:
		if db.Pinecone.ApiKey != nil {
			db.Pinecone.ApiKey = proto.String("")
		}
	case *configv1.VectorUpstreamService_Milvus:
		if db.Milvus.ApiKey != nil {
			db.Milvus.ApiKey = proto.String("")
		}
		if db.Milvus.Password != nil {
			db.Milvus.Password = proto.String("")
		}
	}
}

func stripSecretsFromWebsocketService(s *configv1.WebsocketUpstreamService) {
	if s == nil {
		return
	}
	for _, call := range s.Calls {
		stripSecretsFromWebsocketCall(call)
	}
}

func stripSecretsFromWebrtcService(s *configv1.WebrtcUpstreamService) {
	if s == nil {
		return
	}
	for _, call := range s.Calls {
		stripSecretsFromWebrtcCall(call)
	}
}

func stripSecretsFromGrpcService(_ *configv1.GrpcUpstreamService) {
	// gRPC calls don't have explicit parameter mapping with secrets currently defined in proto.
	// If they do, add logic here.
	_ = "placeholder"
}

func stripSecretsFromOpenapiService(_ *configv1.OpenapiUpstreamService) {
	// OpenAPI calls use generic structures, check if they have secret mappings.
	// Current definition OpenAPICallDefinition doesn't have parameter mappings like HTTP.
	_ = "placeholder"
}

func stripSecretsFromHook(h *configv1.CallHook) {
	if h == nil {
		return
	}
	if wh := h.GetWebhook(); wh != nil {
		// WebhookSecret is a string, clear it.
		wh.WebhookSecret = ""
	}
}

func stripSecretsFromCacheConfig(c *configv1.CacheConfig) {
	if c == nil || c.SemanticConfig == nil {
		return
	}
	// Deprecated ApiKey
	scrubSecretValue(c.SemanticConfig.ApiKey)

	// Provider specific configs
	if openai := c.SemanticConfig.GetOpenai(); openai != nil {
		scrubSecretValue(openai.ApiKey)
	}
	// Add other providers if they have secrets
}

func stripSecretsFromCommandLineCall(c *configv1.CommandLineCallDefinition) {
	if c == nil {
		return
	}
	for _, param := range c.Parameters {
		scrubSecretValue(param.Secret)
	}
}

func stripSecretsFromHTTPCall(c *configv1.HttpCallDefinition) {
	if c == nil {
		return
	}
	for _, param := range c.Parameters {
		scrubSecretValue(param.Secret)
	}
}

func stripSecretsFromWebsocketCall(c *configv1.WebsocketCallDefinition) {
	if c == nil {
		return
	}
	for _, param := range c.Parameters {
		scrubSecretValue(param.Secret)
	}
}

func stripSecretsFromWebrtcCall(c *configv1.WebrtcCallDefinition) {
	if c == nil {
		return
	}
	for _, param := range c.Parameters {
		scrubSecretValue(param.Secret)
	}
}

func stripSecretsFromMcpCall(_ *configv1.MCPCallDefinition) {
	// MCPCallDefinition doesn't seem to have explicit parameter mappings with secrets in the proto definition I read.
	// It uses input_schema and transformers.
	_ = "placeholder"
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
	if _, ok := sv.Value.(*configv1.SecretValue_PlainText); ok {
		sv.Value = nil
	}
}

// HydrateSecretsInService populates the service configuration with resolved secret values.
//
// svc is the svc.
// secrets is the secrets.
func HydrateSecretsInService(svc *configv1.UpstreamServiceConfig, secrets map[string]*configv1.SecretValue) {
	if svc == nil || len(secrets) == 0 {
		return
	}

	if auth := svc.UpstreamAuth; auth != nil {
		hydrateSecretsInAuth(auth, secrets)
	}

	// Hydrate other places if needed (e.g. Env vars in command line service)
	switch s := svc.ServiceConfig.(type) {
	case *configv1.UpstreamServiceConfig_CommandLineService:
		if cmd := s.CommandLineService; cmd != nil {
			hydrateSecretsInEnv(cmd.Env, secrets)
			if ce := cmd.ContainerEnvironment; ce != nil {
				hydrateSecretsInEnv(ce.Env, secrets)
			}
		}
	case *configv1.UpstreamServiceConfig_McpService:
		if mcp := s.McpService; mcp != nil {
			switch conn := mcp.ConnectionType.(type) {
			case *configv1.McpUpstreamService_StdioConnection:
				if stdio := conn.StdioConnection; stdio != nil {
					hydrateSecretsInEnv(stdio.Env, secrets)
				}
			case *configv1.McpUpstreamService_BundleConnection:
				if bundle := conn.BundleConnection; bundle != nil {
					hydrateSecretsInEnv(bundle.Env, secrets)
				}
			}
		}
	case *configv1.UpstreamServiceConfig_HttpService:
		hydrateSecretsInHTTPService(s.HttpService, secrets)
	case *configv1.UpstreamServiceConfig_WebsocketService:
		hydrateSecretsInWebsocketService(s.WebsocketService, secrets)
	case *configv1.UpstreamServiceConfig_WebrtcService:
		hydrateSecretsInWebrtcService(s.WebrtcService, secrets)
	}
}

func hydrateSecretsInHTTPService(s *configv1.HttpUpstreamService, secrets map[string]*configv1.SecretValue) {
	if s == nil {
		return
	}
	for _, call := range s.Calls {
		if call == nil {
			continue
		}
		for _, param := range call.Parameters {
			hydrateSecretValue(param.Secret, secrets)
		}
	}
}

func hydrateSecretsInWebsocketService(s *configv1.WebsocketUpstreamService, secrets map[string]*configv1.SecretValue) {
	if s == nil {
		return
	}
	for _, call := range s.Calls {
		if call == nil {
			continue
		}
		for _, param := range call.Parameters {
			hydrateSecretValue(param.Secret, secrets)
		}
	}
}

func hydrateSecretsInWebrtcService(s *configv1.WebrtcUpstreamService, secrets map[string]*configv1.SecretValue) {
	if s == nil {
		return
	}
	for _, call := range s.Calls {
		if call == nil {
			continue
		}
		for _, param := range call.Parameters {
			hydrateSecretValue(param.Secret, secrets)
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
