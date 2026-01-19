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
	u := &configv1.User{
		Id:   proto.String("user1"),
		// Name is NOT in User struct.
		Authentication: &configv1.Authentication{
			AuthMethod: &configv1.Authentication_BasicAuth{
				BasicAuth: &configv1.BasicAuth{
					Username: proto.String("user"),
					Password: &configv1.SecretValue{
						Value: &configv1.SecretValue_PlainText{PlainText: "secret"},
					},
					PasswordHash: proto.String("hash"),
				},
			},
		},
	}

	sanitized := SanitizeUser(u)
	assert.Equal(t, "user1", *sanitized.Id)

	ba := sanitized.Authentication.GetBasicAuth()
	assert.Equal(t, "user", *ba.Username)
	assert.Equal(t, RedactedString, ba.Password.GetPlainText())
	assert.Equal(t, RedactedString, *ba.PasswordHash)

	// Original should not be modified
	assert.Equal(t, "secret", u.Authentication.GetBasicAuth().Password.GetPlainText())
}

func TestSanitizeUser_Nil(t *testing.T) {
	assert.Nil(t, SanitizeUser(nil))
}

func TestSanitizeCredential(t *testing.T) {
	c := &configv1.Credential{
		Id: proto.String("cred1"),
		Authentication: &configv1.Authentication{
			AuthMethod: &configv1.Authentication_BearerToken{
				BearerToken: &configv1.BearerTokenAuth{ // BearerTokenAuth
					Token: &configv1.SecretValue{
						Value: &configv1.SecretValue_PlainText{PlainText: "token"},
					},
				},
			},
		},
		Token: &configv1.UserToken{
			AccessToken:  proto.String("access"),
			RefreshToken: proto.String("refresh"),
		},
	}

	sanitized := SanitizeCredential(c)
	assert.Equal(t, "cred1", *sanitized.Id)

	bt := sanitized.Authentication.GetBearerToken()
	assert.Equal(t, RedactedString, bt.Token.GetPlainText())

	tok := sanitized.Token
	assert.Equal(t, RedactedString, *tok.AccessToken)
	assert.Equal(t, RedactedString, *tok.RefreshToken)
}

func TestSanitizeCredential_Nil(t *testing.T) {
	assert.Nil(t, SanitizeCredential(nil))
}

func TestSanitizeAuthentication_ApiKey(t *testing.T) {
	a := &configv1.Authentication{
		AuthMethod: &configv1.Authentication_ApiKey{
			ApiKey: &configv1.APIKeyAuth{ // APIKeyAuth
				Value: &configv1.SecretValue{
					Value: &configv1.SecretValue_PlainText{PlainText: "key"},
				},
				VerificationValue: proto.String("verify"),
			},
		},
	}

	sanitized := SanitizeAuthentication(a)

	ak := sanitized.GetApiKey()
	assert.Equal(t, RedactedString, ak.Value.GetPlainText())
	assert.Equal(t, RedactedString, *ak.VerificationValue)
}

func TestSanitizeAuthentication_Oauth2(t *testing.T) {
	a := &configv1.Authentication{
		AuthMethod: &configv1.Authentication_Oauth2{
			Oauth2: &configv1.OAuth2Auth{ // OAuth2Auth
				ClientId: &configv1.SecretValue{
					Value: &configv1.SecretValue_PlainText{PlainText: "id"},
				},
				ClientSecret: &configv1.SecretValue{
					Value: &configv1.SecretValue_PlainText{PlainText: "secret"},
				},
			},
		},
	}

	sanitized := SanitizeAuthentication(a)
	o := sanitized.GetOauth2()
	assert.Equal(t, RedactedString, o.ClientId.GetPlainText())
	assert.Equal(t, RedactedString, o.ClientSecret.GetPlainText())
}

func TestSanitizeAuthentication_TrustedHeader(t *testing.T) {
	a := &configv1.Authentication{
		AuthMethod: &configv1.Authentication_TrustedHeader{
			TrustedHeader: &configv1.TrustedHeaderAuth{ // TrustedHeaderAuth
				HeaderName: proto.String("X-Auth"),
				HeaderValue: proto.String("secret"),
			},
		},
	}

	sanitized := SanitizeAuthentication(a)
	th := sanitized.GetTrustedHeader()
	assert.Equal(t, "X-Auth", *th.HeaderName)
	assert.Equal(t, RedactedString, *th.HeaderValue)
}

func TestSanitizeAuthentication_Nil(t *testing.T) {
	assert.Nil(t, SanitizeAuthentication(nil))
}

func TestSanitizeSecretValue_RemoteContent(t *testing.T) {
	s := &configv1.SecretValue{
		Value: &configv1.SecretValue_RemoteContent{
			RemoteContent: &configv1.RemoteContent{
				HttpUrl: proto.String("http://example.com"), // HttpUrl
				Auth: &configv1.Authentication{
					AuthMethod: &configv1.Authentication_BearerToken{
						BearerToken: &configv1.BearerTokenAuth{
							Token: &configv1.SecretValue{
								Value: &configv1.SecretValue_PlainText{PlainText: "token"},
							},
						},
					},
				},
			},
		},
	}

	sanitized := SanitizeSecretValue(s)
	rc := sanitized.GetRemoteContent()
	assert.Equal(t, RedactedString, rc.Auth.GetBearerToken().Token.GetPlainText())
}

func TestSanitizeSecretValue_Vault(t *testing.T) {
	s := &configv1.SecretValue{
		Value: &configv1.SecretValue_Vault{
			Vault: &configv1.VaultSecret{
				Path: proto.String("path"),
				Token: &configv1.SecretValue{
					Value: &configv1.SecretValue_PlainText{PlainText: "token"},
				},
			},
		},
	}

	sanitized := SanitizeSecretValue(s)
	v := sanitized.GetVault()
	assert.Equal(t, RedactedString, v.Token.GetPlainText())
}
