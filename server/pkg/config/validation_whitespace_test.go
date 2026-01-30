// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"context"
	"strings"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
)

// TestValidationWhitespaceChecks verifies that the validator detects trailing/leading whitespace
// in various URL configuration fields.
func TestValidationWhitespaceChecks(t *testing.T) {
	badURL := "http://example.com "
	ctx := context.Background()

	tests := []struct {
		name      string
		validate  func() error
		errSubstr string
	}{
		{
			name: "HTTP Service Address",
			validate: func() error {
				svc := &configv1.HttpUpstreamService{}
				svc.SetAddress(badURL)
				return validateHTTPService(svc)
			},
			errSubstr: "http address contains hidden whitespace",
		},
		{
			name: "WebSocket Service Address",
			validate: func() error {
				svc := &configv1.WebsocketUpstreamService{}
				svc.SetAddress(badURL)
				return validateWebSocketService(svc)
			},
			errSubstr: "websocket address contains hidden whitespace",
		},
		{
			name: "GraphQL Service Address",
			validate: func() error {
				svc := &configv1.GraphQLUpstreamService{}
				svc.SetAddress(badURL)
				return validateGraphQLService(svc)
			},
			errSubstr: "graphql address contains hidden whitespace",
		},
		{
			name: "WebRTC Service Address",
			validate: func() error {
				svc := &configv1.WebrtcUpstreamService{}
				svc.SetAddress(badURL)
				return validateWebrtcService(svc)
			},
			errSubstr: "webrtc address contains hidden whitespace",
		},
		{
			name: "MCP Service HTTP Address",
			validate: func() error {
				svc := &configv1.McpUpstreamService{}
				httpConn := &configv1.McpStreamableHttpConnection{}
				httpConn.SetHttpAddress(badURL)
				svc.SetHttpConnection(httpConn)
				return validateMcpService(ctx, svc)
			},
			errSubstr: "mcp http_address contains hidden whitespace",
		},
		{
			name: "OpenAPI Address",
			validate: func() error {
				svc := &configv1.OpenapiUpstreamService{}
				svc.SetAddress(badURL)
				return validateOpenAPIService(svc)
			},
			errSubstr: "openapi address contains hidden whitespace",
		},
		{
			name: "OpenAPI SpecURL",
			validate: func() error {
				svc := &configv1.OpenapiUpstreamService{}
				svc.SetSpecUrl(badURL)
				return validateOpenAPIService(svc)
			},
			errSubstr: "openapi spec_url contains hidden whitespace",
		},
		{
			name: "Audit Webhook URL",
			validate: func() error {
				cfg := &configv1.AuditConfig{}
				cfg.SetEnabled(true)
				cfg.SetStorageType(configv1.AuditConfig_STORAGE_TYPE_WEBHOOK)
				cfg.SetWebhookUrl(badURL)
				return validateAuditConfig(cfg)
			},
			errSubstr: "webhook_url contains hidden whitespace",
		},
		{
			name: "OAuth2 Issuer URL",
			validate: func() error {
				auth := &configv1.OAuth2Auth{}
				auth.SetIssuerUrl(badURL)
				id := &configv1.SecretValue{}
				id.SetPlainText("id")
				auth.SetClientId(id)
				secret := &configv1.SecretValue{}
				secret.SetPlainText("secret")
				auth.SetClientSecret(secret)
				return validateOAuth2Auth(ctx, auth)
			},
			errSubstr: "oauth2 issuer_url contains hidden whitespace",
		},
		{
			name: "OAuth2 Token URL",
			validate: func() error {
				auth := &configv1.OAuth2Auth{}
				auth.SetTokenUrl(badURL)
				id := &configv1.SecretValue{}
				id.SetPlainText("id")
				auth.SetClientId(id)
				secret := &configv1.SecretValue{}
				secret.SetPlainText("secret")
				auth.SetClientSecret(secret)
				return validateOAuth2Auth(ctx, auth)
			},
			errSubstr: "oauth2 token_url contains hidden whitespace",
		},
		{
			name: "OIDC Issuer URL",
			validate: func() error {
				auth := &configv1.OIDCAuth{}
				auth.SetIssuer(badURL)
				return validateOIDCAuth(ctx, auth)
			},
			errSubstr: "oidc issuer url contains hidden whitespace",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.validate()
			assert.Error(t, err)
			if err != nil {
				assert.True(t, strings.Contains(err.Error(), tt.errSubstr), "Expected error to contain %q, got %v", tt.errSubstr, err)
				if ae, ok := err.(*ActionableError); ok {
					assert.Contains(t, ae.Suggestion, "remove any trailing spaces", "Suggestion should advise removing spaces")
				}
			}
		})
	}
}
