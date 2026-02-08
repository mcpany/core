package config

import (
	"context"
	"os"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

func TestValidateOIDCAuth(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name      string
		oidc      *configv1.OIDCAuth
		expectErr string
	}{
		{
			name: "valid_oidc",
			oidc: configv1.OIDCAuth_builder{
				Issuer: proto.String("https://accounts.google.com"),
			}.Build(),
			expectErr: "",
		},
		{
			name: "empty_issuer",
			oidc: configv1.OIDCAuth_builder{
				Issuer: proto.String(""),
			}.Build(),
			expectErr: "oidc issuer is empty",
		},
		{
			name: "invalid_issuer_url",
			oidc: configv1.OIDCAuth_builder{
				Issuer: proto.String("not-a-url"),
			}.Build(),
			expectErr: "invalid oidc issuer url: not-a-url",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateOIDCAuth(ctx, tt.oidc)
			if tt.expectErr != "" {
				assert.ErrorContains(t, err, tt.expectErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateTrustedHeaderAuth(t *testing.T) {
	tests := []struct {
		name      string
		th        *configv1.TrustedHeaderAuth
		expectErr string
	}{
		{
			name: "valid_trusted_header",
			th: configv1.TrustedHeaderAuth_builder{
				HeaderName:  proto.String("X-User-ID"),
				HeaderValue: proto.String("admin"),
			}.Build(),
			expectErr: "",
		},
		{
			name: "empty_header_name",
			th: configv1.TrustedHeaderAuth_builder{
				HeaderName:  proto.String(""),
				HeaderValue: proto.String("admin"),
			}.Build(),
			expectErr: "trusted header name is empty",
		},
		{
			name: "empty_header_value",
			th: configv1.TrustedHeaderAuth_builder{
				HeaderName:  proto.String("X-User-ID"),
				HeaderValue: proto.String(""),
			}.Build(),
			expectErr: "trusted header value is empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateTrustedHeaderAuth(tt.th)
			if tt.expectErr != "" {
				assert.ErrorContains(t, err, tt.expectErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateOAuth2Auth_Coverage(t *testing.T) {
	ctx := context.Background()
	os.Setenv("CLIENT_ID", "my-id")
	os.Setenv("CLIENT_SECRET", "my-secret")
	defer os.Unsetenv("CLIENT_ID")
	defer os.Unsetenv("CLIENT_SECRET")

	tests := []struct {
		name      string
		oauth     *configv1.OAuth2Auth
		expectErr string
	}{
		{
			name: "valid_issuer_auto_discovery",
			oauth: configv1.OAuth2Auth_builder{
				TokenUrl:  proto.String(""),
				IssuerUrl: proto.String("https://accounts.google.com"),
				ClientId: configv1.SecretValue_builder{
					EnvironmentVariable: proto.String("CLIENT_ID"),
				}.Build(),
				ClientSecret: configv1.SecretValue_builder{
					EnvironmentVariable: proto.String("CLIENT_SECRET"),
				}.Build(),
			}.Build(),
			expectErr: "",
		},
		{
			name: "invalid_issuer_url",
			oauth: configv1.OAuth2Auth_builder{
				TokenUrl:  proto.String(""), // empty token url
				IssuerUrl: proto.String("not-a-url"),
				ClientId: configv1.SecretValue_builder{
					EnvironmentVariable: proto.String("CLIENT_ID"),
				}.Build(),
				ClientSecret: configv1.SecretValue_builder{
					EnvironmentVariable: proto.String("CLIENT_SECRET"),
				}.Build(),
			}.Build(),
			expectErr: "invalid oauth2 issuer_url",
		},
		{
			name: "invalid_token_url",
			oauth: configv1.OAuth2Auth_builder{
				TokenUrl: proto.String("not-a-url"),
				ClientId: configv1.SecretValue_builder{
					EnvironmentVariable: proto.String("CLIENT_ID"),
				}.Build(),
				ClientSecret: configv1.SecretValue_builder{
					EnvironmentVariable: proto.String("CLIENT_SECRET"),
				}.Build(),
			}.Build(),
			expectErr: "invalid oauth2 token_url",
		},
		{
			name: "missing_token_and_issuer",
			oauth: configv1.OAuth2Auth_builder{
				TokenUrl:  proto.String(""),
				IssuerUrl: proto.String(""),
				ClientId: configv1.SecretValue_builder{
					EnvironmentVariable: proto.String("CLIENT_ID"),
				}.Build(),
				ClientSecret: configv1.SecretValue_builder{
					EnvironmentVariable: proto.String("CLIENT_SECRET"),
				}.Build(),
			}.Build(),
			expectErr: "oauth2 token_url is empty and no issuer_url provided",
		},
		{
			name: "missing_client_id_secret_struct",
			oauth: configv1.OAuth2Auth_builder{
				TokenUrl: proto.String("https://example.com/token"),
				ClientId: nil, // Missing struct
				ClientSecret: configv1.SecretValue_builder{
					EnvironmentVariable: proto.String("CLIENT_SECRET"),
				}.Build(),
			}.Build(),
			expectErr: "client_id", // validateSecretValue might not fail on nil? Check implementation. validateSecretValue(nil) returns nil. validateOAuth2Auth calls validateSecretValue then resolve.
		},
		{
			name: "empty_client_id_resolved",
			oauth: configv1.OAuth2Auth_builder{
				TokenUrl: proto.String("https://example.com/token"),
				ClientId: configv1.SecretValue_builder{
					PlainText: proto.String(""),
				}.Build(),
				ClientSecret: configv1.SecretValue_builder{
					EnvironmentVariable: proto.String("CLIENT_SECRET"),
				}.Build(),
			}.Build(),
			expectErr: "oauth2 client_id is missing or empty",
		},
		{
			name: "empty_client_secret_resolved",
			oauth: configv1.OAuth2Auth_builder{
				TokenUrl: proto.String("https://example.com/token"),
				ClientId: configv1.SecretValue_builder{
					EnvironmentVariable: proto.String("CLIENT_ID"),
				}.Build(),
				ClientSecret: configv1.SecretValue_builder{
					PlainText: proto.String(""),
				}.Build(),
			}.Build(),
			expectErr: "oauth2 client_secret is missing or empty",
		},
        {
			name: "invalid_client_id_secret_validation",
			oauth: configv1.OAuth2Auth_builder{
				TokenUrl: proto.String("https://example.com/token"),
				ClientId: configv1.SecretValue_builder{
					EnvironmentVariable: proto.String("MISSING_VAR"),
				}.Build(),
				ClientSecret: configv1.SecretValue_builder{
					EnvironmentVariable: proto.String("CLIENT_SECRET"),
				}.Build(),
			}.Build(),
			expectErr: "oauth2 client_id validation failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateOAuth2Auth(ctx, tt.oauth)
			if tt.expectErr != "" {
				assert.ErrorContains(t, err, tt.expectErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
