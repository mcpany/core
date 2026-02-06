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

func TestResolveSecret_OAuth2(t *testing.T) {
	t.Setenv("MCPANY_ALLOW_LOOPBACK_SECRETS", "true")

	// 1. Setup Mock Token Server
	tokenServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		// Verify Client Credentials
		username, password, ok := r.BasicAuth()
		if !ok || username != "my-client-id" || password != "my-client-secret" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		// Return access token
		_, _ = fmt.Fprint(w, `{"access_token": "mock-access-token", "token_type": "Bearer", "expires_in": 3600}`)
	}))
	defer tokenServer.Close()

	// 2. Setup Mock Content Server
	contentServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify Bearer Token
		authHeader := r.Header.Get("Authorization")
		if authHeader != "Bearer mock-access-token" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		_, _ = fmt.Fprint(w, "my-remote-secret-via-oauth")
	}))
	defer contentServer.Close()

	// 3. Construct Secret Config
	clientIDSecret := &configv1.SecretValue{}
	clientIDSecret.SetPlainText("my-client-id")

	clientSecretSecret := &configv1.SecretValue{}
	clientSecretSecret.SetPlainText("my-client-secret")

	oauthConfig := &configv1.OAuth2Auth{}
	oauthConfig.SetClientId(clientIDSecret)
	oauthConfig.SetClientSecret(clientSecretSecret)
	oauthConfig.SetTokenUrl(tokenServer.URL)
	oauthConfig.SetScopes("scope1 scope2")

	auth := &configv1.Authentication{}
	auth.SetOauth2(oauthConfig)

	remoteContent := &configv1.RemoteContent{}
	remoteContent.SetHttpUrl(contentServer.URL)
	remoteContent.SetAuth(auth)

	secret := &configv1.SecretValue{}
	secret.SetRemoteContent(remoteContent)

	// 4. Test
	resolved, err := util.ResolveSecret(context.Background(), secret)
	assert.NoError(t, err)
	assert.Equal(t, "my-remote-secret-via-oauth", resolved)
}

func TestResolveSecret_OAuth2_TokenFail(t *testing.T) {
	// Token server fails
	tokenServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer tokenServer.Close()

	clientIDSecret := &configv1.SecretValue{}
	clientIDSecret.SetPlainText("my-client-id")

	clientSecretSecret := &configv1.SecretValue{}
	clientSecretSecret.SetPlainText("my-client-secret")

	oauthConfig := &configv1.OAuth2Auth{}
	oauthConfig.SetClientId(clientIDSecret)
	oauthConfig.SetClientSecret(clientSecretSecret)
	oauthConfig.SetTokenUrl(tokenServer.URL)

	auth := &configv1.Authentication{}
	auth.SetOauth2(oauthConfig)

	remoteContent := &configv1.RemoteContent{}
	remoteContent.SetHttpUrl("http://example.com")
	remoteContent.SetAuth(auth)

	secret := &configv1.SecretValue{}
	secret.SetRemoteContent(remoteContent)

	_, err := util.ResolveSecret(context.Background(), secret)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get oauth2 token")
}

func TestResolveSecret_RemoteContent_AuthErrors(t *testing.T) {
	t.Setenv("MCPANY_ALLOW_LOOPBACK_SECRETS", "true")

	// 1. Bearer Token Resolution Error
	t.Run("BearerTokenError", func(t *testing.T) {
		tokenSecret := &configv1.SecretValue{}
		tokenSecret.SetEnvironmentVariable("MISSING_ENV_VAR")

		bearerAuth := &configv1.BearerTokenAuth{}
		bearerAuth.SetToken(tokenSecret)

		auth := &configv1.Authentication{}
		auth.SetBearerToken(bearerAuth)

		remoteContent := &configv1.RemoteContent{}
		remoteContent.SetHttpUrl("http://example.com")
		remoteContent.SetAuth(auth)

		secret := &configv1.SecretValue{}
		secret.SetRemoteContent(remoteContent)

		_, err := util.ResolveSecret(context.Background(), secret)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to resolve bearer token")
	})

	// 2. API Key Resolution Error
	t.Run("APIKeyError", func(t *testing.T) {
		keySecret := &configv1.SecretValue{}
		keySecret.SetEnvironmentVariable("MISSING_ENV_VAR")

		apiKeyAuth := &configv1.APIKeyAuth{}
		apiKeyAuth.SetValue(keySecret)
		apiKeyAuth.SetParamName("X-API-Key")

		auth := &configv1.Authentication{}
		auth.SetApiKey(apiKeyAuth)

		remoteContent := &configv1.RemoteContent{}
		remoteContent.SetHttpUrl("http://example.com")
		remoteContent.SetAuth(auth)

		secret := &configv1.SecretValue{}
		secret.SetRemoteContent(remoteContent)

		_, err := util.ResolveSecret(context.Background(), secret)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to resolve api key")
	})

	// 3. Basic Auth Password Resolution Error
	t.Run("BasicAuthError", func(t *testing.T) {
		passSecret := &configv1.SecretValue{}
		passSecret.SetEnvironmentVariable("MISSING_ENV_VAR")

		basicAuth := &configv1.BasicAuth{}
		basicAuth.SetPassword(passSecret)
		basicAuth.SetUsername("user")

		auth := &configv1.Authentication{}
		auth.SetBasicAuth(basicAuth)

		remoteContent := &configv1.RemoteContent{}
		remoteContent.SetHttpUrl("http://example.com")
		remoteContent.SetAuth(auth)

		secret := &configv1.SecretValue{}
		secret.SetRemoteContent(remoteContent)

		_, err := util.ResolveSecret(context.Background(), secret)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to resolve password")
	})

	// 4. OAuth2 Client ID Resolution Error
	t.Run("OAuth2ClientIDError", func(t *testing.T) {
		idSecret := &configv1.SecretValue{}
		idSecret.SetEnvironmentVariable("MISSING_ENV_VAR")

		oauthConfig := &configv1.OAuth2Auth{}
		oauthConfig.SetClientId(idSecret)

		auth := &configv1.Authentication{}
		auth.SetOauth2(oauthConfig)

		remoteContent := &configv1.RemoteContent{}
		remoteContent.SetHttpUrl("http://example.com")
		remoteContent.SetAuth(auth)

		secret := &configv1.SecretValue{}
		secret.SetRemoteContent(remoteContent)

		_, err := util.ResolveSecret(context.Background(), secret)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to resolve client id")
	})

	// 5. OAuth2 Client Secret Resolution Error
	t.Run("OAuth2ClientSecretError", func(t *testing.T) {
		idSecret := &configv1.SecretValue{}
		idSecret.SetPlainText("client-id")

		secSecret := &configv1.SecretValue{}
		secSecret.SetEnvironmentVariable("MISSING_ENV_VAR")

		oauthConfig := &configv1.OAuth2Auth{}
		oauthConfig.SetClientId(idSecret)
		oauthConfig.SetClientSecret(secSecret)

		auth := &configv1.Authentication{}
		auth.SetOauth2(oauthConfig)

		remoteContent := &configv1.RemoteContent{}
		remoteContent.SetHttpUrl("http://example.com")
		remoteContent.SetAuth(auth)

		secret := &configv1.SecretValue{}
		secret.SetRemoteContent(remoteContent)

		_, err := util.ResolveSecret(context.Background(), secret)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to resolve client secret")
	})
}
