// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"context"
	"strings"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
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
				return validateHTTPService(&configv1.HttpUpstreamService{Address: &badURL})
			},
			errSubstr: "http address contains hidden whitespace",
		},
		{
			name: "WebSocket Service Address",
			validate: func() error {
				return validateWebSocketService(&configv1.WebsocketUpstreamService{Address: &badURL})
			},
			errSubstr: "websocket address contains hidden whitespace",
		},
		{
			name: "GraphQL Service Address",
			validate: func() error {
				return validateGraphQLService(&configv1.GraphQLUpstreamService{Address: &badURL})
			},
			errSubstr: "graphql address contains hidden whitespace",
		},
		{
			name: "WebRTC Service Address",
			validate: func() error {
				return validateWebrtcService(&configv1.WebrtcUpstreamService{Address: &badURL})
			},
			errSubstr: "webrtc address contains hidden whitespace",
		},
		{
			name: "MCP Service HTTP Address",
			validate: func() error {
				return validateMcpService(&configv1.McpUpstreamService{
					ConnectionType: &configv1.McpUpstreamService_HttpConnection{
						HttpConnection: &configv1.McpStreamableHttpConnection{HttpAddress: &badURL},
					},
				})
			},
			errSubstr: "mcp http_address contains hidden whitespace",
		},
		{
			name: "OpenAPI Address",
			validate: func() error {
				return validateOpenAPIService(&configv1.OpenapiUpstreamService{Address: &badURL})
			},
			errSubstr: "openapi address contains hidden whitespace",
		},
		{
			name: "OpenAPI SpecURL",
			validate: func() error {
				return validateOpenAPIService(&configv1.OpenapiUpstreamService{
					SpecSource: &configv1.OpenapiUpstreamService_SpecUrl{SpecUrl: badURL},
				})
			},
			errSubstr: "openapi spec_url contains hidden whitespace",
		},
		{
			name: "Audit Webhook URL",
			validate: func() error {
				return validateAuditConfig(&configv1.AuditConfig{
					Enabled:     proto.Bool(true),
					StorageType: configv1.AuditConfig_STORAGE_TYPE_WEBHOOK.Enum(),
					WebhookUrl:  &badURL,
				})
			},
			errSubstr: "webhook_url contains hidden whitespace",
		},
		{
			name: "OAuth2 Issuer URL",
			validate: func() error {
				return validateOAuth2Auth(ctx, &configv1.OAuth2Auth{
					IssuerUrl:    &badURL,
					ClientId:     &configv1.SecretValue{Value: &configv1.SecretValue_PlainText{PlainText: "id"}},
					ClientSecret: &configv1.SecretValue{Value: &configv1.SecretValue_PlainText{PlainText: "secret"}},
				})
			},
			errSubstr: "oauth2 issuer_url contains hidden whitespace",
		},
		{
			name: "OAuth2 Token URL",
			validate: func() error {
				return validateOAuth2Auth(ctx, &configv1.OAuth2Auth{
					TokenUrl:     &badURL,
					ClientId:     &configv1.SecretValue{Value: &configv1.SecretValue_PlainText{PlainText: "id"}},
					ClientSecret: &configv1.SecretValue{Value: &configv1.SecretValue_PlainText{PlainText: "secret"}},
				})
			},
			errSubstr: "oauth2 token_url contains hidden whitespace",
		},
		{
			name: "OIDC Issuer URL",
			validate: func() error {
				return validateOIDCAuth(ctx, &configv1.OIDCAuth{Issuer: &badURL})
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
