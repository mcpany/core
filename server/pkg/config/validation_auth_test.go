// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"context"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

func TestValidate_ReproSilentFailures(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name         string
		auth         *configv1.Authentication
		expectErr    bool
		errSubstring string
	}{
		{
			name: "OAuth2 Missing ClientID and Secret",
			auth: &configv1.Authentication{
				AuthMethod: &configv1.Authentication_Oauth2{
					Oauth2: &configv1.OAuth2Auth{
						TokenUrl: proto.String("https://example.com/token"),
						// ClientID and ClientSecret are missing
					},
				},
			},
			expectErr:    true,
			errSubstring: "client_id is missing",
		},
		{
			name: "OIDC Missing Issuer",
			auth: &configv1.Authentication{
				AuthMethod: &configv1.Authentication_Oidc{
					Oidc: &configv1.OIDCAuth{
						// Issuer is missing
						Subject: proto.String("sub"),
					},
				},
			},
			expectErr:    true,
			errSubstring: "oidc issuer is empty",
		},
		{
			name: "TrustedHeader Missing HeaderName",
			auth: &configv1.Authentication{
				AuthMethod: &configv1.Authentication_TrustedHeader{
					TrustedHeader: &configv1.TrustedHeaderAuth{
						// HeaderName is missing
						HeaderValue: proto.String("secret"),
					},
				},
			},
			expectErr:    true,
			errSubstring: "trusted header name is empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Wrap in a dummy service to use Validate
			config := &configv1.McpAnyServerConfig{
				UpstreamServices: []*configv1.UpstreamServiceConfig{
					{
						Name: proto.String("test-service"),
						ServiceConfig: &configv1.UpstreamServiceConfig_HttpService{
							HttpService: &configv1.HttpUpstreamService{
								Address: proto.String("http://example.com"),
							},
						},
						UpstreamAuth: tt.auth,
					},
				},
			}

			errs := Validate(ctx, config, Server)

			if tt.expectErr {
				if len(errs) == 0 {
					t.Fatalf("Expected validation error containing %q, but got none", tt.errSubstring)
				}
				found := false
				for _, e := range errs {
					if assert.Contains(t, e.Err.Error(), tt.errSubstring) {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected error substring %q not found in errors: %v", tt.errSubstring, errs)
				}
			} else {
				assert.Empty(t, errs)
			}
		})
	}
}
