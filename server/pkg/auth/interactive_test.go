// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package auth

import (
	"context"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/storage/memory"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func TestInitiateOAuth(t *testing.T) {
	store := memory.NewStore()
	am := NewManager()
	am.SetStorage(store)
	ctx := context.Background()

	// Seed Service with OAuth
	svcID := "github-service"
	svc := configv1.UpstreamServiceConfig_builder{
		Name: proto.String(svcID),
		UpstreamAuth: configv1.Authentication_builder{
			Oauth2: configv1.OAuth2Auth_builder{
				ClientId:         configv1.SecretValue_builder{PlainText: proto.String("client-id")}.Build(),
				ClientSecret:     configv1.SecretValue_builder{PlainText: proto.String("client-secret")}.Build(),
				AuthorizationUrl: proto.String("https://github.com/login/oauth/authorize"),
				TokenUrl:         proto.String("https://github.com/login/oauth/access_token"),
				Scopes:           proto.String("read:user"),
			}.Build(),
		}.Build(),
	}.Build()
	err := store.SaveService(ctx, svc)
	require.NoError(t, err)

	// Test InitiateOAuth with Service ID
	url, state, err := am.InitiateOAuth(ctx, "user1", svcID, "", "http://127.0.0.1/cb")
	require.NoError(t, err)
	assert.Contains(t, url, "https://github.com/login/oauth/authorize")
	assert.Contains(t, url, "client_id=client-id")
	assert.NotEmpty(t, state)

	// Test InitiateOAuth with Invalid Service
	_, _, err = am.InitiateOAuth(ctx, "user1", "unknown", "", "http://127.0.0.1/cb")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")

	// Test InitiateOAuth with Storage Not Initialized
	amNoStore := NewManager()
	_, _, err = amNoStore.InitiateOAuth(ctx, "user1", svcID, "", "http://127.0.0.1/cb")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "storage not initialized")
}

func TestInitiateOAuth_Credential(t *testing.T) {
	store := memory.NewStore()
	am := NewManager()
	am.SetStorage(store)
	ctx := context.Background()

	// Seed Credential with OAuth
	credID := "cred-1"
	cred := configv1.Credential_builder{
		Id: proto.String(credID),
		Authentication: configv1.Authentication_builder{
			Oauth2: configv1.OAuth2Auth_builder{
				ClientId:         configv1.SecretValue_builder{PlainText: proto.String("client-id")}.Build(),
				ClientSecret:     configv1.SecretValue_builder{PlainText: proto.String("client-secret")}.Build(),
				AuthorizationUrl: proto.String("https://provider.com/auth"),
				TokenUrl:         proto.String("https://provider.com/token"),
				Scopes:           proto.String("scope1"),
			}.Build(),
		}.Build(),
	}.Build()
	err := store.SaveCredential(ctx, cred)
	require.NoError(t, err)

	url, state, err := am.InitiateOAuth(ctx, "user1", "", credID, "http://127.0.0.1/cb")
	require.NoError(t, err)
	assert.Contains(t, url, "https://provider.com/auth")
	assert.Contains(t, url, "client_id=client-id")
	assert.NotEmpty(t, state)

	// Test invalid credential
	_, _, err = am.InitiateOAuth(ctx, "user1", "", "invalid-cred", "http://127.0.0.1/cb")
	assert.Error(t, err)

	// Test credential without Auth config
	credNoAuth := configv1.Credential_builder{Id: proto.String("no-auth")}.Build()
	err = store.SaveCredential(ctx, credNoAuth)
	require.NoError(t, err)
	_, _, err = am.InitiateOAuth(ctx, "user1", "", "no-auth", "http://127.0.0.1/cb")
	assert.Error(t, err)
}

func TestResolveSecretValue(t *testing.T) {
	// Previously tested the removed resolveSecretValue function.
	// Now tests that InitiateOAuth correctly resolves secrets via util.ResolveSecret.
	store := memory.NewStore()
	am := NewManager()
	am.SetStorage(store)
	ctx := context.Background()

	t.Run("PlainText", func(t *testing.T) {
		// Test that plain-text secrets are resolved correctly through auth
		svcID := "plain-text-secret-service"
		svc := configv1.UpstreamServiceConfig_builder{
			Name: proto.String(svcID),
			UpstreamAuth: configv1.Authentication_builder{
				Oauth2: configv1.OAuth2Auth_builder{
					ClientId:         configv1.SecretValue_builder{PlainText: proto.String("plain-client-id")}.Build(),
					ClientSecret:     configv1.SecretValue_builder{PlainText: proto.String("plain-client-secret")}.Build(),
					AuthorizationUrl: proto.String("https://auth.example.com/authorize"),
					TokenUrl:         proto.String("https://auth.example.com/token"),
					Scopes:           proto.String("read"),
				}.Build(),
			}.Build(),
		}.Build()
		require.NoError(t, store.SaveService(ctx, svc))

		url, state, err := am.InitiateOAuth(ctx, "user1", svcID, "", "http://127.0.0.1/cb")
		require.NoError(t, err)
		assert.Contains(t, url, "plain-client-id")
		assert.NotEmpty(t, state)
	})

	t.Run("EnvironmentVariable", func(t *testing.T) {
		// Test that environment variable secrets are resolved through auth
		// Allow test env vars via MCPANY_ALLOWED_ENV
		t.Setenv("MCPANY_ALLOWED_ENV", "TEST_AUTH_CLIENT_ID,TEST_AUTH_CLIENT_SECRET")
		t.Setenv("TEST_AUTH_CLIENT_ID", "env-client-id")
		t.Setenv("TEST_AUTH_CLIENT_SECRET", "env-client-secret")

		svcID := "env-secret-service"
		svc := configv1.UpstreamServiceConfig_builder{
			Name: proto.String(svcID),
			UpstreamAuth: configv1.Authentication_builder{
				Oauth2: configv1.OAuth2Auth_builder{
					ClientId:         configv1.SecretValue_builder{EnvironmentVariable: proto.String("TEST_AUTH_CLIENT_ID")}.Build(),
					ClientSecret:     configv1.SecretValue_builder{EnvironmentVariable: proto.String("TEST_AUTH_CLIENT_SECRET")}.Build(),
					AuthorizationUrl: proto.String("https://env.example.com/authorize"),
					TokenUrl:         proto.String("https://env.example.com/token"),
					Scopes:           proto.String("write"),
				}.Build(),
			}.Build(),
		}.Build()
		require.NoError(t, store.SaveService(ctx, svc))

		url, state, err := am.InitiateOAuth(ctx, "user2", svcID, "", "http://127.0.0.1/cb")
		require.NoError(t, err)
		assert.Contains(t, url, "env-client-id")
		assert.NotEmpty(t, state)
	})
}

func TestHandleOAuthCallback_Validation(t *testing.T) {
	store := memory.NewStore()
	am := NewManager()
	am.SetStorage(store)
	ctx := context.Background()

	// Seed Service
	svcID := "github-service"
	svc := configv1.UpstreamServiceConfig_builder{
		Name: proto.String(svcID),
		UpstreamAuth: configv1.Authentication_builder{
			Oauth2: configv1.OAuth2Auth_builder{
				ClientId:         configv1.SecretValue_builder{PlainText: proto.String("client-id")}.Build(),
				ClientSecret:     configv1.SecretValue_builder{PlainText: proto.String("client-secret")}.Build(),
				AuthorizationUrl: proto.String("https://github.com/login/oauth/authorize"),
				TokenUrl:         proto.String("https://github.com/login/oauth/access_token"),
			}.Build(),
		}.Build(),
	}.Build()
	err := store.SaveService(ctx, svc)
	require.NoError(t, err)

	// Test Callback with Invalid Code behavior (Exchange will fail, verifying validation pass)
	// We can't easily mock oauth2.Config.Exchange without DI, but we can verify it reaches Exchange or fails earlier.

	err = am.HandleOAuthCallback(ctx, "user1", "unknown", "", "code", "http://127.0.0.1/cb")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")

	// Storage fail
	amNoStore := NewManager()
	err = amNoStore.HandleOAuthCallback(ctx, "user1", svcID, "", "code", "http://127.0.0.1/cb")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "storage not initialized")
}
