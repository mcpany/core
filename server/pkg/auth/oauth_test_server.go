// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package auth

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-jose/go-jose/v4"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockOAuth2Server serves as a mock OIDC/OAuth2 provider.
//
// Summary: serves as a mock OIDC/OAuth2 provider.
type MockOAuth2Server struct {
	*httptest.Server
	PrivateKey *rsa.PrivateKey
}

// NewMockOAuth2Server creates a new mock OAuth2 server.
//
// Summary: creates a new mock OAuth2 server.
//
// Parameters:
//   - t: *testing.T. The t.
//
// Returns:
//   - *MockOAuth2Server: The *MockOAuth2Server.
func NewMockOAuth2Server(t *testing.T) *MockOAuth2Server {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	mux := http.NewServeMux()
	server := httptest.NewServer(mux)

	mock := &MockOAuth2Server{
		Server:     server,
		PrivateKey: privateKey,
	}

	mux.HandleFunc("/.well-known/openid-configuration", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		config := map[string]interface{}{
			"issuer":                 server.URL,
			"jwks_uri":               server.URL + "/jwks",
			"authorization_endpoint": server.URL + "/auth",
			"token_endpoint":         server.URL + "/token",
		}
		_ = json.NewEncoder(w).Encode(config)
	})

	mux.HandleFunc("/jwks", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		jwk := jose.JSONWebKey{Key: &privateKey.PublicKey, Algorithm: "RS256", Use: "sig"}
		jwks := map[string]interface{}{
			"keys": []interface{}{jwk},
		}
		_ = json.NewEncoder(w).Encode(jwks)
	})

	// Add endpoints for Auth and Client Credentials flows if needed
	mux.HandleFunc("/auth", func(w http.ResponseWriter, r *http.Request) {
		// Mock Authorization Endpoint
		state := r.URL.Query().Get("state")
		redirectURI := r.URL.Query().Get("redirect_uri")
		if redirectURI == "" {
			http.Error(w, "missing redirect_uri", http.StatusBadRequest)
			return
		}
		// Redirect back with code
		http.Redirect(w, r, redirectURI+"?code=mock_code&state="+state, http.StatusFound)
	})

	mux.HandleFunc("/token", func(w http.ResponseWriter, _ *http.Request) {
		// Mock Token Endpoint
		w.Header().Set("Content-Type", "application/json")
		token := mock.NewIDToken(t, jwt.MapClaims{
			"iss": server.URL,
			"sub": "mock_user",
			"aud": "mock_client",
			"exp": time.Now().Add(time.Hour).Unix(),
		})
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"access_token": token,
			"id_token":     token,
			"token_type":   "Bearer",
			"expires_in":   3600,
		})
	})

	return mock
}

// NewIDToken permits generating custom tokens signed by this server.
//
// Summary: permits generating custom tokens signed by this server.
//
// Parameters:
//   - t: *testing.T. The t.
//   - claims: jwt.MapClaims. The claims.
//
// Returns:
//   - string: The string.
func (s *MockOAuth2Server) NewIDToken(t *testing.T, claims jwt.MapClaims) string {
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	signedToken, err := token.SignedString(s.PrivateKey)
	assert.NoError(t, err)
	return signedToken
}
