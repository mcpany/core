/*
 * Copyright 2025 Author(s) of MCP-XY
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

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/square/go-jose.v2"
	"gopkg.in/square/go-jose.v2/jwt"
)

func TestAuthManager(t *testing.T) {
	// Generate a mock RSA private key
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	// Create a JWK from the private key
	jwk := jose.JSONWebKey{Key: privateKey, Algorithm: string(jose.RS256)}

	// Create a mock OIDC provider
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/.well-known/openid-configuration" {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(struct {
				Issuer  string `json:"issuer"`
				JWKSURI string `json:"jwks_uri"`
			}{
				Issuer:  "http://" + r.Host,
				JWKSURI: "http://" + r.Host + "/jwks",
			})
		} else if r.URL.Path == "/jwks" {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(jose.JSONWebKeySet{
				Keys: []jose.JSONWebKey{jwk.Public()},
			})
		}
	}))
	defer server.Close()

	// Create a new AuthManager with the mock provider's URL
	authManager := NewAuthManager(server.URL, "test-audience", "http://localhost:50050")
	require.NotNil(t, authManager)

	// Create a signer from the private key
	signer, err := jose.NewSigner(jose.SigningKey{Algorithm: jose.RS256, Key: jwk}, nil)
	require.NoError(t, err)

	t.Run("successful_token_verification", func(t *testing.T) {
		claims := jwt.Claims{
			Issuer:   server.URL,
			Audience: jwt.Audience{"test-audience"},
			Expiry:   jwt.NewNumericDate(time.Now().Add(1 * time.Hour)),
		}
		rawToken, err := jwt.Signed(signer).Claims(claims).CompactSerialize()
		require.NoError(t, err)

		_, err = authManager.VerifyToken(context.Background(), rawToken)
		assert.NoError(t, err)
	})

	t.Run("failed_token_verification_wrong_issuer", func(t *testing.T) {
		claims := jwt.Claims{
			Issuer:   "wrong-issuer",
			Audience: jwt.Audience{"test-audience"},
			Expiry:   jwt.NewNumericDate(time.Now().Add(1 * time.Hour)),
		}
		rawToken, err := jwt.Signed(signer).Claims(claims).CompactSerialize()
		require.NoError(t, err)

		_, err = authManager.VerifyToken(context.Background(), rawToken)
		assert.Error(t, err)
	})

	t.Run("failed_token_verification_wrong_audience", func(t *testing.T) {
		claims := jwt.Claims{
			Issuer:   server.URL,
			Audience: jwt.Audience{"wrong-audience"},
			Expiry:   jwt.NewNumericDate(time.Now().Add(1 * time.Hour)),
		}
		rawToken, err := jwt.Signed(signer).Claims(claims).CompactSerialize()
		require.NoError(t, err)

		_, err = authManager.VerifyToken(context.Background(), rawToken)
		assert.Error(t, err)
	})

	t.Run("failed_token_verification_expired_token", func(t *testing.T) {
		claims := jwt.Claims{
			Issuer:   server.URL,
			Audience: jwt.Audience{"test-audience"},
			Expiry:   jwt.NewNumericDate(time.Now().Add(-1 * time.Hour)),
		}
		rawToken, err := jwt.Signed(signer).Claims(claims).CompactSerialize()
		require.NoError(t, err)

		_, err = authManager.VerifyToken(context.Background(), rawToken)
		assert.Error(t, err)
	})
}
