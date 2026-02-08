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
				svc := configv1.HttpUpstreamService_builder{
					Address: proto.String(badURL),
				}.Build()
				return validateHTTPService(svc)
			},
			errSubstr: "http address contains hidden whitespace",
		},
		{
			name: "WebSocket Service Address",
			validate: func() error {
				svc := configv1.WebsocketUpstreamService_builder{
					Address: proto.String(badURL),
				}.Build()
				return validateWebSocketService(svc)
			},
			errSubstr: "websocket address contains hidden whitespace",
		},
		{
			name: "GraphQL Service Address",
			validate: func() error {
				svc := configv1.GraphQLUpstreamService_builder{
					Address: proto.String(badURL),
				}.Build()
				return validateGraphQLService(svc)
			},
			errSubstr: "graphql address contains hidden whitespace",
		},
		{
			name: "WebRTC Service Address",
			validate: func() error {
				svc := configv1.WebrtcUpstreamService_builder{
					Address: proto.String(badURL),
				}.Build()
				return validateWebrtcService(svc)
			},
			errSubstr: "webrtc address contains hidden whitespace",
		},
		{
			name: "MCP Service HTTP Address",
			validate: func() error {
				svc := configv1.McpUpstreamService_builder{
					HttpConnection: configv1.McpStreamableHttpConnection_builder{
						HttpAddress: proto.String(badURL),
					}.Build(),
				}.Build()
				return validateMcpService(ctx, svc)
			},
			errSubstr: "mcp http_address contains hidden whitespace",
		},
		{
			name: "OpenAPI Address",
			validate: func() error {
				svc := configv1.OpenapiUpstreamService_builder{
					Address: proto.String(badURL),
				}.Build()
				return validateOpenAPIService(svc)
			},
			errSubstr: "openapi address contains hidden whitespace",
		},
		{
			name: "OpenAPI SpecURL",
			validate: func() error {
				svc := configv1.OpenapiUpstreamService_builder{
					SpecUrl: proto.String(badURL),
				}.Build()
				return validateOpenAPIService(svc)
			},
			errSubstr: "openapi spec_url contains hidden whitespace",
		},
		{
			name: "Audit Webhook URL",
			validate: func() error {
				cfg := configv1.AuditConfig_builder{
					Enabled:     proto.Bool(true),
					StorageType: configv1.AuditConfig_STORAGE_TYPE_WEBHOOK.Enum(),
					WebhookUrl:  proto.String(badURL),
				}.Build()
				return validateAuditConfig(cfg)
			},
			errSubstr: "webhook_url contains hidden whitespace",
		},
		{
			name: "OAuth2 Issuer URL",
			validate: func() error {
				auth := configv1.OAuth2Auth_builder{
					IssuerUrl: proto.String(badURL),
					ClientId: configv1.SecretValue_builder{
						PlainText: proto.String("id"),
					}.Build(),
					ClientSecret: configv1.SecretValue_builder{
						PlainText: proto.String("secret"),
					}.Build(),
				}.Build()
				return validateOAuth2Auth(ctx, auth)
			},
			errSubstr: "oauth2 issuer_url contains hidden whitespace",
		},
		{
			name: "OAuth2 Token URL",
			validate: func() error {
				auth := configv1.OAuth2Auth_builder{
					TokenUrl: proto.String(badURL),
					ClientId: configv1.SecretValue_builder{
						PlainText: proto.String("id"),
					}.Build(),
					ClientSecret: configv1.SecretValue_builder{
						PlainText: proto.String("secret"),
					}.Build(),
				}.Build()
				return validateOAuth2Auth(ctx, auth)
			},
			errSubstr: "oauth2 token_url contains hidden whitespace",
		},
		{
			name: "OIDC Issuer URL",
			validate: func() error {
				auth := configv1.OIDCAuth_builder{
					Issuer: proto.String(badURL),
				}.Build()
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
