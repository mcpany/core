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
	"github.com/mcpany/core/server/pkg/util/passhash"
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
		config := configv1.Authentication_builder{
			Oauth2: configv1.OAuth2Auth_builder{
				Audience: proto.String("aud"),
				// Missing IssuerUrl
			}.Build(),
		}.Build()
		err := ValidateAuthentication(context.Background(), config, httptest.NewRequest("GET", "/", nil))
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "missing issuer_url")
	})

	// Test OAuth2 failed to create authenticator (unreachable issuer)
	t.Run("oauth2_unreachable_issuer", func(t *testing.T) {
		config := configv1.Authentication_builder{
			Oauth2: configv1.OAuth2Auth_builder{
				IssuerUrl: proto.String("http://127.0.0.1:12345"), // Unlikely to exist
				Audience:  proto.String("aud"),
			}.Build(),
		}.Build()
		// Short timeout is handled inside ValidateAuthentication now (10s), but connection refutes are fast usually
		err := ValidateAuthentication(context.Background(), config, httptest.NewRequest("GET", "/", nil))
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to create oauth2 authenticator")
	})

	// Test Basic Auth invalid config
	t.Run("basic_auth_invalid", func(t *testing.T) {
		config := configv1.Authentication_builder{
			BasicAuth: configv1.BasicAuth_builder{
				// Missing PasswordHash
			}.Build(),
		}.Build()
		err := ValidateAuthentication(context.Background(), config, httptest.NewRequest("GET", "/", nil))
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid Basic Auth configuration")
	})

	// Test Trusted Header invalid config
	t.Run("trusted_header_invalid", func(t *testing.T) {
		config := configv1.Authentication_builder{
			TrustedHeader: configv1.TrustedHeaderAuth_builder{
				// Missing HeaderName
			}.Build(),
		}.Build()
		err := ValidateAuthentication(context.Background(), config, httptest.NewRequest("GET", "/", nil))
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid Trusted Header configuration")
	})

	// Test OIDC missing issuer
	t.Run("oidc_missing_issuer", func(t *testing.T) {
		config := configv1.Authentication_builder{
			Oidc: configv1.OIDCAuth_builder{
				// Missing Issuer
			}.Build(),
		}.Build()
		err := ValidateAuthentication(context.Background(), config, httptest.NewRequest("GET", "/", nil))
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "missing issuer")
	})

	// Test OIDC valid config (mocked)
	t.Run("oidc_valid_mock", func(t *testing.T) {
		server := mockOIDCServer(t)
		defer server.Close()

		config := configv1.Authentication_builder{
			Oidc: configv1.OIDCAuth_builder{
				Issuer:   proto.String(server.URL),
				Audience: []string{"aud"},
			}.Build(),
		}.Build()

		// Authenticator creation should succeed, but authentication fails (no token)
		req := httptest.NewRequest("GET", "/", nil)
		err := ValidateAuthentication(context.Background(), config, req)
		assert.Error(t, err)
		assert.Equal(t, "unauthorized", err.Error())
	})

	// Test OIDC unreachable
	t.Run("oidc_unreachable", func(t *testing.T) {
		config := configv1.Authentication_builder{
			Oidc: configv1.OIDCAuth_builder{
				Issuer: proto.String("http://127.0.0.1:54321"),
			}.Build(),
		}.Build()
		err := ValidateAuthentication(context.Background(), config, httptest.NewRequest("GET", "/", nil))
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to create oidc authenticator")
	})
}

// TestContextHelpers is already in context_test.go, but we can add more specific cases or just skip it here.
// I will remove it from here to avoid redeclaration.

func TestNewManager_Accessors(t *testing.T) {
	m := NewManager()

	// Users
	u := configv1.User_builder{Id: proto.String("u1")}.Build()
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
	config := configv1.APIKeyAuth_builder{
		ParamName:         proto.String("auth_cookie"),
		VerificationValue: proto.String("secret"),
		In:                configv1.APIKeyAuth_COOKIE.Enum(),
	}.Build()
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
	config := configv1.TrustedHeaderAuth_builder{
		HeaderName: proto.String("X-User"),
		// HeaderValue empty -> any value ok
	}.Build()
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

func TestManager_CheckBasicAuthWithUsers(t *testing.T) {
	manager := NewManager()

	password := "secret123"
	hashed, _ := passhash.Password(password)

	user1 := configv1.User_builder{
		Id:    proto.String("user1"),
		Roles: []string{"admin"},
		Authentication: configv1.Authentication_builder{
			BasicAuth: configv1.BasicAuth_builder{
				PasswordHash: proto.String(hashed),
			}.Build(),
		}.Build(),
	}.Build()

	userNoAuth := configv1.User_builder{
		Id:    proto.String("userNoAuth"),
		Roles: []string{"guest"},
	}.Build()

	manager.SetUsers([]*configv1.User{user1, userNoAuth})

	t.Run("success", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		req.SetBasicAuth("user1", password)

		ctx, err := manager.Authenticate(context.Background(), "some-service", req)
		assert.NoError(t, err)

		userID, ok := UserFromContext(ctx)
		assert.True(t, ok)
		assert.Equal(t, "user1", userID)

		roles, ok := RolesFromContext(ctx)
		assert.True(t, ok)
		assert.Equal(t, []string{"admin"}, roles)
	})

	t.Run("invalid_password", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		req.SetBasicAuth("user1", "wrong")

		_, err := manager.Authenticate(context.Background(), "some-service", req)
		assert.Error(t, err)
	})

	t.Run("unknown_user", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		req.SetBasicAuth("unknown", password)

		_, err := manager.Authenticate(context.Background(), "some-service", req)
		assert.Error(t, err)
	})

	t.Run("user_without_basic_auth", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		req.SetBasicAuth("userNoAuth", password)

		_, err := manager.Authenticate(context.Background(), "some-service", req)
		assert.Error(t, err)
	})

	t.Run("no_basic_auth_header", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		// No header set

		_, err := manager.Authenticate(context.Background(), "some-service", req)
		assert.Error(t, err)
	})
}
