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
	"context"
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

// newTestKey creates a new RSA private key for signing JWTs.
func newTestKey(t *testing.T) *rsa.PrivateKey {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	return privateKey
}

// newIDToken creates a new JWT ID token with the specified claims.
func newIDToken(t *testing.T, privateKey *rsa.PrivateKey, claims jwt.MapClaims) string {
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	signedToken, err := token.SignedString(privateKey)
	require.NoError(t, err)
	return signedToken
}

func TestNewOAuth2Authenticator(t *testing.T) {
	privateKey := newTestKey(t)

	// Mock OIDC provider
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/.well-known/openid-configuration" {
			w.Header().Set("Content-Type", "application/json")
			_, err := w.Write([]byte(`{
				"issuer": "http://` + r.Host + `",
				"jwks_uri": "http://` + r.Host + `/jwks"
			}`))
			assert.NoError(t, err)
		} else if r.URL.Path == "/jwks" {
			w.Header().Set("Content-Type", "application/json")
			jwk := jose.JSONWebKey{Key: &privateKey.PublicKey, Algorithm: "RS256", Use: "sig"}
			_, err := w.Write([]byte(`{"keys": [` + string(mustMarshal(t, jwk)) + `]}`))
			assert.NoError(t, err)
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	config := &OAuth2Config{
		IssuerURL: server.URL,
		Audience:  "test-audience",
	}

	authenticator, err := NewOAuth2Authenticator(context.Background(), config)
	require.NoError(t, err)
	assert.NotNil(t, authenticator)
}

func TestOAuth2Authenticator_Authenticate(t *testing.T) {
	privateKey := newTestKey(t)

	// Mock OIDC provider
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/.well-known/openid-configuration" {
			w.Header().Set("Content-Type", "application/json")
			_, err := w.Write([]byte(`{
				"issuer": "http://` + r.Host + `",
				"jwks_uri": "http://` + r.Host + `/jwks"
			}`))
			assert.NoError(t, err)
		} else if r.URL.Path == "/jwks" {
			w.Header().Set("Content-Type", "application/json")
			jwk := jose.JSONWebKey{Key: &privateKey.PublicKey, Algorithm: "RS256", Use: "sig"}
			_, err := w.Write([]byte(`{"keys": [` + string(mustMarshal(t, jwk)) + `]}`))
			assert.NoError(t, err)
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	config := &OAuth2Config{
		IssuerURL: server.URL,
		Audience:  "test-audience",
	}

	authenticator, err := NewOAuth2Authenticator(context.Background(), config)
	require.NoError(t, err)
	require.NotNil(t, authenticator)

	t.Run("successful_authentication", func(t *testing.T) {
		claims := jwt.MapClaims{
			"iss":   server.URL,
			"aud":   "test-audience",
			"exp":   time.Now().Add(time.Hour).Unix(),
			"email": "test@example.com",
		}
		token := newIDToken(t, privateKey, claims)

		req := httptest.NewRequest("GET", "/", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		ctx, err := authenticator.Authenticate(context.Background(), req)
		require.NoError(t, err)
		assert.Equal(t, "test@example.com", ctx.Value("user"))
	})

	t.Run("missing_authorization_header", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		_, err := authenticator.Authenticate(context.Background(), req)
		assert.Error(t, err)
		assert.Equal(t, "missing Authorization header", err.Error())
	})

	t.Run("invalid_authorization_header_format", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		req.Header.Set("Authorization", "invalid-token")
		_, err := authenticator.Authenticate(context.Background(), req)
		assert.Error(t, err)
		assert.Equal(t, "invalid Authorization header format", err.Error())
	})

	t.Run("token_verification_failed", func(t *testing.T) {
		// Token signed with a different key
		wrongKey := newTestKey(t)
		claims := jwt.MapClaims{
			"iss":   server.URL,
			"aud":   "test-audience",
			"exp":   time.Now().Add(time.Hour).Unix(),
			"email": "test@example.com",
		}
		token := newIDToken(t, wrongKey, claims)

		req := httptest.NewRequest("GET", "/", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		_, err := authenticator.Authenticate(context.Background(), req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to verify token")
	})

	t.Run("expired_token", func(t *testing.T) {
		claims := jwt.MapClaims{
			"iss":   server.URL,
			"aud":   "test-audience",
			"exp":   time.Now().Add(-time.Hour).Unix(),
			"email": "test@example.com",
		}
		token := newIDToken(t, privateKey, claims)

		req := httptest.NewRequest("GET", "/", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		_, err := authenticator.Authenticate(context.Background(), req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to verify token")
	})
}

// mustMarshal is a helper function to marshal JSON without returning an error.
func mustMarshal(t *testing.T, v interface{}) []byte {
	bytes, err := json.Marshal(v)
	require.NoError(t, err)
	return bytes
}
