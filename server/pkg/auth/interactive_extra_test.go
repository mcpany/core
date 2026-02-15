package auth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/storage/memory"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func TestHandleOAuthCallbackExtra(t *testing.T) {
	// Setup mock oauth2 server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/token" {
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{"access_token": "mock_access_token", "refresh_token": "mock_refresh_token", "token_type": "Bearer", "expires_in": 3600, "scope": "read"}`))
			return
		}
		http.NotFound(w, r)
	}))
	defer server.Close()

	store := memory.NewStore()
	manager := NewManager()
	manager.SetStorage(store)
	ctx := context.Background()

	// Test case 1: Successful OAuth callback with Credential
	t.Run("SuccessWithCredential", func(t *testing.T) {
		cred := configv1.Credential_builder{
			Id: proto.String("test-cred"),
			Authentication: configv1.Authentication_builder{
				Oauth2: configv1.OAuth2Auth_builder{
					ClientId:     configv1.SecretValue_builder{PlainText: proto.String("client-id")}.Build(),
					ClientSecret: configv1.SecretValue_builder{PlainText: proto.String("client-secret")}.Build(),
					TokenUrl:     proto.String(server.URL + "/token"),
					Scopes:       proto.String("read"),
				}.Build(),
			}.Build(),
		}.Build()
		require.NoError(t, store.SaveCredential(ctx, cred))

		err := manager.HandleOAuthCallback(ctx, "user1", "", "test-cred", "auth-code", "http://127.0.0.1/cb")
		require.NoError(t, err)

		updatedCred, err := store.GetCredential(ctx, "test-cred")
		require.NoError(t, err)
		assert.NotNil(t, updatedCred.GetToken())
		assert.Equal(t, "mock_access_token", updatedCred.GetToken().GetAccessToken())
		assert.Equal(t, "read", updatedCred.GetToken().GetScope())
	})

	// Test case 2: Successful OAuth callback with Service (UserToken storage)
	t.Run("SuccessWithService", func(t *testing.T) {
		svc := configv1.UpstreamServiceConfig_builder{
			Name: proto.String("test-service"),
			UpstreamAuth: configv1.Authentication_builder{
				Oauth2: configv1.OAuth2Auth_builder{
					ClientId:     configv1.SecretValue_builder{PlainText: proto.String("client-id")}.Build(),
					ClientSecret: configv1.SecretValue_builder{PlainText: proto.String("client-secret")}.Build(),
					TokenUrl:     proto.String(server.URL + "/token"),
				}.Build(),
			}.Build(),
		}.Build()
		require.NoError(t, store.SaveService(ctx, svc))

		err := manager.HandleOAuthCallback(ctx, "user1", "test-service", "", "auth-code", "http://127.0.0.1/cb")
		require.NoError(t, err)

		token, err := store.GetToken(ctx, "user1", "test-service")
		require.NoError(t, err)
		assert.NotNil(t, token)
		assert.Equal(t, "mock_access_token", token.GetAccessToken())
	})

	// Test case 4: Credential not found
	t.Run("CredentialNotFound", func(t *testing.T) {
		err := manager.HandleOAuthCallback(ctx, "u", "", "missing-cred", "code", "url")
		assert.ErrorContains(t, err, "credential missing-cred not found")
	})

	// Test case 5: Credential not configured for OAuth2
	t.Run("CredentialNoOAuth", func(t *testing.T) {
		cred := configv1.Credential_builder{
			Id: proto.String("api-key-cred"),
			Authentication: configv1.Authentication_builder{
				ApiKey: configv1.APIKeyAuth_builder{
					Value: configv1.SecretValue_builder{PlainText: proto.String("key")}.Build(),
				}.Build(),
			}.Build(),
		}.Build()
		require.NoError(t, store.SaveCredential(ctx, cred))
		err := manager.HandleOAuthCallback(ctx, "u", "", "api-key-cred", "code", "url")
		assert.ErrorContains(t, err, "not configured for OAuth2")
	})

	// Test case 6: Service not found
	t.Run("ServiceNotFound", func(t *testing.T) {
		err := manager.HandleOAuthCallback(ctx, "u", "missing-svc", "", "code", "url")
		assert.ErrorContains(t, err, "service missing-svc not found")
	})

	// Test case 7: Service has no upstream auth
	t.Run("ServiceNoAuth", func(t *testing.T) {
		svc := configv1.UpstreamServiceConfig_builder{
			Name: proto.String("no-auth-svc"),
		}.Build()
		require.NoError(t, store.SaveService(ctx, svc))
		err := manager.HandleOAuthCallback(ctx, "u", "no-auth-svc", "", "code", "url")
		assert.ErrorContains(t, err, "no upstream auth configuration")
	})

	// Test case 8: Service not OAuth2
	t.Run("ServiceNotOAuth", func(t *testing.T) {
		svc := configv1.UpstreamServiceConfig_builder{
			Name: proto.String("apikey-svc"),
			UpstreamAuth: configv1.Authentication_builder{
				ApiKey: configv1.APIKeyAuth_builder{}.Build(),
			}.Build(),
		}.Build()
		require.NoError(t, store.SaveService(ctx, svc))
		err := manager.HandleOAuthCallback(ctx, "u", "apikey-svc", "", "code", "url")
		assert.ErrorContains(t, err, "not configured for OAuth2")
	})

	// Test case 9: Missing IDs
	t.Run("MissingIDs", func(t *testing.T) {
		err := manager.HandleOAuthCallback(ctx, "u", "", "", "code", "url")
		assert.ErrorContains(t, err, "either service_id or credential_id must be provided")
	})

	// Test case 10: Exchange failure
	t.Run("ExchangeFailure", func(t *testing.T) {
		cred := configv1.Credential_builder{
			Id: proto.String("fail-cred"),
			Authentication: configv1.Authentication_builder{
				Oauth2: configv1.OAuth2Auth_builder{
					ClientId:     configv1.SecretValue_builder{PlainText: proto.String("client-id")}.Build(),
					ClientSecret: configv1.SecretValue_builder{PlainText: proto.String("client-secret")}.Build(),
					TokenUrl:     proto.String(server.URL + "/404"), // Invalid endpoint
				}.Build(),
			}.Build(),
		}.Build()
		require.NoError(t, store.SaveCredential(ctx, cred))
		err := manager.HandleOAuthCallback(ctx, "u", "", "fail-cred", "code", "url")
		assert.ErrorContains(t, err, "failed to exchange code")
	})
}

func TestInitiateOAuthExtra(t *testing.T) {
	store := memory.NewStore()
	manager := NewManager()
	manager.SetStorage(store)
	ctx := context.Background()

	t.Run("NoAuthUrl", func(t *testing.T) {
		cred := configv1.Credential_builder{
			Id: proto.String("bad-cred"),
			Authentication: configv1.Authentication_builder{
				Oauth2: configv1.OAuth2Auth_builder{
					ClientId: configv1.SecretValue_builder{PlainText: proto.String("client-id")}.Build(),
					// Missing Auth URL
					TokenUrl: proto.String("http://token.com"),
				}.Build(),
			}.Build(),
		}.Build()
		require.NoError(t, store.SaveCredential(ctx, cred))
		_, _, err := manager.InitiateOAuth(ctx, "u", "", "bad-cred", "cb")
		assert.ErrorContains(t, err, "authorization_url is required")
	})

	t.Run("OIDCNotImplemented", func(t *testing.T) {
		cred := configv1.Credential_builder{
			Id: proto.String("oidc-cred"),
			Authentication: configv1.Authentication_builder{
				Oauth2: configv1.OAuth2Auth_builder{
					IssuerUrl: proto.String("http://issuer.com"),
				}.Build(),
			}.Build(),
		}.Build()
		require.NoError(t, store.SaveCredential(ctx, cred))
		_, _, err := manager.InitiateOAuth(ctx, "u", "", "oidc-cred", "cb")
		assert.ErrorContains(t, err, "OIDC discovery not implemented")
	})
}

// Helper for testing expiration
func tokenWithExpiry(d time.Duration) *configv1.UserToken {
	return configv1.UserToken_builder{
		AccessToken: proto.String("access"),
		Expiry:      proto.String(time.Now().Add(d).Format(time.RFC3339)),
	}.Build()
}

func TestOAuth2AuthenticateExtra(t *testing.T) {
	// We need to test OAuth2Authenticator.Authenticate
	// It's in oauth.go

	// Since we can't easily mock oauth2.Config.TokenSource (it's internal to Authenticate usually),
	// we will rely on integration tests or mocking net/http.

	// However, `Authenticate` method in `oauth.go` checks if token is valid.

	t.Run("InvalidToken", func(t *testing.T) {
		// Just test missing Authorization header logic
		auth := &OAuth2Authenticator{}
		req := httptest.NewRequest("GET", "/", nil)
		_, err := auth.Authenticate(context.Background(), req)
		assert.ErrorContains(t, err, "unauthorized")
	})

	t.Run("InvalidHeaderFormat", func(t *testing.T) {
		auth := &OAuth2Authenticator{}
		req := httptest.NewRequest("GET", "/", nil)
		req.Header.Set("Authorization", "InvalidFormat")
		_, err := auth.Authenticate(context.Background(), req)
		assert.ErrorContains(t, err, "unauthorized")
	})

	t.Run("InvalidScheme", func(t *testing.T) {
		auth := &OAuth2Authenticator{}
		req := httptest.NewRequest("GET", "/", nil)
		req.Header.Set("Authorization", "Basic user:pass")
		_, err := auth.Authenticate(context.Background(), req)
		assert.ErrorContains(t, err, "unauthorized")
	})
}
