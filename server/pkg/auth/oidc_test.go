// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package auth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewOIDCProvider(t *testing.T) {
	mockServer := NewMockOAuth2Server(t)
	defer mockServer.Close()

	ctx := context.Background()
	config := OIDCConfig{
		Issuer:       mockServer.URL,
		ClientID:     "test-client",
		ClientSecret: "test-secret",
		RedirectURL:  "http://localhost:8080/callback",
	}

	provider, err := NewOIDCProvider(ctx, config)
	require.NoError(t, err)
	assert.NotNil(t, provider)
	assert.Equal(t, config, provider.config)
	assert.NotNil(t, provider.provider)
	assert.NotNil(t, provider.verifier)
	assert.Equal(t, mockServer.URL+"/token", provider.oauth2Config.Endpoint.TokenURL)
	assert.Equal(t, mockServer.URL+"/auth", provider.oauth2Config.Endpoint.AuthURL)
}

func TestOIDCProvider_HandleLogin_Redirect(t *testing.T) {
	mockServer := NewMockOAuth2Server(t)
	defer mockServer.Close()

	ctx := context.Background()
	config := OIDCConfig{
		Issuer:       mockServer.URL,
		ClientID:     "test-client",
		ClientSecret: "test-secret",
		RedirectURL:  "http://localhost:8080/callback",
	}

	provider, err := NewOIDCProvider(ctx, config)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/login", nil)
	w := httptest.NewRecorder()

	provider.HandleLogin(w, req)

	resp := w.Result()
	assert.Equal(t, http.StatusFound, resp.StatusCode)

	// Check Redirect URL
	location := resp.Header.Get("Location")
	parsedURL, err := url.Parse(location)
	require.NoError(t, err)
	assert.Equal(t, mockServer.URL, parsedURL.Scheme+"://"+parsedURL.Host)
	assert.Equal(t, "/auth", parsedURL.Path)

	query := parsedURL.Query()
	assert.Equal(t, "test-client", query.Get("client_id"))
	assert.Equal(t, "http://localhost:8080/callback", query.Get("redirect_uri"))
	assert.Equal(t, "code", query.Get("response_type"))
	assert.Contains(t, query.Get("scope"), "openid")

	// Check State Cookie
	cookies := resp.Cookies()
	require.Len(t, cookies, 1)
	cookie := cookies[0]
	assert.Equal(t, "oauth_state", cookie.Name)
	assert.NotEmpty(t, cookie.Value)
	assert.True(t, cookie.HttpOnly)
	assert.Equal(t, "/", cookie.Path)
	assert.Equal(t, 300, cookie.MaxAge)
}
