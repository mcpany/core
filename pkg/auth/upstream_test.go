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
	"net/http"
	"os"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func TestNewUpstreamAuthenticator(t *testing.T) {
	t.Run("NilConfig", func(t *testing.T) {
		auth, err := NewUpstreamAuthenticator(nil)
		assert.NoError(t, err)
		assert.Nil(t, auth)
	})

	t.Run("APIKey", func(t *testing.T) {
		config := (&configv1.UpstreamAuthentication_builder{
			ApiKey: (&configv1.UpstreamAPIKeyAuth_builder{
				HeaderName: proto.String("X-API-Key"),
				ApiKey:     &configv1.SecretValue{Value: &configv1.SecretValue_PlainText{PlainText: "test-key"}},
			}).Build(),
		}).Build()
		auth, err := NewUpstreamAuthenticator(config)
		require.NoError(t, err)
		require.NotNil(t, auth)

		req, _ := http.NewRequest("GET", "/", nil)
		err = auth.Authenticate(req)
		assert.NoError(t, err)
		assert.Equal(t, "test-key", req.Header.Get("X-API-Key"))
	})

	t.Run("BearerToken", func(t *testing.T) {
		config := (&configv1.UpstreamAuthentication_builder{
			BearerToken: (&configv1.UpstreamBearerTokenAuth_builder{
				Token: &configv1.SecretValue{Value: &configv1.SecretValue_PlainText{PlainText: "test-token"}},
			}).Build(),
		}).Build()
		auth, err := NewUpstreamAuthenticator(config)
		require.NoError(t, err)
		require.NotNil(t, auth)

		req, _ := http.NewRequest("GET", "/", nil)
		err = auth.Authenticate(req)
		assert.NoError(t, err)
		assert.Equal(t, "Bearer test-token", req.Header.Get("Authorization"))
	})

	t.Run("BasicAuth", func(t *testing.T) {
		config := (&configv1.UpstreamAuthentication_builder{
			BasicAuth: (&configv1.UpstreamBasicAuth_builder{
				Username: &configv1.SecretValue{Value: &configv1.SecretValue_PlainText{PlainText: "user"}},
				Password: &configv1.SecretValue{Value: &configv1.SecretValue_PlainText{PlainText: "pass"}},
			}).Build(),
		}).Build()
		auth, err := NewUpstreamAuthenticator(config)
		require.NoError(t, err)
		require.NotNil(t, auth)

		req, _ := http.NewRequest("GET", "/", nil)
		err = auth.Authenticate(req)
		assert.NoError(t, err)
		user, pass, ok := req.BasicAuth()
		assert.True(t, ok)
		assert.Equal(t, "user", user)
		assert.Equal(t, "pass", pass)
	})

	t.Run("NoAuthMethod", func(t *testing.T) {
		config := &configv1.UpstreamAuthentication{}
		auth, err := NewUpstreamAuthenticator(config)
		assert.NoError(t, err)
		assert.Nil(t, auth)
	})
}

func TestAPIKeyAuth_Authenticate(t *testing.T) {
	auth := &APIKeyAuth{
		HeaderName:  "X-Custom-Auth",
		HeaderValue: "secret-key",
	}
	req, _ := http.NewRequest("GET", "/", nil)
	err := auth.Authenticate(req)
	assert.NoError(t, err)
	assert.Equal(t, "secret-key", req.Header.Get("X-Custom-Auth"))
}

func TestBearerTokenAuth_Authenticate(t *testing.T) {
	auth := &BearerTokenAuth{
		Token: "secret-token",
	}
	req, _ := http.NewRequest("GET", "/", nil)
	err := auth.Authenticate(req)
	assert.NoError(t, err)
	assert.Equal(t, "Bearer secret-token", req.Header.Get("Authorization"))
}

func TestBasicAuth_Authenticate(t *testing.T) {
	auth := &BasicAuth{
		Username: "testuser",
		Password: "testpassword",
	}
	req, _ := http.NewRequest("GET", "/", nil)
	err := auth.Authenticate(req)
	assert.NoError(t, err)
	user, pass, ok := req.BasicAuth()
	assert.True(t, ok)
	assert.Equal(t, "testuser", user)
	assert.Equal(t, "testpassword", pass)
}

func TestSubstituteEnvVars(t *testing.T) {
	os.Setenv("TEST_API_KEY", "test-key")
	os.Setenv("TEST_BEARER_TOKEN", "test-token")
	os.Setenv("TEST_USERNAME", "test-user")
	os.Setenv("TEST_PASSWORD", "test-password")

	t.Run("APIKeyAuth", func(t *testing.T) {
		authConfig := (&configv1.UpstreamAuthentication_builder{
			ApiKey: (&configv1.UpstreamAPIKeyAuth_builder{
				HeaderName: proto.String("X-API-Key"),
				ApiKey:     &configv1.SecretValue{Value: &configv1.SecretValue_PlainText{PlainText: "{{TEST_API_KEY}}"}},
			}).Build(),
		}).Build()
		err := substituteEnvVars(authConfig)
		require.NoError(t, err)
		val, err := ResolveSecretValue(authConfig.GetApiKey().GetApiKey())
		require.NoError(t, err)
		require.Equal(t, "test-key", val)
	})

	t.Run("BearerTokenAuth", func(t *testing.T) {
		authConfig := (&configv1.UpstreamAuthentication_builder{
			BearerToken: (&configv1.UpstreamBearerTokenAuth_builder{
				Token: &configv1.SecretValue{Value: &configv1.SecretValue_PlainText{PlainText: "{{TEST_BEARER_TOKEN}}"}},
			}).Build(),
		}).Build()
		err := substituteEnvVars(authConfig)
		require.NoError(t, err)
		val, err := ResolveSecretValue(authConfig.GetBearerToken().GetToken())
		require.NoError(t, err)
		require.Equal(t, "test-token", val)
	})

	t.Run("BasicAuth", func(t *testing.T) {
		authConfig := (&configv1.UpstreamAuthentication_builder{
			BasicAuth: (&configv1.UpstreamBasicAuth_builder{
				Username: &configv1.SecretValue{Value: &configv1.SecretValue_PlainText{PlainText: "{{TEST_USERNAME}}"}},
				Password: &configv1.SecretValue{Value: &configv1.SecretValue_PlainText{PlainText: "{{TEST_PASSWORD}}"}},
			}).Build(),
		}).Build()
		err := substituteEnvVars(authConfig)
		require.NoError(t, err)
		val, err := ResolveSecretValue(authConfig.GetBasicAuth().GetUsername())
		require.NoError(t, err)
		require.Equal(t, "test-user", val)
		val, err = ResolveSecretValue(authConfig.GetBasicAuth().GetPassword())
		require.NoError(t, err)
		require.Equal(t, "test-password", val)
	})
}
