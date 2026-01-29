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

const (
	testAudience    = "test-audience"
	testEmail       = "test@example.com"
	wellKnownPath   = "/.well-known/openid-configuration"
	jwksPath        = "/jwks"
	contentTypeJSON = "application/json"
	authHeader      = "Authorization"
	bearerPrefix    = "Bearer "
)

func TestNewOAuth2Authenticator(t *testing.T) {
	privateKey := newTestKey(t)

	// Mock OIDC provider
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case wellKnownPath:
			w.Header().Set("Content-Type", contentTypeJSON)
			_, err := w.Write([]byte(`{
				"issuer": "http://` + r.Host + `",
				"jwks_uri": "http://` + r.Host + `/jwks"
			}`))
			assert.NoError(t, err)
		case "/jwks":
			w.Header().Set("Content-Type", "application/json")
			jwk := jose.JSONWebKey{Key: &privateKey.PublicKey, Algorithm: "RS256", Use: "sig"}
			_, err := w.Write([]byte(`{"keys": [` + string(mustMarshal(t, jwk)) + `]}`))
			assert.NoError(t, err)
		default:
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
		switch r.URL.Path {
		case wellKnownPath:
			w.Header().Set("Content-Type", contentTypeJSON)
			_, err := w.Write([]byte(`{
				"issuer": "http://` + r.Host + `",
				"jwks_uri": "http://` + r.Host + `/jwks"
			}`))
			assert.NoError(t, err)
		case jwksPath:
			w.Header().Set("Content-Type", contentTypeJSON)
			jwk := jose.JSONWebKey{Key: &privateKey.PublicKey, Algorithm: "RS256", Use: "sig"}
			_, err := w.Write([]byte(`{"keys": [` + string(mustMarshal(t, jwk)) + `]}`))
			assert.NoError(t, err)
		default:
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
			"iss":            server.URL,
			"aud":            "test-audience",
			"exp":            time.Now().Add(time.Hour).Unix(),
			"email":          "test@example.com",
			"email_verified": true,
		}
		token := newIDToken(t, privateKey, claims)

		req := httptest.NewRequest("GET", "/", nil)
		req.Header.Set(authHeader, bearerPrefix+token)

		ctx, err := authenticator.Authenticate(context.Background(), req)
		require.NoError(t, err)
		assert.Equal(t, testEmail, ctx.Value(UserContextKey))
	})

	t.Run("missing_authorization_header", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		_, err := authenticator.Authenticate(context.Background(), req)
		assert.Error(t, err)
		assert.Equal(t, "unauthorized", err.Error())
	})

	t.Run("invalid_authorization_header_format", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		req.Header.Set(authHeader, "invalid-token")
		_, err := authenticator.Authenticate(context.Background(), req)
		assert.Error(t, err)
		assert.Equal(t, "unauthorized", err.Error())
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
		req.Header.Set(authHeader, bearerPrefix+token)

		_, err := authenticator.Authenticate(context.Background(), req)
		assert.Error(t, err)
		assert.Equal(t, "unauthorized", err.Error())
	})

	t.Run("expired_token", func(t *testing.T) {
		claims := jwt.MapClaims{
			"iss":   server.URL,
			"aud":   testAudience,
			"exp":   time.Now().Add(-time.Hour).Unix(),
			"email": testEmail,
		}
		token := newIDToken(t, privateKey, claims)

		req := httptest.NewRequest("GET", "/", nil)
		req.Header.Set(authHeader, bearerPrefix+token)

		_, err := authenticator.Authenticate(context.Background(), req)
		assert.Error(t, err)
		assert.Equal(t, "unauthorized", err.Error())
	})
}

// mustMarshal is a helper function to marshal JSON without returning an error.
func mustMarshal(t *testing.T, v interface{}) []byte {
	bytes, err := json.Marshal(v)
	require.NoError(t, err)
	return bytes
}

func TestNewOAuth2Authenticator_Error(t *testing.T) {
	config := &OAuth2Config{
		IssuerURL: "http://127.0.0.1:12345",
	}
	_, err := NewOAuth2Authenticator(context.Background(), config)
	assert.Error(t, err)
}

func TestOAuth2Authenticator_Authenticate_ClaimError(t *testing.T) {
	privateKey := newTestKey(t)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case wellKnownPath:
			w.Header().Set("Content-Type", contentTypeJSON)
			_, err := w.Write([]byte(`{"issuer": "http://` + r.Host + `", "jwks_uri": "http://` + r.Host + `/jwks"}`))
			assert.NoError(t, err)
		case jwksPath:
			w.Header().Set("Content-Type", contentTypeJSON)
			jwk := jose.JSONWebKey{Key: &privateKey.PublicKey, Algorithm: "RS256", Use: "sig"}
			_, err := w.Write([]byte(`{"keys": [` + string(mustMarshal(t, jwk)) + `]}`))
			assert.NoError(t, err)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	config := &OAuth2Config{
		IssuerURL: server.URL,
		Audience:  testAudience,
	}

	authenticator, err := NewOAuth2Authenticator(context.Background(), config)
	require.NoError(t, err)
	require.NotNil(t, authenticator)

	claims := jwt.MapClaims{
		"iss":   server.URL,
		"aud":   testAudience,
		"exp":   time.Now().Add(time.Hour).Unix(),
		"email": 123, // Invalid email claim
	}
	token := newIDToken(t, privateKey, claims)

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set(authHeader, bearerPrefix+token)

	_, err = authenticator.Authenticate(context.Background(), req)
	assert.Error(t, err)
	assert.Equal(t, "unauthorized", err.Error())
}
