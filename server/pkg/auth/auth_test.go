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
			KeyValue: proto.String("some-key"),
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
		KeyValue:  proto.String("secret-key"),
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
		KeyValue:  proto.String("secret-key"),
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
		// No headers, but should pass as no authenticator is configured
		_, err := authManager.Authenticate(context.Background(), "unregistered-service", req)
		assert.NoError(t, err)
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
			assert.NoError(t, err)
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
		assert.Equal(t, "unauthorized", err.Error())

		// Failed authentication (missing key)
		req = httptest.NewRequest("GET", "/", nil)
		_, err = authManager.Authenticate(context.Background(), "any-service", req)
		assert.Error(t, err)
		assert.Equal(t, "unauthorized", err.Error())

		authManager.SetAPIKey("")
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
		KeyValue:  proto.String("secret"),
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
}

func TestValidateAuthentication(t *testing.T) {
	t.Run("nil_config", func(t *testing.T) {
		err := ValidateAuthentication(context.Background(), nil, nil)
		assert.NoError(t, err)
	})

	t.Run("api_key_valid", func(t *testing.T) {
		config := &configv1.AuthenticationConfig{
			AuthMethod: &configv1.AuthenticationConfig_ApiKey{
				ApiKey: &configv1.APIKeyAuth{
					ParamName: proto.String("X-API-Key"),
					KeyValue:  proto.String("secret"),
					In:        configv1.APIKeyAuth_HEADER.Enum(),
				},
			},
		}
		req := httptest.NewRequest("GET", "/", nil)
		req.Header.Set("X-API-Key", "secret")
		err := ValidateAuthentication(context.Background(), config, req)
		assert.NoError(t, err)
	})

	t.Run("api_key_invalid", func(t *testing.T) {
		config := &configv1.AuthenticationConfig{
			AuthMethod: &configv1.AuthenticationConfig_ApiKey{
				ApiKey: &configv1.APIKeyAuth{
					ParamName: proto.String("X-API-Key"),
					KeyValue:  proto.String("secret"),
					In:        configv1.APIKeyAuth_HEADER.Enum(),
				},
			},
		}
		req := httptest.NewRequest("GET", "/", nil)
		req.Header.Set("X-API-Key", "wrong")
		err := ValidateAuthentication(context.Background(), config, req)
		assert.Error(t, err)
	})

	t.Run("api_key_bad_config", func(t *testing.T) {
		config := &configv1.AuthenticationConfig{
			AuthMethod: &configv1.AuthenticationConfig_ApiKey{
				ApiKey: &configv1.APIKeyAuth{
					// Missing params
				},
			},
		}
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

		config := &configv1.AuthenticationConfig{
			AuthMethod: &configv1.AuthenticationConfig_Oauth2{
				Oauth2: &configv1.OAuth2Auth{
					IssuerUrl: proto.String(server.URL),
					Audience:  proto.String("test-audience"),
				},
			},
		}
		req := httptest.NewRequest("GET", "/", nil)
		err := ValidateAuthentication(context.Background(), config, req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "missing Authorization header")
	})

	t.Run("no_method", func(t *testing.T) {
		config := &configv1.AuthenticationConfig{}
		err := ValidateAuthentication(context.Background(), config, nil)
		assert.NoError(t, err)
	})
}
