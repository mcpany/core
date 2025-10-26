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

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/require"
	"gopkg.in/square/go-jose.v2"
)

func TestOAuth2Authenticator_Authenticate(t *testing.T) {
	// Create a key for signing the JWT
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	// Create a mock OAuth2 server
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/.well-known/openid-configuration" {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(struct {
				Issuer  string `json:"issuer"`
				JWKSURI string `json:"jwks_uri"`
			}{
				Issuer:  "http://" + r.Host,
				JWKSURI: "http://" + r.Host + "/.well-known/jwks.json",
			})
		} else if r.URL.Path == "/.well-known/jwks.json" {
			w.Header().Set("Content-Type", "application/json")
			jwk := jose.JSONWebKey{Key: &privateKey.PublicKey, Algorithm: "RS256"}
			json.NewEncoder(w).Encode(jose.JSONWebKeySet{
				Keys: []jose.JSONWebKey{jwk},
			})
		}
	}))
	defer mockServer.Close()

	// Create a new OAuth2 authenticator
	auth, err := NewOAuth2Authenticator(context.Background(), &OAuth2Config{
		IssuerURL: mockServer.URL,
		Audience:  "test-audience",
	})
	require.NoError(t, err)

	// Create a valid token
	claims := jwt.MapClaims{
		"exp": time.Now().Add(time.Hour).Unix(),
		"aud": "test-audience",
		"iss": mockServer.URL,
	}
	signedToken, err := jwt.NewWithClaims(jwt.SigningMethodRS256, claims).SignedString(privateKey)
	require.NoError(t, err)

	// Create a request with the valid token
	req, err := http.NewRequest("GET", "/", nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+signedToken)

	// Authenticate the request
	_, err = auth.Authenticate(context.Background(), req)
	require.NoError(t, err)

	// Create an invalid token
	invalidToken := "invalid-token"

	// Create a request with the invalid token
	req, err = http.NewRequest("GET", "/", nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+invalidToken)

	// Authenticate the request
	_, err = auth.Authenticate(context.Background(), req)
	require.Error(t, err)
}
