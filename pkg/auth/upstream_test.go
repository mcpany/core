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
		secret := (&configv1.SecretValue_builder{
			PlainText: proto.String("test-key"),
		}).Build()
		config := (&configv1.UpstreamAuthentication_builder{
			ApiKey: (&configv1.UpstreamAPIKeyAuth_builder{
				HeaderName: proto.String("X-API-Key"),
				ApiKey:     secret,
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
		secret := (&configv1.SecretValue_builder{
			PlainText: proto.String("test-token"),
		}).Build()
		config := (&configv1.UpstreamAuthentication_builder{
			BearerToken: (&configv1.UpstreamBearerTokenAuth_builder{
				Token: secret,
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
		secret := (&configv1.SecretValue_builder{
			PlainText: proto.String("pass"),
		}).Build()
		config := (&configv1.UpstreamAuthentication_builder{
			BasicAuth: (&configv1.UpstreamBasicAuth_builder{
				Username: proto.String("user"),
				Password: secret,
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

	t.Run("Validation", func(t *testing.T) {
		t.Run("APIKey_MissingHeader", func(t *testing.T) {
			secret := (&configv1.SecretValue_builder{
				PlainText: proto.String("test-key"),
			}).Build()
			config := (&configv1.UpstreamAuthentication_builder{
				ApiKey: (&configv1.UpstreamAPIKeyAuth_builder{
					ApiKey: secret,
				}).Build(),
			}).Build()
			_, err := NewUpstreamAuthenticator(config)
			assert.ErrorContains(t, err, "API key authentication requires a header name")
		})

		t.Run("APIKey_MissingKey", func(t *testing.T) {
			config := (&configv1.UpstreamAuthentication_builder{
				ApiKey: (&configv1.UpstreamAPIKeyAuth_builder{
					HeaderName: proto.String("X-API-Key"),
				}).Build(),
			}).Build()
			_, err := NewUpstreamAuthenticator(config)
			assert.ErrorContains(t, err, "API key authentication requires an API key")
		})

		t.Run("BearerToken_MissingToken", func(t *testing.T) {
			config := (&configv1.UpstreamAuthentication_builder{
				BearerToken: (&configv1.UpstreamBearerTokenAuth_builder{}).Build(),
			}).Build()
			_, err := NewUpstreamAuthenticator(config)
			assert.ErrorContains(t, err, "bearer token authentication requires a token")
		})

		t.Run("BasicAuth_MissingUsername", func(t *testing.T) {
			secret := (&configv1.SecretValue_builder{
				PlainText: proto.String("pass"),
			}).Build()
			config := (&configv1.UpstreamAuthentication_builder{
				BasicAuth: (&configv1.UpstreamBasicAuth_builder{
					Password: secret,
				}).Build(),
			}).Build()
			_, err := NewUpstreamAuthenticator(config)
			assert.ErrorContains(t, err, "basic authentication requires a username")
		})

		t.Run("BasicAuth_MissingPassword", func(t *testing.T) {
			config := (&configv1.UpstreamAuthentication_builder{
				BasicAuth: (&configv1.UpstreamBasicAuth_builder{
					Username: proto.String("user"),
				}).Build(),
			}).Build()
			_, err := NewUpstreamAuthenticator(config)
			assert.ErrorContains(t, err, "basic authentication requires a password")
		})

		t.Run("OAuth2_MissingClientID", func(t *testing.T) {
			clientSecret := (&configv1.SecretValue_builder{
				PlainText: proto.String("secret"),
			}).Build()
			config := (&configv1.UpstreamAuthentication_builder{
				Oauth2: (&configv1.UpstreamOAuth2Auth_builder{
					ClientSecret: clientSecret,
					TokenUrl:     proto.String("http://token.url"),
				}).Build(),
			}).Build()
			_, err := NewUpstreamAuthenticator(config)
			assert.ErrorContains(t, err, "OAuth2 authentication requires a client ID")
		})

		t.Run("OAuth2_MissingClientSecret", func(t *testing.T) {
			clientID := (&configv1.SecretValue_builder{
				PlainText: proto.String("id"),
			}).Build()
			config := (&configv1.UpstreamAuthentication_builder{
				Oauth2: (&configv1.UpstreamOAuth2Auth_builder{
					ClientId: clientID,
					TokenUrl: proto.String("http://token.url"),
				}).Build(),
			}).Build()
			_, err := NewUpstreamAuthenticator(config)
			assert.ErrorContains(t, err, "OAuth2 authentication requires a client secret")
		})

		t.Run("OAuth2_MissingTokenURL", func(t *testing.T) {
			clientID := (&configv1.SecretValue_builder{
				PlainText: proto.String("id"),
			}).Build()
			clientSecret := (&configv1.SecretValue_builder{
				PlainText: proto.String("secret"),
			}).Build()
			config := (&configv1.UpstreamAuthentication_builder{
				Oauth2: (&configv1.UpstreamOAuth2Auth_builder{
					ClientId:     clientID,
					ClientSecret: clientSecret,
				}).Build(),
			}).Build()
			_, err := NewUpstreamAuthenticator(config)
			assert.ErrorContains(t, err, "OAuth2 authentication requires a token URL")
		})
	})
}

func TestAPIKeyAuth_Authenticate(t *testing.T) {
	secret := (&configv1.SecretValue_builder{
		PlainText: proto.String("secret-key"),
	}).Build()
	auth := &APIKeyAuth{
		HeaderName:  "X-Custom-Auth",
		HeaderValue: secret,
	}
	req, _ := http.NewRequest("GET", "/", nil)
	err := auth.Authenticate(req)
	assert.NoError(t, err)
	assert.Equal(t, "secret-key", req.Header.Get("X-Custom-Auth"))
}

func TestBearerTokenAuth_Authenticate(t *testing.T) {
	secret := (&configv1.SecretValue_builder{
		PlainText: proto.String("secret-token"),
	}).Build()
	auth := &BearerTokenAuth{
		Token: secret,
	}
	req, _ := http.NewRequest("GET", "/", nil)
	err := auth.Authenticate(req)
	assert.NoError(t, err)
	assert.Equal(t, "Bearer secret-token", req.Header.Get("Authorization"))
}

func TestBasicAuth_Authenticate(t *testing.T) {
	secret := (&configv1.SecretValue_builder{
		PlainText: proto.String("testpassword"),
	}).Build()
	auth := &BasicAuth{
		Username: "testuser",
		Password: secret,
	}
	req, _ := http.NewRequest("GET", "/", nil)
	err := auth.Authenticate(req)
	assert.NoError(t, err)
	user, pass, ok := req.BasicAuth()
	assert.True(t, ok)
	assert.Equal(t, "testuser", user)
	assert.Equal(t, "testpassword", pass)
}
