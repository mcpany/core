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
	svc := &configv1.UpstreamServiceConfig{
		Name: proto.String(svcID),
		UpstreamAuth: &configv1.Authentication{
			AuthMethod: &configv1.Authentication_Oauth2{
				Oauth2: &configv1.OAuth2Auth{
					ClientId:         &configv1.SecretValue{Value: &configv1.SecretValue_PlainText{PlainText: "client-id"}},
					ClientSecret:     &configv1.SecretValue{Value: &configv1.SecretValue_PlainText{PlainText: "client-secret"}},
					AuthorizationUrl: proto.String("https://github.com/login/oauth/authorize"),
					TokenUrl:         proto.String("https://github.com/login/oauth/access_token"),
					Scopes:           proto.String("read:user"),
				},
			},
		},
	}
	err := store.SaveService(ctx, svc)
	require.NoError(t, err)

	// Test InitiateOAuth with Service ID
	url, state, err := am.InitiateOAuth(ctx, "user1", svcID, "", "http://localhost/cb")
	require.NoError(t, err)
	assert.Contains(t, url, "https://github.com/login/oauth/authorize")
	assert.Contains(t, url, "client_id=client-id")
	assert.NotEmpty(t, state)

	// Test InitiateOAuth with Invalid Service
	_, _, err = am.InitiateOAuth(ctx, "user1", "unknown", "", "http://localhost/cb")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")

	// Test InitiateOAuth with Storage Not Initialized
	amNoStore := NewManager()
	_, _, err = amNoStore.InitiateOAuth(ctx, "user1", svcID, "", "http://localhost/cb")
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
	cred := &configv1.Credential{
		Id: proto.String(credID),
		Authentication: &configv1.Authentication{
			AuthMethod: &configv1.Authentication_Oauth2{
				Oauth2: &configv1.OAuth2Auth{
					ClientId:         &configv1.SecretValue{Value: &configv1.SecretValue_PlainText{PlainText: "client-id"}},
					ClientSecret:     &configv1.SecretValue{Value: &configv1.SecretValue_PlainText{PlainText: "client-secret"}},
					AuthorizationUrl: proto.String("https://provider.com/auth"),
					TokenUrl:         proto.String("https://provider.com/token"),
					Scopes:           proto.String("scope1"),
				},
			},
		},
	}
	err := store.SaveCredential(ctx, cred)
	require.NoError(t, err)

	url, state, err := am.InitiateOAuth(ctx, "user1", "", credID, "http://localhost/cb")
	require.NoError(t, err)
	assert.Contains(t, url, "https://provider.com/auth")
	assert.Contains(t, url, "client_id=client-id")
	assert.NotEmpty(t, state)

	// Test invalid credential
	_, _, err = am.InitiateOAuth(ctx, "user1", "", "invalid-cred", "http://localhost/cb")
	assert.Error(t, err)

	// Test credential without Auth config
	credNoAuth := &configv1.Credential{Id: proto.String("no-auth")}
	err = store.SaveCredential(ctx, credNoAuth)
	require.NoError(t, err)
	_, _, err = am.InitiateOAuth(ctx, "user1", "", "no-auth", "http://localhost/cb")
	assert.Error(t, err)
}

func TestResolveSecretValue(t *testing.T) {
	t.Run("PlainText", func(t *testing.T) {
		sv := &configv1.SecretValue{
			Value: &configv1.SecretValue_PlainText{PlainText: "secret"},
		}
		assert.Equal(t, "secret", resolveSecretValue(sv))
	})

	t.Run("EnvironmentVariable", func(t *testing.T) {
		// Not implemented yet, should return empty
		sv := &configv1.SecretValue{
			Value: &configv1.SecretValue_EnvironmentVariable{EnvironmentVariable: "ENV_VAR"},
		}
		assert.Equal(t, "", resolveSecretValue(sv))
	})

	t.Run("Nil", func(t *testing.T) {
		assert.Equal(t, "", resolveSecretValue(nil))
	})
}

func TestHandleOAuthCallback_Validation(t *testing.T) {
	store := memory.NewStore()
	am := NewManager()
	am.SetStorage(store)
	ctx := context.Background()

	// Seed Service
	svcID := "github-service"
	svc := &configv1.UpstreamServiceConfig{
		Name: proto.String(svcID),
		UpstreamAuth: &configv1.Authentication{
			AuthMethod: &configv1.Authentication_Oauth2{
				Oauth2: &configv1.OAuth2Auth{
					ClientId:         &configv1.SecretValue{Value: &configv1.SecretValue_PlainText{PlainText: "client-id"}},
					ClientSecret:     &configv1.SecretValue{Value: &configv1.SecretValue_PlainText{PlainText: "client-secret"}},
					AuthorizationUrl: proto.String("https://github.com/login/oauth/authorize"),
					TokenUrl:         proto.String("https://github.com/login/oauth/access_token"),
				},
			},
		},
	}
	err := store.SaveService(ctx, svc)
	require.NoError(t, err)

	// Test Callback with Invalid Code behavior (Exchange will fail, verifying validation pass)
	// We can't easily mock oauth2.Config.Exchange without DI, but we can verify it reaches Exchange or fails earlier.

	err = am.HandleOAuthCallback(ctx, "user1", "unknown", "", "code", "http://localhost/cb")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")

	// Storage fail
	amNoStore := NewManager()
	err = amNoStore.HandleOAuthCallback(ctx, "user1", svcID, "", "code", "http://localhost/cb")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "storage not initialized")
}
