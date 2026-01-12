// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
)

func TestStripSecretsFromAuth_OAuth2(t *testing.T) {
	auth := &configv1.Authentication{
		AuthMethod: &configv1.Authentication_Oauth2{
			Oauth2: &configv1.OAuth2Auth{
				ClientId: &configv1.SecretValue{
					Value: &configv1.SecretValue_PlainText{PlainText: "client-id"},
				},
				ClientSecret: &configv1.SecretValue{
					Value: &configv1.SecretValue_PlainText{PlainText: "client-secret"},
				},
			},
		},
	}

	StripSecretsFromAuth(auth)

	oauth := auth.GetOauth2()
	assert.NotNil(t, oauth)

	// ClientID should be scrubbed now
	assert.Nil(t, oauth.ClientId.Value, "Plain text ClientId should be cleared")

	// ClientSecret should be scrubbed
	assert.Nil(t, oauth.ClientSecret.Value, "Plain text ClientSecret should be cleared")
}
