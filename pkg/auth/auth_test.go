/*
 * Copyright 2025 Author(s) of MCP-XY
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
	"net/http/httptest"
	"testing"

	configv1 "github.com/mcpxy/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAPIKeyAuthenticator(t *testing.T) {
	config := &configv1.APIKeyAuth{}
	config.SetParamName("X-API-Key")
	config.SetKeyValue("secret-key")

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
	authManager := NewAuthManager()
	require.NotNil(t, authManager)

	config := &configv1.APIKeyAuth{}
	config.SetParamName("X-API-Key")
	config.SetKeyValue("secret-key")
	apiKeyAuth := NewAPIKeyAuthenticator(config)

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
}
