// Copyright 2026 Author(s) of MCP Any
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
	"google.golang.org/protobuf/proto"
)

// Helper to create a mock OIDC server
func mockOIDCServer(t *testing.T) *httptest.Server {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/.well-known/openid-configuration" {
			w.Header().Set("Content-Type", "application/json")
			// Return minimal config
			fmt.Fprintf(w, `{"issuer": "%s", "jwks_uri": "%s/jwks"}`, "http://"+r.Host, "http://"+r.Host)
			return
		}
		if r.URL.Path == "/jwks" {
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprintln(w, `{"keys": []}`)
			return
		}
		http.NotFound(w, r)
	}))
	return server
}

func TestValidateAuthentication_Extended(t *testing.T) {
	// Test OAuth2 invalid config (missing issuer)
	t.Run("oauth2_missing_issuer", func(t *testing.T) {
		config := &configv1.Authentication{
			AuthMethod: &configv1.Authentication_Oauth2{
				Oauth2: &configv1.OAuth2Auth{
					Audience: proto.String("aud"),
					// Missing IssuerUrl
				},
			},
		}
		err := ValidateAuthentication(context.Background(), config, httptest.NewRequest("GET", "/", nil))
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "missing issuer_url")
	})

	// Test OAuth2 failed to create authenticator (unreachable issuer)
	t.Run("oauth2_unreachable_issuer", func(t *testing.T) {
		config := &configv1.Authentication{
			AuthMethod: &configv1.Authentication_Oauth2{
				Oauth2: &configv1.OAuth2Auth{
					IssuerUrl: proto.String("http://localhost:12345"), // Unlikely to exist
					Audience:  proto.String("aud"),
				},
			},
		}
		// Short timeout is handled inside ValidateAuthentication now (10s), but connection refutes are fast usually
		err := ValidateAuthentication(context.Background(), config, httptest.NewRequest("GET", "/", nil))
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to create oauth2 authenticator")
	})

	// Test Basic Auth invalid config
	t.Run("basic_auth_invalid", func(t *testing.T) {
		config := &configv1.Authentication{
			AuthMethod: &configv1.Authentication_BasicAuth{
				BasicAuth: &configv1.BasicAuth{
					// Missing PasswordHash
				},
			},
		}
		err := ValidateAuthentication(context.Background(), config, httptest.NewRequest("GET", "/", nil))
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid Basic Auth configuration")
	})

	// Test Trusted Header invalid config
	t.Run("trusted_header_invalid", func(t *testing.T) {
		config := &configv1.Authentication{
			AuthMethod: &configv1.Authentication_TrustedHeader{
				TrustedHeader: &configv1.TrustedHeaderAuth{
					// Missing HeaderName
				},
			},
		}
		err := ValidateAuthentication(context.Background(), config, httptest.NewRequest("GET", "/", nil))
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid Trusted Header configuration")
	})

	// Test OIDC missing issuer
	t.Run("oidc_missing_issuer", func(t *testing.T) {
		config := &configv1.Authentication{
			AuthMethod: &configv1.Authentication_Oidc{
				Oidc: &configv1.OIDCAuth{
					// Missing Issuer
				},
			},
		}
		err := ValidateAuthentication(context.Background(), config, httptest.NewRequest("GET", "/", nil))
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "missing issuer")
	})

	// Test OIDC valid config (mocked)
	t.Run("oidc_valid_mock", func(t *testing.T) {
		server := mockOIDCServer(t)
		defer server.Close()

		config := &configv1.Authentication{
			AuthMethod: &configv1.Authentication_Oidc{
				Oidc: &configv1.OIDCAuth{
					Issuer:   proto.String(server.URL),
					Audience: []string{"aud"},
				},
			},
		}

		// Authenticator creation should succeed, but authentication fails (no token)
		req := httptest.NewRequest("GET", "/", nil)
		err := ValidateAuthentication(context.Background(), config, req)
		assert.Error(t, err)
		assert.Equal(t, "unauthorized", err.Error())
	})

	// Test OIDC unreachable
	t.Run("oidc_unreachable", func(t *testing.T) {
		config := &configv1.Authentication{
			AuthMethod: &configv1.Authentication_Oidc{
				Oidc: &configv1.OIDCAuth{
					Issuer: proto.String("http://localhost:54321"),
				},
			},
		}
		err := ValidateAuthentication(context.Background(), config, httptest.NewRequest("GET", "/", nil))
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to create oidc authenticator")
	})
}

func TestContextHelpers(t *testing.T) {
	ctx := context.Background()

	// APIKey
	ctx = ContextWithAPIKey(ctx, "key")
	val, ok := APIKeyFromContext(ctx)
	assert.True(t, ok)
	assert.Equal(t, "key", val)

	// User
	ctx = ContextWithUser(ctx, "user")
	val, ok = UserFromContext(ctx)
	assert.True(t, ok)
	assert.Equal(t, "user", val)

	// ProfileID
	ctx = ContextWithProfileID(ctx, "profile")
	val, ok = ProfileIDFromContext(ctx)
	assert.True(t, ok)
	assert.Equal(t, "profile", val)
}

func TestNewManager_Accessors(t *testing.T) {
	m := NewManager()

	// Users
	u := &configv1.User{Id: proto.String("u1")}
	m.SetUsers([]*configv1.User{u})
	got, ok := m.GetUser("u1")
	assert.True(t, ok)
	assert.Equal(t, u, got)

	_, ok = m.GetUser("u2")
	assert.False(t, ok)

	// Storage (just coverage)
	m.SetStorage(nil)
}

func TestAPIKeyAuthenticator_Cookie(t *testing.T) {
	config := &configv1.APIKeyAuth{
		ParamName: proto.String("auth_cookie"),
		VerificationValue: proto.String("secret"),
		In: configv1.APIKeyAuth_COOKIE.Enum(),
	}
	auth := NewAPIKeyAuthenticator(config)

	t.Run("success", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		req.AddCookie(&http.Cookie{Name: "auth_cookie", Value: "secret"})
		_, err := auth.Authenticate(context.Background(), req)
		assert.NoError(t, err)
	})

	t.Run("fail_wrong_value", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		req.AddCookie(&http.Cookie{Name: "auth_cookie", Value: "wrong"})
		_, err := auth.Authenticate(context.Background(), req)
		assert.Error(t, err)
	})

	t.Run("fail_missing_cookie", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		_, err := auth.Authenticate(context.Background(), req)
		assert.Error(t, err)
	})
}

func TestTrustedHeaderAuthenticator_NoValue(t *testing.T) {
	config := &configv1.TrustedHeaderAuth{
		HeaderName: proto.String("X-User"),
		// HeaderValue empty -> any value ok
	}
	auth := NewTrustedHeaderAuthenticator(config)

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("X-User", "bob")
	_, err := auth.Authenticate(context.Background(), req)
	assert.NoError(t, err)

	req = httptest.NewRequest("GET", "/", nil)
	// Missing header
	_, err = auth.Authenticate(context.Background(), req)
	assert.Error(t, err)
}
