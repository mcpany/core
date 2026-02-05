// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package auth

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"
	"google.golang.org/protobuf/proto"
)

func TestNewUpstreamAuthenticator(t *testing.T) {
	t.Run("NilConfig", func(t *testing.T) {
		auth, err := NewUpstreamAuthenticator(nil)
		assert.NoError(t, err)
		assert.Nil(t, auth)
	})

	t.Run("OAuth2", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(oauth2.Token{
				AccessToken: "test-token",
				TokenType:   "Bearer",
				Expiry:      time.Now().Add(time.Hour),
			})
		}))
		defer ts.Close()

		clientID := configv1.SecretValue_builder{
			PlainText: proto.String("id"),
		}.Build()
		clientSecret := configv1.SecretValue_builder{
			PlainText: proto.String("secret"),
		}.Build()
		config := configv1.Authentication_builder{
			Oauth2: configv1.OAuth2Auth_builder{
				ClientId:     clientID,
				ClientSecret: clientSecret,
				TokenUrl:     proto.String(ts.URL),
			}.Build(),
		}.Build()
		auth, err := NewUpstreamAuthenticator(config)
		require.NoError(t, err)
		require.NotNil(t, auth)

		req, _ := http.NewRequest("GET", "/", nil)
		err = auth.Authenticate(req)
		assert.NoError(t, err)
		assert.Equal(t, "Bearer test-token", req.Header.Get("Authorization"))
	})

	t.Run("APIKey", func(t *testing.T) {
		secret := configv1.SecretValue_builder{
			PlainText: proto.String("test-key"),
		}.Build()
		config := configv1.Authentication_builder{
			ApiKey: configv1.APIKeyAuth_builder{
				ParamName: proto.String("X-API-Key"),
				Value:     secret,
			}.Build(),
		}.Build()
		auth, err := NewUpstreamAuthenticator(config)
		require.NoError(t, err)
		require.NotNil(t, auth)

		req, _ := http.NewRequest("GET", "/", nil)
		err = auth.Authenticate(req)
		assert.NoError(t, err)
		assert.Equal(t, "test-key", req.Header.Get("X-API-Key"))
	})

	t.Run("BearerToken", func(t *testing.T) {
		secret := configv1.SecretValue_builder{
			PlainText: proto.String("test-token"),
		}.Build()
		config := configv1.Authentication_builder{
			BearerToken: configv1.BearerTokenAuth_builder{
				Token: secret,
			}.Build(),
		}.Build()
		auth, err := NewUpstreamAuthenticator(config)
		require.NoError(t, err)
		require.NotNil(t, auth)

		req, _ := http.NewRequest("GET", "/", nil)
		err = auth.Authenticate(req)
		assert.NoError(t, err)
		assert.Equal(t, "Bearer test-token", req.Header.Get("Authorization"))
	})

	t.Run("BasicAuth", func(t *testing.T) {
		secret := configv1.SecretValue_builder{
			PlainText: proto.String("pass"),
		}.Build()
		config := configv1.Authentication_builder{
			BasicAuth: configv1.BasicAuth_builder{
				Username: proto.String("user"),
				Password: secret,
			}.Build(),
		}.Build()
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
		config := configv1.Authentication_builder{}.Build()
		auth, err := NewUpstreamAuthenticator(config)
		assert.NoError(t, err)
		assert.Nil(t, auth)
	})

	t.Run("Validation", func(t *testing.T) {
		t.Run("APIKey_MissingHeader", func(t *testing.T) {
			secret := configv1.SecretValue_builder{
				PlainText: proto.String("test-key"),
			}.Build()
			config := configv1.Authentication_builder{
				ApiKey: configv1.APIKeyAuth_builder{
					Value: secret,
				}.Build(),
			}.Build()
			_, err := NewUpstreamAuthenticator(config)
			assert.ErrorContains(t, err, "API key authentication requires a parameter name")
		})

		t.Run("APIKey_MissingKey", func(t *testing.T) {
			config := configv1.Authentication_builder{
				ApiKey: configv1.APIKeyAuth_builder{
					ParamName: proto.String("X-API-Key"),
				}.Build(),
			}.Build()
			_, err := NewUpstreamAuthenticator(config)
			assert.ErrorContains(t, err, "API key authentication requires an API key")
		})

		t.Run("BearerToken_MissingToken", func(t *testing.T) {
			config := configv1.Authentication_builder{
				BearerToken: configv1.BearerTokenAuth_builder{}.Build(),
			}.Build()
			_, err := NewUpstreamAuthenticator(config)
			assert.ErrorContains(t, err, "bearer token authentication requires a token")
		})

		t.Run("BasicAuth_MissingUsername", func(t *testing.T) {
			secret := configv1.SecretValue_builder{
				PlainText: proto.String("pass"),
			}.Build()
			config := configv1.Authentication_builder{
				BasicAuth: configv1.BasicAuth_builder{
					Password: secret,
				}.Build(),
			}.Build()
			_, err := NewUpstreamAuthenticator(config)
			assert.ErrorContains(t, err, "basic authentication requires a username")
		})

		t.Run("BasicAuth_MissingPassword", func(t *testing.T) {
			config := configv1.Authentication_builder{
				BasicAuth: configv1.BasicAuth_builder{
					Username: proto.String("user"),
				}.Build(),
			}.Build()
			_, err := NewUpstreamAuthenticator(config)
			assert.ErrorContains(t, err, "basic authentication requires a password")
		})

		t.Run("OAuth2_MissingClientID", func(t *testing.T) {
			clientSecret := configv1.SecretValue_builder{
				PlainText: proto.String("secret"),
			}.Build()
			config := configv1.Authentication_builder{
				Oauth2: configv1.OAuth2Auth_builder{
					ClientSecret: clientSecret,
					TokenUrl:     proto.String("http://token.url"),
				}.Build(),
			}.Build()
			_, err := NewUpstreamAuthenticator(config)
			assert.ErrorContains(t, err, "OAuth2 authentication requires a client ID")
		})

		t.Run("OAuth2_MissingClientSecret", func(t *testing.T) {
			clientID := configv1.SecretValue_builder{
				PlainText: proto.String("id"),
			}.Build()
			config := configv1.Authentication_builder{
				Oauth2: configv1.OAuth2Auth_builder{
					ClientId: clientID,
					TokenUrl: proto.String("http://token.url"),
				}.Build(),
			}.Build()
			_, err := NewUpstreamAuthenticator(config)
			assert.ErrorContains(t, err, "OAuth2 authentication requires a client secret")
		})

		t.Run("OAuth2_MissingTokenURL", func(t *testing.T) {
			clientID := configv1.SecretValue_builder{
				PlainText: proto.String("id"),
			}.Build()
			clientSecret := configv1.SecretValue_builder{
				PlainText: proto.String("secret"),
			}.Build()
			config := configv1.Authentication_builder{
				Oauth2: configv1.OAuth2Auth_builder{
					ClientId:     clientID,
					ClientSecret: clientSecret,
				}.Build(),
			}.Build()
			_, err := NewUpstreamAuthenticator(config)
			assert.ErrorContains(t, err, "OAuth2 authentication requires a token URL or an issuer URL")
		})

		t.Run("Empty_UpstreamAuthentication", func(t *testing.T) {
			config := configv1.Authentication_builder{}.Build()
			auth, err := NewUpstreamAuthenticator(config)
			assert.NoError(t, err)
			assert.Nil(t, auth)
		})
	})
}

func TestAPIKeyAuth_Authenticate(t *testing.T) {
	secret := configv1.SecretValue_builder{
		PlainText: proto.String("secret-key"),
	}.Build()

	t.Run("Header (Default)", func(t *testing.T) {
		auth := &APIKeyAuth{
			ParamName: "X-Custom-Auth",
			Value:     secret,
		}
		req, _ := http.NewRequest("GET", "/", nil)
		err := auth.Authenticate(req)
		assert.NoError(t, err)
		assert.Equal(t, "secret-key", req.Header.Get("X-Custom-Auth"))
	})

	t.Run("Query", func(t *testing.T) {
		auth := &APIKeyAuth{
			ParamName: "api_key",
			Value:     secret,
			Location:  configv1.APIKeyAuth_QUERY,
		}
		req, _ := http.NewRequest("GET", "/", nil)
		err := auth.Authenticate(req)
		assert.NoError(t, err)
		assert.Equal(t, "secret-key", req.URL.Query().Get("api_key"))
	})

	t.Run("Cookie", func(t *testing.T) {
		auth := &APIKeyAuth{
			ParamName: "auth_cookie",
			Value:     secret,
			Location:  configv1.APIKeyAuth_COOKIE,
		}
		req, _ := http.NewRequest("GET", "/", nil)
		err := auth.Authenticate(req)
		assert.NoError(t, err)
		cookie, err := req.Cookie("auth_cookie")
		assert.NoError(t, err)
		assert.Equal(t, "secret-key", cookie.Value)
	})
}

func TestBearerTokenAuth_Authenticate(t *testing.T) {
	secret := configv1.SecretValue_builder{
		PlainText: proto.String("secret-token"),
	}.Build()
	auth := &BearerTokenAuth{
		Token: secret,
	}
	req, _ := http.NewRequest("GET", "/", nil)
	err := auth.Authenticate(req)
	assert.NoError(t, err)
	assert.Equal(t, "Bearer secret-token", req.Header.Get("Authorization"))
}

func TestBasicAuth_Authenticate(t *testing.T) {
	secret := configv1.SecretValue_builder{
		PlainText: proto.String("testpassword"),
	}.Build()
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

func TestAPIKeyAuth_Authenticate_Error(t *testing.T) {
	auth := &APIKeyAuth{
		ParamName: "X-Custom-Auth",
		Value:     nil,
	}
	req, _ := http.NewRequest("GET", "/", nil)
	err := auth.Authenticate(req)
	assert.Error(t, err)
}

func TestBearerTokenAuth_Authenticate_Error(t *testing.T) {
	auth := &BearerTokenAuth{
		Token: nil,
	}
	req, _ := http.NewRequest("GET", "/", nil)
	err := auth.Authenticate(req)
	assert.Error(t, err)
}

func TestBasicAuth_Authenticate_Error(t *testing.T) {
	auth := &BasicAuth{
		Username: "testuser",
		Password: nil,
	}
	req, _ := http.NewRequest("GET", "/", nil)
	err := auth.Authenticate(req)
	assert.Error(t, err)
}

func TestOAuth2Auth_Authenticate_Errors(t *testing.T) {
	clientID := configv1.SecretValue_builder{
		PlainText: proto.String("id"),
	}.Build()
	clientSecret := configv1.SecretValue_builder{
		PlainText: proto.String("secret"),
	}.Build()

	t.Run("bad_token_url", func(t *testing.T) {
		auth := &OAuth2Auth{
			ClientID:     clientID,
			ClientSecret: clientSecret,
			TokenURL:     "not-a-url",
		}
		req, _ := http.NewRequest("GET", "/", nil)
		err := auth.Authenticate(req)
		assert.Error(t, err)
	})

	t.Run("token_fetch_error", func(t *testing.T) {
		auth := &OAuth2Auth{
			ClientID:     clientID,
			ClientSecret: clientSecret,
			TokenURL:     "http://127.0.0.1:12345/token",
		}
		req, _ := http.NewRequest("GET", "/", nil)
		err := auth.Authenticate(req)
		assert.Error(t, err)
	})

	t.Run("client_id_secret_error", func(t *testing.T) {
		auth := &OAuth2Auth{
			ClientID:     nil,
			ClientSecret: clientSecret,
			TokenURL:     "http://127.0.0.1",
		}
		req, _ := http.NewRequest("GET", "/", nil)
		err := auth.Authenticate(req)
		assert.Error(t, err)

		auth = &OAuth2Auth{
			ClientID:     clientID,
			ClientSecret: nil,
			TokenURL:     "http://127.0.0.1",
		}
		req, _ = http.NewRequest("GET", "/", nil)
		err = auth.Authenticate(req)
		assert.Error(t, err)
	})
}
