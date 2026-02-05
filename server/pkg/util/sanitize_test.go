// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

func TestSanitizeUser(t *testing.T) {
	u := configv1.User_builder{
		Id: proto.String("user1"),
		Authentication: configv1.Authentication_builder{
			BasicAuth: configv1.BasicAuth_builder{
				Username: proto.String("user"),
				Password: configv1.SecretValue_builder{
					PlainText: proto.String("secret"),
				}.Build(),
				PasswordHash: proto.String("hash"),
			}.Build(),
		}.Build(),
	}.Build()

	sanitized := SanitizeUser(u)
	assert.Equal(t, "user1", sanitized.GetId())

	ba := sanitized.GetAuthentication().GetBasicAuth()
	assert.Equal(t, "user", ba.GetUsername())
	assert.Equal(t, RedactedString, ba.GetPassword().GetPlainText())
	assert.Equal(t, RedactedString, ba.GetPasswordHash())

	// Original should not be modified
	assert.Equal(t, "secret", u.GetAuthentication().GetBasicAuth().GetPassword().GetPlainText())
}

func TestSanitizeUser_Nil(t *testing.T) {
	assert.Nil(t, SanitizeUser(nil))
}

func TestSanitizeCredential(t *testing.T) {
	c := configv1.Credential_builder{
		Id: proto.String("cred1"),
		Authentication: configv1.Authentication_builder{
			BearerToken: configv1.BearerTokenAuth_builder{
				Token: configv1.SecretValue_builder{
					PlainText: proto.String("token"),
				}.Build(),
			}.Build(),
		}.Build(),
		Token: configv1.UserToken_builder{
			AccessToken:  proto.String("access"),
			RefreshToken: proto.String("refresh"),
		}.Build(),
	}.Build()

	sanitized := SanitizeCredential(c)
	assert.Equal(t, "cred1", sanitized.GetId())

	bt := sanitized.GetAuthentication().GetBearerToken()
	assert.Equal(t, RedactedString, bt.GetToken().GetPlainText())

	tok := sanitized.GetToken()
	assert.Equal(t, RedactedString, tok.GetAccessToken())
	assert.Equal(t, RedactedString, tok.GetRefreshToken())
}

func TestSanitizeCredential_Nil(t *testing.T) {
	assert.Nil(t, SanitizeCredential(nil))
}

func TestSanitizeAuthentication_ApiKey(t *testing.T) {
	a := configv1.Authentication_builder{
		ApiKey: configv1.APIKeyAuth_builder{
			Value: configv1.SecretValue_builder{
				PlainText: proto.String("key"),
			}.Build(),
			VerificationValue: proto.String("verify"),
		}.Build(),
	}.Build()

	sanitized := SanitizeAuthentication(a)

	ak := sanitized.GetApiKey()
	assert.Equal(t, RedactedString, ak.GetValue().GetPlainText())
	assert.Equal(t, RedactedString, ak.GetVerificationValue())
}

func TestSanitizeAuthentication_Oauth2(t *testing.T) {
	a := configv1.Authentication_builder{
		Oauth2: configv1.OAuth2Auth_builder{
			ClientId: configv1.SecretValue_builder{
				PlainText: proto.String("id"),
			}.Build(),
			ClientSecret: configv1.SecretValue_builder{
				PlainText: proto.String("secret"),
			}.Build(),
		}.Build(),
	}.Build()

	sanitized := SanitizeAuthentication(a)
	o := sanitized.GetOauth2()
	assert.Equal(t, RedactedString, o.GetClientId().GetPlainText())
	assert.Equal(t, RedactedString, o.GetClientSecret().GetPlainText())
}

func TestSanitizeAuthentication_TrustedHeader(t *testing.T) {
	a := configv1.Authentication_builder{
		TrustedHeader: configv1.TrustedHeaderAuth_builder{
			HeaderName:  proto.String("X-Auth"),
			HeaderValue: proto.String("secret"),
		}.Build(),
	}.Build()

	sanitized := SanitizeAuthentication(a)
	th := sanitized.GetTrustedHeader()
	assert.Equal(t, "X-Auth", th.GetHeaderName())
	assert.Equal(t, RedactedString, th.GetHeaderValue())
}

func TestSanitizeAuthentication_Nil(t *testing.T) {
	assert.Nil(t, SanitizeAuthentication(nil))
}

func TestSanitizeSecretValue_RemoteContent(t *testing.T) {
	s := configv1.SecretValue_builder{
		RemoteContent: configv1.RemoteContent_builder{
			HttpUrl: proto.String("http://example.com"),
			Auth: configv1.Authentication_builder{
				BearerToken: configv1.BearerTokenAuth_builder{
					Token: configv1.SecretValue_builder{
						PlainText: proto.String("token"),
					}.Build(),
				}.Build(),
			}.Build(),
		}.Build(),
	}.Build()

	sanitized := SanitizeSecretValue(s)
	rc := sanitized.GetRemoteContent()
	assert.Equal(t, RedactedString, rc.GetAuth().GetBearerToken().GetToken().GetPlainText())
}

func TestSanitizeSecretValue_Vault(t *testing.T) {
	s := configv1.SecretValue_builder{
		Vault: configv1.VaultSecret_builder{
			Path: proto.String("path"),
			Token: configv1.SecretValue_builder{
				PlainText: proto.String("token"),
			}.Build(),
		}.Build(),
	}.Build()

	sanitized := SanitizeSecretValue(s)
	v := sanitized.GetVault()
	assert.Equal(t, RedactedString, v.GetToken().GetPlainText())
}
