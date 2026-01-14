// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package util_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/util"
	"github.com/stretchr/testify/assert"
)

func TestResolveSecret_Coverage(t *testing.T) {
	t.Setenv("MCPANY_ALLOW_LOOPBACK_SECRETS", "true")

	t.Run("RemoteContent recursion errors", func(t *testing.T) {
		// Mock server not needed as it fails before request

		// API Key recursion error
		secret := &configv1.SecretValue{}
		remoteContent := &configv1.RemoteContent{}
		remoteContent.SetHttpUrl("http://example.com")

		apiKeyAuth := &configv1.APIKeyAuth{}
		apiKeyAuth.SetParamName("X-API-Key")
		// Recursive secret that fails
		badSecret := &configv1.SecretValue{}
		badSecret.SetEnvironmentVariable("NON_EXISTENT_VAR_FOR_TEST")
		apiKeyAuth.SetValue(badSecret)

		auth := &configv1.Authentication{}
		auth.SetApiKey(apiKeyAuth)
		remoteContent.SetAuth(auth)
		secret.SetRemoteContent(remoteContent)

		_, err := util.ResolveSecret(context.Background(), secret)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to resolve api key")

		// Bearer Token recursion error
		bearerTokenAuth := &configv1.BearerTokenAuth{}
		bearerTokenAuth.SetToken(badSecret)
		auth.SetApiKey(nil)
		auth.SetBearerToken(bearerTokenAuth)
		remoteContent.SetAuth(auth)
		secret.SetRemoteContent(remoteContent)

		_, err = util.ResolveSecret(context.Background(), secret)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to resolve bearer token")

		// Basic Auth Password recursion error
		basicAuth := &configv1.BasicAuth{}
		basicAuth.SetUsername("user")
		basicAuth.SetPassword(badSecret)
		auth.SetBearerToken(nil)
		auth.SetBasicAuth(basicAuth)
		remoteContent.SetAuth(auth)
		secret.SetRemoteContent(remoteContent)

		_, err = util.ResolveSecret(context.Background(), secret)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to resolve password")
	})

	t.Run("RemoteContent OAuth2", func(t *testing.T) {
		// Mock Token Endpoint
		tokenServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "POST", r.Method)
			// Check basic auth for client id/secret
			user, pass, ok := r.BasicAuth()
			assert.True(t, ok)
			assert.Equal(t, "my-client-id", user)
			assert.Equal(t, "my-client-secret", pass)

			w.Header().Set("Content-Type", "application/json")
			_, _ = fmt.Fprint(w, `{"access_token": "my-oauth-token", "token_type": "Bearer", "expires_in": 3600}`)
		}))
		defer tokenServer.Close()

		// Mock Resource Server
		resourceServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "Bearer my-oauth-token", r.Header.Get("Authorization"))
			_, _ = fmt.Fprint(w, "my-remote-secret")
		}))
		defer resourceServer.Close()

		oauth2Auth := &configv1.OAuth2Auth{}

		clientID := &configv1.SecretValue{}
		clientID.SetPlainText("my-client-id")
		oauth2Auth.SetClientId(clientID)

		clientSecret := &configv1.SecretValue{}
		clientSecret.SetPlainText("my-client-secret")
		oauth2Auth.SetClientSecret(clientSecret)

		oauth2Auth.SetTokenUrl(tokenServer.URL)
		oauth2Auth.SetScopes("scope1 scope2")

		auth := &configv1.Authentication{}
		auth.SetOauth2(oauth2Auth)

		remoteContent := &configv1.RemoteContent{}
		remoteContent.SetHttpUrl(resourceServer.URL)
		remoteContent.SetAuth(auth)

		secret := &configv1.SecretValue{}
		secret.SetRemoteContent(remoteContent)

		resolved, err := util.ResolveSecret(context.Background(), secret)
		assert.NoError(t, err)
		assert.Equal(t, "my-remote-secret", resolved)
	})

	t.Run("RemoteContent OAuth2 recursion errors", func(t *testing.T) {
		badSecret := &configv1.SecretValue{}
		badSecret.SetEnvironmentVariable("NON_EXISTENT_VAR_FOR_TEST")

		oauth2Auth := &configv1.OAuth2Auth{}

		// Client ID fail
		oauth2Auth.SetClientId(badSecret)
		// Need dummy secret to pass nil check inside proto getter if accessed via getter?
		// No, getter returns nil if not set, but we set it.

		// Need dummy client secret
		goodSecret := &configv1.SecretValue{}
		goodSecret.SetPlainText("secret")
		oauth2Auth.SetClientSecret(goodSecret)

		auth := &configv1.Authentication{}
		auth.SetOauth2(oauth2Auth)
		remoteContent := &configv1.RemoteContent{}
		remoteContent.SetHttpUrl("http://example.com")
		remoteContent.SetAuth(auth)
		secret := &configv1.SecretValue{}
		secret.SetRemoteContent(remoteContent)

		_, err := util.ResolveSecret(context.Background(), secret)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to resolve client id")

		// Client Secret fail
		oauth2Auth.SetClientId(goodSecret)
		oauth2Auth.SetClientSecret(badSecret)
		_, err = util.ResolveSecret(context.Background(), secret)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to resolve client secret")
	})

	t.Run("RemoteContent OAuth2 Token Error", func(t *testing.T) {
		// Mock Token Endpoint that returns error
		tokenServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusBadRequest)
		}))
		defer tokenServer.Close()

		oauth2Auth := &configv1.OAuth2Auth{}
		clientID := &configv1.SecretValue{}
		clientID.SetPlainText("id")
		oauth2Auth.SetClientId(clientID)
		clientSecret := &configv1.SecretValue{}
		clientSecret.SetPlainText("secret")
		oauth2Auth.SetClientSecret(clientSecret)
		oauth2Auth.SetTokenUrl(tokenServer.URL)

		auth := &configv1.Authentication{}
		auth.SetOauth2(oauth2Auth)
		remoteContent := &configv1.RemoteContent{}
		remoteContent.SetHttpUrl("http://example.com")
		remoteContent.SetAuth(auth)
		secret := &configv1.SecretValue{}
		secret.SetRemoteContent(remoteContent)

		_, err := util.ResolveSecret(context.Background(), secret)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to get oauth2 token")
	})
}
