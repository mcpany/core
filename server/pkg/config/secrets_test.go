// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

func TestStripSecretsFromService(t *testing.T) {
	svc := &configv1.UpstreamServiceConfig{
		Name: proto.String("test-service"),
		UpstreamAuthentication: &configv1.UpstreamAuthentication{
			AuthMethod: &configv1.UpstreamAuthentication_ApiKey{
				ApiKey: &configv1.UpstreamAPIKeyAuth{
					HeaderName: proto.String("X-API-Key"),
					ApiKey: &configv1.SecretValue{
						Value: &configv1.SecretValue_PlainText{PlainText: "secret-key"},
					},
				},
			},
		},
	}

	StripSecretsFromService(svc)

	assert.NotNil(t, svc.UpstreamAuthentication)
	assert.NotNil(t, svc.UpstreamAuthentication.GetApiKey())
	assert.NotNil(t, svc.UpstreamAuthentication.GetApiKey().ApiKey)
	assert.Nil(t, svc.UpstreamAuthentication.GetApiKey().ApiKey.Value, "Plain text secret should be cleared")
}

func TestStripSecretsFromProfile(t *testing.T) {
	profile := &configv1.ProfileDefinition{
		Name: proto.String("test-profile"),
		Secrets: map[string]*configv1.SecretValue{
			"TEST_SECRET": {Value: &configv1.SecretValue_PlainText{PlainText: "secret-value"}},
		},
	}

	StripSecretsFromProfile(profile)

	secret := profile.Secrets["TEST_SECRET"]
	assert.NotNil(t, secret)
	assert.Nil(t, secret.Value, "Plain text secret should be cleared")
}

func TestStripSecretsFromCollection(t *testing.T) {
	collection := &configv1.UpstreamServiceCollectionShare{
		Name: proto.String("test-collection"),
		Services: []*configv1.UpstreamServiceConfig{
			{
				Name: proto.String("svc1"),
				UpstreamAuthentication: &configv1.UpstreamAuthentication{
					AuthMethod: &configv1.UpstreamAuthentication_BasicAuth{
						BasicAuth: &configv1.UpstreamBasicAuth{
							Username: proto.String("user"),
							Password: &configv1.SecretValue{
								Value: &configv1.SecretValue_PlainText{PlainText: "secret-password"},
							},
						},
					},
				},
			},
		},
	}

	StripSecretsFromCollection(collection)

	svc := collection.Services[0]
	assert.NotNil(t, svc.UpstreamAuthentication)
	assert.Nil(t, svc.UpstreamAuthentication.GetBasicAuth().Password.Value, "Plain text secret should be cleared")
}

func TestHydrateSecretsInService(t *testing.T) {
	svc := &configv1.UpstreamServiceConfig{
		Name: proto.String("test-service"),
		UpstreamAuthentication: &configv1.UpstreamAuthentication{
			AuthMethod: &configv1.UpstreamAuthentication_ApiKey{
				ApiKey: &configv1.UpstreamAPIKeyAuth{
					HeaderName: proto.String("X-API-Key"),
					ApiKey: &configv1.SecretValue{
						Value: &configv1.SecretValue_EnvironmentVariable{EnvironmentVariable: "API_KEY_VAR"},
					},
				},
			},
		},
	}

	secrets := map[string]*configv1.SecretValue{
		"API_KEY_VAR": {Value: &configv1.SecretValue_PlainText{PlainText: "resolved-secret"}},
	}

	HydrateSecretsInService(svc, secrets)

	val := svc.UpstreamAuthentication.GetApiKey().ApiKey.Value.(*configv1.SecretValue_PlainText)
	assert.Equal(t, "resolved-secret", val.PlainText)
}
