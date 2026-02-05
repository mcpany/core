// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package auth

import (
	"context"
	"crypto/tls"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOIDCProvider_HandleLogin_CookieSecurity(t *testing.T) {
	// Setup
	config := OIDCConfig{
		Issuer:       "https://issuer.example.com",
		ClientID:     "client-id",
		ClientSecret: "client-secret",
		RedirectURL:  "http://127.0.0.1/callback",
	}

	// Mock OIDC discovery
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/.well-known/openid-configuration" {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{
				"issuer": "` + config.Issuer + `",
				"authorization_endpoint": "` + config.Issuer + `/auth",
				"token_endpoint": "` + config.Issuer + `/token",
				"jwks_uri": "` + config.Issuer + `/jwks"
			}`))
			return
		}
	}))
	defer server.Close()

	// Update issuer to point to the mock server
	config.Issuer = server.URL

	provider, err := NewOIDCProvider(context.Background(), config)
	require.NoError(t, err)

	// Test Case 1: HTTP Request (Insecure)
	req := httptest.NewRequest("GET", "/login", nil)
	w := httptest.NewRecorder()

	provider.HandleLogin(w, req)

	resp := w.Result()
	cookies := resp.Cookies()
	require.Len(t, cookies, 1)
	cookie := cookies[0]

	assert.Equal(t, "oauth_state", cookie.Name)
	assert.True(t, cookie.HttpOnly)
	assert.Equal(t, "/", cookie.Path)

	// Check NEW Behavior
	assert.False(t, cookie.Secure, "Secure flag should be false for HTTP request")
	assert.Equal(t, http.SameSiteLaxMode, cookie.SameSite, "SameSite should be Lax")

	// Test Case 2: X-Forwarded-Proto HTTPS (Should be Secure in improved impl)
	reqSecure := httptest.NewRequest("GET", "/login", nil)
	reqSecure.Header.Set("X-Forwarded-Proto", "https")
	wSecure := httptest.NewRecorder()

	provider.HandleLogin(wSecure, reqSecure)
	respSecure := wSecure.Result()
	cookiesSecure := respSecure.Cookies()
	require.Len(t, cookiesSecure, 1)
	assert.True(t, cookiesSecure[0].Secure, "Secure flag should be true with X-Forwarded-Proto")
	assert.Equal(t, http.SameSiteLaxMode, cookiesSecure[0].SameSite, "SameSite should be Lax")

	// Test Case 3: TLS (Should be Secure)
	reqTLS := httptest.NewRequest("GET", "/login", nil)
	reqTLS.TLS = &tls.ConnectionState{}
	wTLS := httptest.NewRecorder()

	provider.HandleLogin(wTLS, reqTLS)
	respTLS := wTLS.Result()
	cookiesTLS := respTLS.Cookies()
	require.Len(t, cookiesTLS, 1)
	assert.True(t, cookiesTLS[0].Secure, "Secure flag should be true for TLS request")
	assert.Equal(t, http.SameSiteLaxMode, cookiesTLS[0].SameSite, "SameSite should be Lax")
}
