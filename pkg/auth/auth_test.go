/*
 * Copyright 2025 Author(s) of MCP Any
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package auth

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewAPIKeyAuthenticator(t *testing.T) {
	t.Run("valid_config", func(t *testing.T) {
		authenticator := NewAPIKeyAuthenticator("X-API-Key", "secret-key")
		assert.NotNil(t, authenticator)
	})
}

func TestAPIKeyAuthenticator(t *testing.T) {
	authenticator := NewAPIKeyAuthenticator("X-API-Key", "secret-key")
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
	})

	t.Run("failed_authentication_missing_header", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		_, err := authenticator.Authenticate(context.Background(), req)
		assert.Error(t, err)
	})
}

func TestAuthManager(t *testing.T) {
	authManager := NewAuthManager()
	require.NotNil(t, authManager)

	apiKeyAuth := NewAPIKeyAuthenticator("X-API-Key", "secret-key")

	serviceID := "test-service"
	authManager.AddAuthenticator(serviceID, apiKeyAuth)

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

	t.Run("remove_authenticator", func(t *testing.T) {
		// Add an authenticator to remove
		authManager.AddAuthenticator("service-to-remove", apiKeyAuth)

		// Verify it was added
		_, ok := authManager.GetAuthenticator("service-to-remove")
		assert.True(t, ok)

		// Remove the authenticator
		authManager.RemoveAuthenticator("service-to-remove")

		// Verify it was removed
		_, ok = authManager.GetAuthenticator("service-to-remove")
		assert.False(t, ok)
	})
}

func TestAddOAuth2Authenticator(t *testing.T) {
	authManager := NewAuthManager()
	require.NotNil(t, authManager)

	// Mock OIDC provider
	var server *httptest.Server
	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/.well-known/openid-configuration" {
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprintln(w, `{"issuer": "`+server.URL+`", "jwks_uri": "`+server.URL+`/jwks"}`)
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
