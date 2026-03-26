// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package auth

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/util/passhash"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func TestNewAPIKeyAuthenticator(t *testing.T) {
	t.Run("nil_config", func(t *testing.T) {
		authenticator := NewAPIKeyAuthenticator(nil)
		assert.Nil(t, authenticator)
	})

	t.Run("invalid_nil_config", func(t *testing.T) {
		authenticator := NewAPIKeyAuthenticator(configv1.APIKeyAuth_builder{}.Build())
		assert.Nil(t, authenticator)
	})

	t.Run("empty_param_name", func(t *testing.T) {
		config := configv1.APIKeyAuth_builder{
			VerificationValue: proto.String("some-key"),
		}.Build()
		authenticator := NewAPIKeyAuthenticator(config)
		assert.Nil(t, authenticator)
	})

	t.Run("empty_key_value", func(t *testing.T) {
		config := configv1.APIKeyAuth_builder{
			ParamName: proto.String("X-API-Key"),
		}.Build()
		authenticator := NewAPIKeyAuthenticator(config)
		assert.Nil(t, authenticator)
	})
}

func TestAPIKeyAuthenticator(t *testing.T) {
	config := configv1.APIKeyAuth_builder{
		ParamName: proto.String("X-API-Key"),
		VerificationValue:  proto.String("secret-key"),
	}.Build()

	authenticator := NewAPIKeyAuthenticator(config)
	require.NotNil(t, authenticator)

	t.Run("successful_authentication", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		req.Header.Set("X-API-Key", "secret-key")
		_, err := authenticator.Authenticate(context.Background(), req)
		assert.NoError(t, err)
	})

	t.Run("failed_authentication_wrong_key", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		req.Header.Set("X-API-Key", "wrong-key")
		_, err := authenticator.Authenticate(context.Background(), req)
		assert.Error(t, err)
		assert.Equal(t, "unauthorized", err.Error())
	})

	t.Run("failed_authentication_missing_header", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		_, err := authenticator.Authenticate(context.Background(), req)
		assert.Error(t, err)
		assert.Equal(t, "unauthorized", err.Error())
	})
}

func TestAuthManager(t *testing.T) {
	authManager := NewManager()
	require.NotNil(t, authManager)

	config := configv1.APIKeyAuth_builder{
		ParamName: proto.String("X-API-Key"),
		VerificationValue:  proto.String("secret-key"),
	}.Build()
	apiKeyAuth := NewAPIKeyAuthenticator(config)

	serviceID := "test-service"
	_ = authManager.AddAuthenticator(serviceID, apiKeyAuth)

	t.Run("get_authenticator", func(t *testing.T) {
		authenticator, ok := authManager.GetAuthenticator(serviceID)
		assert.True(t, ok)
		assert.Equal(t, apiKeyAuth, authenticator)

		_, ok = authManager.GetAuthenticator("non-existent-service")
		assert.False(t, ok)
	})

	t.Run("authenticate_with_registered_service", func(t *testing.T) {
		// Successful authentication
		req := httptest.NewRequest("GET", "/", nil)
		req.Header.Set("X-API-Key", "secret-key")
		_, err := authManager.Authenticate(context.Background(), serviceID, req)
		assert.NoError(t, err)

		// Failed authentication
		req = httptest.NewRequest("GET", "/", nil)
		req.Header.Set("X-API-Key", "wrong-key")
		_, err = authManager.Authenticate(context.Background(), serviceID, req)
		assert.Error(t, err)
	})

	t.Run("authenticate_with_unregistered_service", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		// No headers, and no authenticator configured -> Fail Closed
		_, err := authManager.Authenticate(context.Background(), "unregistered-service", req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no authentication configured")
	})

	t.Run("add_nil_authenticator", func(t *testing.T) {
		err := authManager.AddAuthenticator("nil-service", nil)
		assert.Error(t, err)
	})

	t.Run("authenticate_with_nil_authenticator", func(t *testing.T) {
		// This test ensures that a nil authenticator doesn't cause a panic.
		// The AddAuthenticator function now prevents adding nil authenticators,
		// but this test is a safeguard.
		// authManager.authenticators.Store("nil-service", nil)
		req := httptest.NewRequest("GET", "/", nil)
		// Since xsync.Map does not allow storing nil values, this test is no longer valid.
		// The AddAuthenticator function already prevents adding nil authenticators.
		// We will just ensure that authenticating with a non-existent service does not panic.
		assert.NotPanics(t, func() {
			_, err := authManager.Authenticate(context.Background(), "nil-service", req)
			assert.Error(t, err)
		})
	})

	t.Run("remove_authenticator", func(t *testing.T) {
		// Add an authenticator to remove
		_ = authManager.AddAuthenticator("service-to-remove", apiKeyAuth)

		// Verify it was added
		_, ok := authManager.GetAuthenticator("service-to-remove")
		assert.True(t, ok)

		// Remove the authenticator
		authManager.RemoveAuthenticator("service-to-remove")

		// Verify it was removed
		_, ok = authManager.GetAuthenticator("service-to-remove")
		assert.False(t, ok)
	})

	t.Run("add_authenticator_with_empty_service_id", func(t *testing.T) {
		err := authManager.AddAuthenticator("", apiKeyAuth)
		assert.NoError(t, err)

		authenticator, ok := authManager.GetAuthenticator("")
		assert.True(t, ok)
		assert.Equal(t, apiKeyAuth, authenticator)
	})

	t.Run("authenticate_with_global_api_key", func(t *testing.T) {
		authManager.SetAPIKey("global-secret-key")

		// Successful authentication
		req := httptest.NewRequest("GET", "/", nil)
		req.Header.Set("X-API-Key", "global-secret-key")
		_, err := authManager.Authenticate(context.Background(), "any-service", req)
		assert.NoError(t, err)

		// Failed authentication
		req = httptest.NewRequest("GET", "/", nil)
		req.Header.Set("X-API-Key", "wrong-key")
		_, err = authManager.Authenticate(context.Background(), "any-service", req)
		assert.Error(t, err)

		// Failed authentication (missing key)
		req = httptest.NewRequest("GET", "/", nil)
		_, err = authManager.Authenticate(context.Background(), "any-service", req)
		assert.Error(t, err)

		authManager.SetAPIKey("")
	})

	t.Run("set_and_get_users", func(t *testing.T) {
		users := []*configv1.User{
			configv1.User_builder{
				Id:    proto.String("user1"),
				Roles: []string{"admin"},
			}.Build(),
			configv1.User_builder{
				Id:    proto.String("user2"),
				Roles: []string{"user"},
			}.Build(),
		}
		authManager.SetUsers(users)

		u1, ok := authManager.GetUser("user1")
		assert.True(t, ok)
		assert.Equal(t, "user1", u1.GetId())
		assert.Equal(t, []string{"admin"}, u1.GetRoles())

		u2, ok := authManager.GetUser("user2")
		assert.True(t, ok)
		assert.Equal(t, "user2", u2.GetId())

		_, ok = authManager.GetUser("non-existent")
		assert.False(t, ok)
	})

	t.Run("set_storage", func(t *testing.T) {
		// Just ensure it doesn't panic and sets the storage (internal field)
		// Since we can't easily check private field, we trust the setter.
		// Or we could mock storage and check if it's used in other methods like InitiateOAuth
		// But those are tested in interactive_test.go
		authManager.SetStorage(nil)
	})
}

func TestContextHelpers(t *testing.T) {
	ctx := context.Background()

	t.Run("user_context", func(t *testing.T) {
		userID := "user-123"
		ctxWithUser := ContextWithUser(ctx, userID)
		val, ok := UserFromContext(ctxWithUser)
		assert.True(t, ok)
		assert.Equal(t, userID, val)

		_, ok = UserFromContext(ctx)
		assert.False(t, ok)
	})

	t.Run("profile_id_context", func(t *testing.T) {
		profileID := "profile-abc"
		ctxWithProfile := ContextWithProfileID(ctx, profileID)
		val, ok := ProfileIDFromContext(ctxWithProfile)
		assert.True(t, ok)
		assert.Equal(t, profileID, val)

		_, ok = ProfileIDFromContext(ctx)
		assert.False(t, ok)
	})

	t.Run("api_key_context", func(t *testing.T) {
		apiKey := "key-xyz"
		ctxWithKey := ContextWithAPIKey(ctx, apiKey)
		val, ok := APIKeyFromContext(ctxWithKey)
		assert.True(t, ok)
		assert.Equal(t, apiKey, val)

		_, ok = APIKeyFromContext(ctx)
		assert.False(t, ok)
	})
}

func TestAddOAuth2Authenticator(t *testing.T) {
	authManager := NewManager()
	require.NotNil(t, authManager)

	// Mock OIDC provider
	var server *httptest.Server
	const wellKnownPath = "/.well-known/openid-configuration"
	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == wellKnownPath {
			w.Header().Set("Content-Type", "application/json")
			_, _ = fmt.Fprintln(w, `{"issuer": "`+server.URL+`", "jwks_uri": "`+server.URL+`/jwks"}`)
		}
	}))
	defer server.Close()

	config := &OAuth2Config{
		IssuerURL: server.URL,
	}

	serviceID := "test-service"

	t.Run("successful_add", func(t *testing.T) {
		err := authManager.AddOAuth2Authenticator(context.Background(), serviceID, config)
		assert.NoError(t, err)

		authenticator, ok := authManager.GetAuthenticator(serviceID)
		assert.True(t, ok)
		assert.NotNil(t, authenticator)
	})

	t.Run("nil_config", func(t *testing.T) {
		err := authManager.AddOAuth2Authenticator(context.Background(), "another-service", nil)
		assert.NoError(t, err)

		_, ok := authManager.GetAuthenticator("another-service")
		assert.False(t, ok)
	})

	t.Run("failed_creation", func(t *testing.T) {
		config := &OAuth2Config{
			IssuerURL: "http://invalid-url",
		}
		err := authManager.AddOAuth2Authenticator(context.Background(), "failed-service", config)
		assert.Error(t, err)
	})
}

func TestAPIKeyAuthenticator_Query(t *testing.T) {
	config := configv1.APIKeyAuth_builder{
		ParamName: proto.String("api_key"),
		VerificationValue:  proto.String("secret"),
		In:        configv1.APIKeyAuth_QUERY.Enum(),
	}.Build()

	authenticator := NewAPIKeyAuthenticator(config)
	require.NotNil(t, authenticator)

	t.Run("successful_query_auth", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/?api_key=secret", nil)
		_, err := authenticator.Authenticate(context.Background(), req)
		assert.NoError(t, err)
	})

	t.Run("failed_query_auth", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/?api_key=wrong", nil)
		_, err := authenticator.Authenticate(context.Background(), req)
		assert.Error(t, err)
	})

		t.Run("authentication_with_configured_username", func(t *testing.T) {
			password := "secret123"
			hashed, _ := passhash.Password(password)
			configWithUser := configv1.BasicAuth_builder{
				Username:     proto.String("admin"),
				PasswordHash: proto.String(hashed),
			}.Build()
			authWithUser := NewBasicAuthenticator(configWithUser)
			require.NotNil(t, authWithUser)

			t.Run("correct_username", func(t *testing.T) {
				req := httptest.NewRequest("GET", "/", nil)
				req.SetBasicAuth("admin", password)
				_, err := authWithUser.Authenticate(context.Background(), req)
				assert.NoError(t, err)
			})

		t.Run("wrong_username", func(t *testing.T) {
			req := httptest.NewRequest("GET", "/", nil)
			req.SetBasicAuth("guest", password)
			_, err := authWithUser.Authenticate(context.Background(), req)
			assert.Error(t, err)
		})
	})
}

func TestValidateAuthentication(t *testing.T) {
	t.Run("nil_config", func(t *testing.T) {
		err := ValidateAuthentication(context.Background(), nil, nil)
		assert.NoError(t, err)
	})

	t.Run("api_key_valid", func(t *testing.T) {
		config := configv1.Authentication_builder{
			ApiKey: configv1.APIKeyAuth_builder{
				ParamName:         proto.String("X-API-Key"),
				VerificationValue: proto.String("secret"),
				In:                configv1.APIKeyAuth_HEADER.Enum(),
			}.Build(),
		}.Build()
		req := httptest.NewRequest("GET", "/", nil)
		req.Header.Set("X-API-Key", "secret")
		err := ValidateAuthentication(context.Background(), config, req)
		assert.NoError(t, err)
	})

	t.Run("api_key_invalid", func(t *testing.T) {
		config := configv1.Authentication_builder{
			ApiKey: configv1.APIKeyAuth_builder{
				ParamName:         proto.String("X-API-Key"),
				VerificationValue: proto.String("secret"),
				In:                configv1.APIKeyAuth_HEADER.Enum(),
			}.Build(),
		}.Build()
		req := httptest.NewRequest("GET", "/", nil)
		req.Header.Set("X-API-Key", "wrong")
		err := ValidateAuthentication(context.Background(), config, req)
		assert.Error(t, err)
	})

	t.Run("api_key_bad_config", func(t *testing.T) {
		config := configv1.Authentication_builder{
			ApiKey: configv1.APIKeyAuth_builder{
				// Missing params
			}.Build(),
		}.Build()
		req := httptest.NewRequest("GET", "/", nil)
		err := ValidateAuthentication(context.Background(), config, req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid API key configuration")
	})

	t.Run("oauth2_valid_config_missing_token", func(t *testing.T) {
		// Mock OIDC provider
		var server *httptest.Server
		const wellKnownPath = "/.well-known/openid-configuration"
		server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == wellKnownPath {
				w.Header().Set("Content-Type", "application/json")
				_, _ = fmt.Fprintln(w, `{"issuer": "`+server.URL+`", "jwks_uri": "`+server.URL+`/jwks"}`)
			} else if r.URL.Path == "/jwks" {
				w.Header().Set("Content-Type", "application/json")
				_, _ = fmt.Fprintln(w, `{"keys": []}`)
			}
		}))
		defer server.Close()

		config := configv1.Authentication_builder{
			Oauth2: configv1.OAuth2Auth_builder{
				IssuerUrl: proto.String(server.URL),
				Audience:  proto.String("test-audience"),
			}.Build(),
		}.Build()
		req := httptest.NewRequest("GET", "/", nil)
		err := ValidateAuthentication(context.Background(), config, req)
		assert.Error(t, err)
		assert.Equal(t, "unauthorized", err.Error())
	})

	t.Run("oidc_method", func(t *testing.T) {
		// Mock OIDC provider
		var server *httptest.Server
		const wellKnownPath = "/.well-known/openid-configuration"
		server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == wellKnownPath {
				w.Header().Set("Content-Type", "application/json")
				_, _ = fmt.Fprintln(w, `{"issuer": "`+server.URL+`", "jwks_uri": "`+server.URL+`/jwks"}`)
			} else if r.URL.Path == "/jwks" {
				w.Header().Set("Content-Type", "application/json")
				_, _ = fmt.Fprintln(w, `{"keys": []}`)
			}
		}))
		defer server.Close()

		config := configv1.Authentication_builder{
			Oidc: configv1.OIDCAuth_builder{
				Issuer:   proto.String(server.URL),
				Audience: []string{"test-audience"},
			}.Build(),
		}.Build()
		req := httptest.NewRequest("GET", "/", nil)
		// Should fail unauthorized because no token, but prove it tried to authenticate using OIDC config
		err := ValidateAuthentication(context.Background(), config, req)
		assert.Error(t, err)
		assert.Equal(t, "unauthorized", err.Error())
	})

	t.Run("oidc_method_missing_issuer", func(t *testing.T) {
		config := configv1.Authentication_builder{
			Oidc: configv1.OIDCAuth_builder{
				// Missing issuer
			}.Build(),
		}.Build()
		req := httptest.NewRequest("GET", "/", nil)
		err := ValidateAuthentication(context.Background(), config, req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid OIDC configuration")
	})

	t.Run("oauth2_method_missing_issuer", func(t *testing.T) {
		config := configv1.Authentication_builder{
			Oauth2: configv1.OAuth2Auth_builder{
				// Missing issuer
			}.Build(),
		}.Build()
		req := httptest.NewRequest("GET", "/", nil)
		err := ValidateAuthentication(context.Background(), config, req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid OAuth2 configuration")
	})

	t.Run("no_method", func(t *testing.T) {
		config := &configv1.Authentication{}
		err := ValidateAuthentication(context.Background(), config, nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unsupported or missing authentication method")
	})
}

func TestBasicAuthenticator(t *testing.T) {
	password := "secret123"
	hashed, _ := passhash.Password(password)
	config := configv1.BasicAuth_builder{
		PasswordHash: proto.String(hashed),
	}.Build()

	authenticator := NewBasicAuthenticator(config)
	require.NotNil(t, authenticator)

	t.Run("successful_authentication", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		req.SetBasicAuth("user", password)
		_, err := authenticator.Authenticate(context.Background(), req)
		assert.NoError(t, err)
	})

	t.Run("failed_authentication_wrong_password", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		req.SetBasicAuth("user", "wrong")
		_, err := authenticator.Authenticate(context.Background(), req)
		assert.Error(t, err)
	})

	t.Run("failed_authentication_missing_header", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		_, err := authenticator.Authenticate(context.Background(), req)
		assert.Error(t, err)
	})
}

func TestTrustedHeaderAuthenticator(t *testing.T) {
	config := configv1.TrustedHeaderAuth_builder{
		HeaderName:  proto.String("X-Trusted-User"),
		HeaderValue: proto.String("verified"),
	}.Build()
	authenticator := NewTrustedHeaderAuthenticator(config)
	require.NotNil(t, authenticator)

	t.Run("successful_authentication", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		req.Header.Set("X-Trusted-User", "verified")
		_, err := authenticator.Authenticate(context.Background(), req)
		assert.NoError(t, err)
	})

	t.Run("failed_authentication_wrong_value", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		req.Header.Set("X-Trusted-User", "unverified")
		_, err := authenticator.Authenticate(context.Background(), req)
		assert.Error(t, err)
	})

	t.Run("failed_authentication_missing_header", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		_, err := authenticator.Authenticate(context.Background(), req)
		assert.Error(t, err)
	})
}
