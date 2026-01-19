// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"context"
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
			oidc: &configv1.OIDCAuth{
				Issuer: proto.String("https://accounts.google.com"),
			},
			expectErr: "",
		},
		{
			name: "empty_issuer",
			oidc: &configv1.OIDCAuth{
				Issuer: proto.String(""),
			},
			expectErr: "oidc issuer is empty",
		},
		{
			name: "invalid_issuer_url",
			oidc: &configv1.OIDCAuth{
				Issuer: proto.String("not-a-url"),
			},
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
			th: &configv1.TrustedHeaderAuth{
				HeaderName:  proto.String("X-User-ID"),
				HeaderValue: proto.String("admin"),
			},
			expectErr: "",
		},
		{
			name: "empty_header_name",
			th: &configv1.TrustedHeaderAuth{
				HeaderName:  proto.String(""),
				HeaderValue: proto.String("admin"),
			},
			expectErr: "trusted header name is empty",
		},
		{
			name: "empty_header_value",
			th: &configv1.TrustedHeaderAuth{
				HeaderName:  proto.String("X-User-ID"),
				HeaderValue: proto.String(""),
			},
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

	tests := []struct {
		name      string
		oauth     *configv1.OAuth2Auth
		expectErr string
	}{
		{
			name: "invalid_issuer_url",
			oauth: &configv1.OAuth2Auth{
				TokenUrl:     proto.String(""), // empty token url
				IssuerUrl:    proto.String("not-a-url"),
				ClientId:     &configv1.SecretValue{Value: &configv1.SecretValue_EnvironmentVariable{EnvironmentVariable: "CLIENT_ID"}},
				ClientSecret: &configv1.SecretValue{Value: &configv1.SecretValue_EnvironmentVariable{EnvironmentVariable: "CLIENT_SECRET"}},
			},
			expectErr: "invalid oauth2 issuer_url",
		},
		{
			name: "invalid_token_url",
			oauth: &configv1.OAuth2Auth{
				TokenUrl:     proto.String("not-a-url"),
				ClientId:     &configv1.SecretValue{Value: &configv1.SecretValue_EnvironmentVariable{EnvironmentVariable: "CLIENT_ID"}},
				ClientSecret: &configv1.SecretValue{Value: &configv1.SecretValue_EnvironmentVariable{EnvironmentVariable: "CLIENT_SECRET"}},
			},
			expectErr: "invalid oauth2 token_url",
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
