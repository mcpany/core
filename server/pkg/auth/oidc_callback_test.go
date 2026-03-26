// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

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
)

func TestOIDCProvider_HandleCallback(t *testing.T) {
	// Generate RSA key for signing tokens
	rsaKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	signer, err := jose.NewSigner(jose.SigningKey{Algorithm: jose.RS256, Key: rsaKey}, nil)
	require.NoError(t, err)

	// Mock OIDC Provider Server
	providerServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/.well-known/openid-configuration":
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]string{
				"issuer":                 "http://" + r.Host,
				"authorization_endpoint": "http://" + r.Host + "/auth",
				"token_endpoint":         "http://" + r.Host + "/token",
				"jwks_uri":               "http://" + r.Host + "/jwks",
			})
		case "/jwks":
			jwk := jose.JSONWebKey{
				Key:       &rsaKey.PublicKey,
				KeyID:     "test-key",
				Algorithm: "RS256",
				Use:       "sig",
			}
			jwks := jose.JSONWebKeySet{
				Keys: []jose.JSONWebKey{jwk},
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(jwks)
		case "/token":
			// Create ID Token
			claims := map[string]interface{}{
				"iss":   "http://" + r.Host,
				"sub":   "test-user-id",
				"aud":   "client-id",
				"exp":   time.Now().Add(time.Hour).Unix(),
				"iat":   time.Now().Unix(),
				"email": "test@example.com",
			}
			payload, err := json.Marshal(claims)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			signed, err := signer.Sign(payload)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			idToken, err := signed.CompactSerialize()
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"access_token": "mock-access-token",
				"token_type":   "Bearer",
				"expires_in":   3600,
				"id_token":     idToken,
			})
		default:
			http.NotFound(w, r)
		}
	}))
	defer providerServer.Close()

	config := OIDCConfig{
		Issuer:       providerServer.URL,
		ClientID:     "client-id",
		ClientSecret: "client-secret",
		RedirectURL:  "http://127.0.0.1/callback",
	}

	provider, err := NewOIDCProvider(context.Background(), config)
	require.NoError(t, err)

	// Override verifier with one that trusts the mock server's keys
	// NOTE: oidc.NewProvider fetches keys on initialization. Since we initialized it with the mock server URL,
	// it should already have the keys. However, the verifier needs to know the correct issuer.
	// In the real code, NewOIDCProvider creates the verifier.
	// We need to make sure the issuer in the token matches the provider URL.

	t.Run("Valid Callback", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/callback?code=valid-code&state=test-state", nil)
		req.AddCookie(&http.Cookie{
			Name:  "oauth_state",
			Value: "test-state",
		})
		w := httptest.NewRecorder()

		provider.HandleCallback(w, req)

		resp := w.Result()
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var body map[string]string
		err := json.NewDecoder(resp.Body).Decode(&body)
		require.NoError(t, err)
		assert.Equal(t, "Authenticated", body["message"])
		assert.Equal(t, "test@example.com", body["email"])
		assert.Equal(t, "test-user-id", body["user_id"])

		// Check if state cookie is cleared
		cookies := resp.Cookies()
		require.Len(t, cookies, 1)
		assert.Equal(t, "oauth_state", cookies[0].Name)
		assert.Equal(t, "", cookies[0].Value)
		assert.True(t, cookies[0].MaxAge < 0)
	})

	t.Run("Missing State Cookie", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/callback?code=valid-code&state=test-state", nil)
		w := httptest.NewRecorder()

		provider.HandleCallback(w, req)

		resp := w.Result()
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("Invalid State", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/callback?code=valid-code&state=invalid-state", nil)
		req.AddCookie(&http.Cookie{
			Name:  "oauth_state",
			Value: "test-state",
		})
		w := httptest.NewRecorder()

		provider.HandleCallback(w, req)

		resp := w.Result()
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("Exchange Error", func(t *testing.T) {
		// Create a provider that points to a server that returns error on token endpoint
		errorServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/token" {
				http.Error(w, "Exchange failed", http.StatusInternalServerError)
				return
			}
			// Forward other requests to original provider server
			// But for simplicity, we just need discovery to work to create the provider
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
			if r.URL.Path == "/jwks" {
				// Empty JWKS or whatever, we fail before verification
				w.Header().Set("Content-Type", "application/json")
				w.Write([]byte(`{"keys":[]}`))
				return
			}
		}))
		defer errorServer.Close()

		configErr := config
		configErr.Issuer = errorServer.URL
		providerErr, err := NewOIDCProvider(context.Background(), configErr)
		require.NoError(t, err)

		req := httptest.NewRequest("GET", "/callback?code=valid-code&state=test-state", nil)
		req.AddCookie(&http.Cookie{
			Name:  "oauth_state",
			Value: "test-state",
		})
		w := httptest.NewRecorder()

		providerErr.HandleCallback(w, req)

		resp := w.Result()
		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	})
}
