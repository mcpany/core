// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package auth

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupMockOIDCServer creates a test HTTP server that mocks an OIDC discovery endpoint.
func setupMockOIDCServer(t *testing.T) *httptest.Server {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/.well-known/openid-configuration" {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]string{
				"issuer":                 "http://" + r.Host,
				"authorization_endpoint": "http://" + r.Host + "/auth",
				"token_endpoint":         "http://" + r.Host + "/token",
				"jwks_uri":               "http://" + r.Host + "/jwks",
			})
			return
		}
		http.NotFound(w, r)
	}))
	return server
}

func TestNewOIDCProvider(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		server := setupMockOIDCServer(t)
		defer server.Close()

		config := OIDCConfig{
			Issuer:       server.URL,
			ClientID:     "client-id",
			ClientSecret: "client-secret",
			RedirectURL:  "http://localhost/callback",
		}

		provider, err := NewOIDCProvider(context.Background(), config)
		require.NoError(t, err)
		assert.NotNil(t, provider)
		assert.Equal(t, config, provider.config)
	})

	t.Run("Failure_InvalidIssuer", func(t *testing.T) {
		config := OIDCConfig{
			Issuer:       "http://invalid-url-that-does-not-exist.local",
			ClientID:     "client-id",
			ClientSecret: "client-secret",
			RedirectURL:  "http://localhost/callback",
		}

		provider, err := NewOIDCProvider(context.Background(), config)
		assert.Error(t, err)
		assert.Nil(t, provider)
	})

	t.Run("Failure_ServerReturnsError", func(t *testing.T) {
		// Server that returns 500
		errorServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer errorServer.Close()

		config := OIDCConfig{
			Issuer:       errorServer.URL,
			ClientID:     "client-id",
			ClientSecret: "client-secret",
			RedirectURL:  "http://localhost/callback",
		}

		provider, err := NewOIDCProvider(context.Background(), config)
		assert.Error(t, err)
		assert.Nil(t, provider)
	})
}

func TestHandleLogin_Redirect(t *testing.T) {
	server := setupMockOIDCServer(t)
	defer server.Close()

	config := OIDCConfig{
		Issuer:       server.URL,
		ClientID:     "test-client-id",
		ClientSecret: "test-secret",
		RedirectURL:  "http://localhost/callback",
	}

	provider, err := NewOIDCProvider(context.Background(), config)
	require.NoError(t, err)

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/login", nil)

	provider.HandleLogin(w, r)

	resp := w.Result()
	assert.Equal(t, http.StatusFound, resp.StatusCode)

	location := resp.Header.Get("Location")
	parsedURL, err := url.Parse(location)
	require.NoError(t, err)

	// Verify base URL
	assert.Equal(t, "http", parsedURL.Scheme)
	assert.Equal(t, server.Listener.Addr().String(), parsedURL.Host)
	assert.Equal(t, "/auth", parsedURL.Path)

	// Verify Query Params
	query := parsedURL.Query()
	assert.Equal(t, "test-client-id", query.Get("client_id"))
	assert.Equal(t, "http://localhost/callback", query.Get("redirect_uri"))
	assert.Equal(t, "code", query.Get("response_type"))
	assert.Contains(t, query.Get("scope"), "openid")
	assert.Contains(t, query.Get("scope"), "profile")
	assert.Contains(t, query.Get("scope"), "email")
	assert.NotEmpty(t, query.Get("state"))

	// Verify Cookie
	cookies := resp.Cookies()
	require.Len(t, cookies, 1)
	cookie := cookies[0]
	assert.Equal(t, "oauth_state", cookie.Name)
	assert.Equal(t, query.Get("state"), cookie.Value)
}

func TestHandleLogin_StateRandomness(t *testing.T) {
	server := setupMockOIDCServer(t)
	defer server.Close()

	config := OIDCConfig{
		Issuer:       server.URL,
		ClientID:     "test-client-id",
		ClientSecret: "test-secret",
		RedirectURL:  "http://localhost/callback",
	}

	provider, err := NewOIDCProvider(context.Background(), config)
	require.NoError(t, err)

	// First Request
	w1 := httptest.NewRecorder()
	r1 := httptest.NewRequest("GET", "/login", nil)
	provider.HandleLogin(w1, r1)
	loc1 := w1.Result().Header.Get("Location")
	u1, _ := url.Parse(loc1)
	state1 := u1.Query().Get("state")

	// Second Request
	w2 := httptest.NewRecorder()
	r2 := httptest.NewRequest("GET", "/login", nil)
	provider.HandleLogin(w2, r2)
	loc2 := w2.Result().Header.Get("Location")
	u2, _ := url.Parse(loc2)
	state2 := u2.Query().Get("state")

	assert.NotEmpty(t, state1)
	assert.NotEmpty(t, state2)
	assert.NotEqual(t, state1, state2, "States should be random and different")
}
