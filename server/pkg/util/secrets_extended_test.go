// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"context"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
)

func TestResolveSecret_Extended(t *testing.T) {
	t.Run("RemoteContent_APIKey_Error", func(t *testing.T) {
		// Create a secret that fails resolution
		failSecret := &configv1.SecretValue{}
		failSecret.SetEnvironmentVariable("NON_EXISTENT_VAR_FOR_API_KEY")

		apiKeyAuth := &configv1.APIKeyAuth{}
		apiKeyAuth.SetParamName("X-API-Key")
		apiKeyAuth.SetValue(failSecret)

		auth := &configv1.Authentication{}
		auth.SetApiKey(apiKeyAuth)

		remoteContent := &configv1.RemoteContent{}
		remoteContent.SetHttpUrl("http://localhost")
		remoteContent.SetAuth(auth)

		secret := &configv1.SecretValue{}
		secret.SetRemoteContent(remoteContent)

		_, err := ResolveSecret(context.Background(), secret)
		if err == nil {
			t.Error("expected error, got nil")
		}
	})

	t.Run("RemoteContent_BearerToken_Error", func(t *testing.T) {
		failSecret := &configv1.SecretValue{}
		failSecret.SetEnvironmentVariable("NON_EXISTENT_VAR_FOR_TOKEN")

		bearerAuth := &configv1.BearerTokenAuth{}
		bearerAuth.SetToken(failSecret)

		auth := &configv1.Authentication{}
		auth.SetBearerToken(bearerAuth)

		remoteContent := &configv1.RemoteContent{}
		remoteContent.SetHttpUrl("http://localhost")
		remoteContent.SetAuth(auth)

		secret := &configv1.SecretValue{}
		secret.SetRemoteContent(remoteContent)

		_, err := ResolveSecret(context.Background(), secret)
		if err == nil {
			t.Error("expected error, got nil")
		}
	})

	t.Run("RemoteContent_BasicAuth_Error", func(t *testing.T) {
		failSecret := &configv1.SecretValue{}
		failSecret.SetEnvironmentVariable("NON_EXISTENT_VAR_FOR_PASSWORD")

		basicAuth := &configv1.BasicAuth{}
		basicAuth.SetUsername("user")
		basicAuth.SetPassword(failSecret)

		auth := &configv1.Authentication{}
		auth.SetBasicAuth(basicAuth)

		remoteContent := &configv1.RemoteContent{}
		remoteContent.SetHttpUrl("http://localhost")
		remoteContent.SetAuth(auth)

		secret := &configv1.SecretValue{}
		secret.SetRemoteContent(remoteContent)

		_, err := ResolveSecret(context.Background(), secret)
		if err == nil {
			t.Error("expected error, got nil")
		}
	})

	t.Run("RemoteContent_OAuth2_ClientID_Error", func(t *testing.T) {
		failSecret := &configv1.SecretValue{}
		failSecret.SetEnvironmentVariable("NON_EXISTENT_VAR_FOR_CLIENT_ID")

		oauth2Auth := &configv1.OAuth2Auth{}
		oauth2Auth.SetClientId(failSecret)
		oauth2Auth.SetClientSecret(&configv1.SecretValue{
			Value: &configv1.SecretValue_PlainText{PlainText: "secret"},
		})
		oauth2Auth.SetTokenUrl("http://localhost/token")

		auth := &configv1.Authentication{}
		auth.SetOauth2(oauth2Auth)

		remoteContent := &configv1.RemoteContent{}
		remoteContent.SetHttpUrl("http://localhost")
		remoteContent.SetAuth(auth)

		secret := &configv1.SecretValue{}
		secret.SetRemoteContent(remoteContent)

		_, err := ResolveSecret(context.Background(), secret)
		if err == nil {
			t.Error("expected error, got nil")
		}
	})

	t.Run("RemoteContent_OAuth2_ClientSecret_Error", func(t *testing.T) {
		failSecret := &configv1.SecretValue{}
		failSecret.SetEnvironmentVariable("NON_EXISTENT_VAR_FOR_CLIENT_SECRET")

		oauth2Auth := &configv1.OAuth2Auth{}
		oauth2Auth.SetClientId(&configv1.SecretValue{
			Value: &configv1.SecretValue_PlainText{PlainText: "id"},
		})
		oauth2Auth.SetClientSecret(failSecret)
		oauth2Auth.SetTokenUrl("http://localhost/token")

		auth := &configv1.Authentication{}
		auth.SetOauth2(oauth2Auth)

		remoteContent := &configv1.RemoteContent{}
		remoteContent.SetHttpUrl("http://localhost")
		remoteContent.SetAuth(auth)

		secret := &configv1.SecretValue{}
		secret.SetRemoteContent(remoteContent)

		_, err := ResolveSecret(context.Background(), secret)
		if err == nil {
			t.Error("expected error, got nil")
		}
	})

	t.Run("Vault_Token_Error", func(t *testing.T) {
		failSecret := &configv1.SecretValue{}
		failSecret.SetEnvironmentVariable("NON_EXISTENT_VAR_FOR_VAULT_TOKEN")

		vaultSecret := &configv1.VaultSecret{}
		vaultSecret.SetAddress("http://localhost")
		vaultSecret.SetToken(failSecret)
		vaultSecret.SetPath("secret/foo")
		vaultSecret.SetKey("bar")

		secret := &configv1.SecretValue{}
		secret.SetVault(vaultSecret)

		_, err := ResolveSecret(context.Background(), secret)
		if err == nil {
			t.Error("expected error, got nil")
		}
	})
}
