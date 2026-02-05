// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"context"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"google.golang.org/protobuf/proto"
)

func TestResolveSecret_Extended(t *testing.T) {
	t.Run("RemoteContent_APIKey_Error", func(t *testing.T) {
		// Create a secret that fails resolution
		failSecret := configv1.SecretValue_builder{
			EnvironmentVariable: proto.String("NON_EXISTENT_VAR_FOR_API_KEY"),
		}.Build()

		apiKeyAuth := configv1.APIKeyAuth_builder{
			ParamName: proto.String("X-API-Key"),
			Value:     failSecret,
		}.Build()

		auth := configv1.Authentication_builder{
			ApiKey: apiKeyAuth,
		}.Build()

		remoteContent := configv1.RemoteContent_builder{
			HttpUrl: proto.String("http://127.0.0.1"),
			Auth:    auth,
		}.Build()

		secret := configv1.SecretValue_builder{
			RemoteContent: remoteContent,
		}.Build()

		_, err := ResolveSecret(context.Background(), secret)
		if err == nil {
			t.Error("expected error, got nil")
		}
	})

	t.Run("RemoteContent_BearerToken_Error", func(t *testing.T) {
		failSecret := configv1.SecretValue_builder{
			EnvironmentVariable: proto.String("NON_EXISTENT_VAR_FOR_TOKEN"),
		}.Build()

		bearerAuth := configv1.BearerTokenAuth_builder{
			Token: failSecret,
		}.Build()

		auth := configv1.Authentication_builder{
			BearerToken: bearerAuth,
		}.Build()

		remoteContent := configv1.RemoteContent_builder{
			HttpUrl: proto.String("http://127.0.0.1"),
			Auth:    auth,
		}.Build()

		secret := configv1.SecretValue_builder{
			RemoteContent: remoteContent,
		}.Build()

		_, err := ResolveSecret(context.Background(), secret)
		if err == nil {
			t.Error("expected error, got nil")
		}
	})

	t.Run("RemoteContent_BasicAuth_Error", func(t *testing.T) {
		failSecret := configv1.SecretValue_builder{
			EnvironmentVariable: proto.String("NON_EXISTENT_VAR_FOR_PASSWORD"),
		}.Build()

		basicAuth := configv1.BasicAuth_builder{
			Username: proto.String("user"),
			Password: failSecret,
		}.Build()

		auth := configv1.Authentication_builder{
			BasicAuth: basicAuth,
		}.Build()

		remoteContent := configv1.RemoteContent_builder{
			HttpUrl: proto.String("http://127.0.0.1"),
			Auth:    auth,
		}.Build()

		secret := configv1.SecretValue_builder{
			RemoteContent: remoteContent,
		}.Build()

		_, err := ResolveSecret(context.Background(), secret)
		if err == nil {
			t.Error("expected error, got nil")
		}
	})

	t.Run("RemoteContent_OAuth2_ClientID_Error", func(t *testing.T) {
		failSecret := configv1.SecretValue_builder{
			EnvironmentVariable: proto.String("NON_EXISTENT_VAR_FOR_CLIENT_ID"),
		}.Build()

		oauth2Auth := configv1.OAuth2Auth_builder{
			ClientId: failSecret,
			ClientSecret: configv1.SecretValue_builder{
				PlainText: proto.String("secret"),
			}.Build(),
			TokenUrl: proto.String("http://127.0.0.1/token"),
		}.Build()

		auth := configv1.Authentication_builder{
			Oauth2: oauth2Auth,
		}.Build()

		remoteContent := configv1.RemoteContent_builder{
			HttpUrl: proto.String("http://127.0.0.1"),
			Auth:    auth,
		}.Build()

		secret := configv1.SecretValue_builder{
			RemoteContent: remoteContent,
		}.Build()

		_, err := ResolveSecret(context.Background(), secret)
		if err == nil {
			t.Error("expected error, got nil")
		}
	})

	t.Run("RemoteContent_OAuth2_ClientSecret_Error", func(t *testing.T) {
		failSecret := configv1.SecretValue_builder{
			EnvironmentVariable: proto.String("NON_EXISTENT_VAR_FOR_CLIENT_SECRET"),
		}.Build()

		oauth2Auth := configv1.OAuth2Auth_builder{
			ClientId: configv1.SecretValue_builder{
				PlainText: proto.String("id"),
			}.Build(),
			ClientSecret: failSecret,
			TokenUrl:     proto.String("http://127.0.0.1/token"),
		}.Build()

		auth := configv1.Authentication_builder{
			Oauth2: oauth2Auth,
		}.Build()

		remoteContent := configv1.RemoteContent_builder{
			HttpUrl: proto.String("http://127.0.0.1"),
			Auth:    auth,
		}.Build()

		secret := configv1.SecretValue_builder{
			RemoteContent: remoteContent,
		}.Build()

		_, err := ResolveSecret(context.Background(), secret)
		if err == nil {
			t.Error("expected error, got nil")
		}
	})

	t.Run("Vault_Token_Error", func(t *testing.T) {
		failSecret := configv1.SecretValue_builder{
			EnvironmentVariable: proto.String("NON_EXISTENT_VAR_FOR_VAULT_TOKEN"),
		}.Build()

		vaultSecret := configv1.VaultSecret_builder{
			Address: proto.String("http://127.0.0.1"),
			Token:   failSecret,
			Path:    proto.String("secret/foo"),
			Key:     proto.String("bar"),
		}.Build()

		secret := configv1.SecretValue_builder{
			Vault: vaultSecret,
		}.Build()

		_, err := ResolveSecret(context.Background(), secret)
		if err == nil {
			t.Error("expected error, got nil")
		}
	})
}
