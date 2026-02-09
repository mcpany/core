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
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeOIDCProvider struct {
	server     *httptest.Server
	privateKey *rsa.PrivateKey
	publicKey  *rsa.PublicKey
	signer     jose.Signer
	issuerURL  string
}

func newFakeOIDCProvider(t *testing.T) *fakeOIDCProvider {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	publicKey := &privateKey.PublicKey

	signer, err := jose.NewSigner(jose.SigningKey{Algorithm: jose.RS256, Key: privateKey}, nil)
	require.NoError(t, err)

	p := &fakeOIDCProvider{
		privateKey: privateKey,
		publicKey:  publicKey,
		signer:     signer,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/.well-known/openid-configuration", p.handleDiscovery)
	mux.HandleFunc("/jwks", p.handleJWKS)
	mux.HandleFunc("/token", p.handleToken)

	p.server = httptest.NewServer(mux)
	p.issuerURL = p.server.URL

	return p
}

func (p *fakeOIDCProvider) Close() {
	p.server.Close()
}

func (p *fakeOIDCProvider) handleDiscovery(w http.ResponseWriter, r *http.Request) {
	config := map[string]interface{}{
		"issuer":                                p.issuerURL,
		"authorization_endpoint":                p.issuerURL + "/auth",
		"token_endpoint":                        p.issuerURL + "/token",
		"jwks_uri":                              p.issuerURL + "/jwks",
		"response_types_supported":              []string{"code", "id_token", "token"},
		"subject_types_supported":               []string{"public"},
		"id_token_signing_alg_values_supported": []string{"RS256"},
	}
	json.NewEncoder(w).Encode(config)
}

func (p *fakeOIDCProvider) handleJWKS(w http.ResponseWriter, r *http.Request) {
	jwk := jose.JSONWebKey{
		Key:       p.publicKey,
		KeyID:     "test-key-id",
		Algorithm: "RS256",
		Use:       "sig",
	}
	jwks := jose.JSONWebKeySet{
		Keys: []jose.JSONWebKey{jwk},
	}
	json.NewEncoder(w).Encode(jwks)
}

func (p *fakeOIDCProvider) handleToken(w http.ResponseWriter, r *http.Request) {
	// Simple token response for testing
	claims := map[string]interface{}{
		"sub":   "test-user-id",
		"email": "test@example.com",
		"iss":   p.issuerURL,
		"aud":   "test-client-id",
		"exp":   time.Now().Add(time.Hour).Unix(),
		"iat":   time.Now().Unix(),
	}

	// Check if code is valid (simple check)
	code := r.FormValue("code")
	if code == "invalid_code" {
		http.Error(w, "invalid_grant", http.StatusBadRequest)
		return
	}

	payload, err := json.Marshal(claims)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	signed, err := p.signer.Sign(payload)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	idToken, err := signed.CompactSerialize()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	resp := map[string]interface{}{
		"access_token": "fake-access-token",
		"token_type":   "Bearer",
		"id_token":     idToken,
		"expires_in":   3600,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func TestNewOIDCProvider(t *testing.T) {
	provider := newFakeOIDCProvider(t)
	defer provider.Close()

	config := OIDCConfig{
		Issuer:       provider.issuerURL,
		ClientID:     "test-client-id",
		ClientSecret: "test-client-secret",
		RedirectURL:  "http://localhost/callback",
	}

	p, err := NewOIDCProvider(context.Background(), config)
	require.NoError(t, err)
	assert.NotNil(t, p)
	assert.NotNil(t, p.provider)
	assert.NotNil(t, p.verifier)
}

func TestNewOIDCProvider_Error(t *testing.T) {
	config := OIDCConfig{
		Issuer:       "http://invalid-url",
		ClientID:     "test-client-id",
		ClientSecret: "test-client-secret",
		RedirectURL:  "http://localhost/callback",
	}

	p, err := NewOIDCProvider(context.Background(), config)
	assert.Error(t, err)
	assert.Nil(t, p)
}

func TestHandleLogin(t *testing.T) {
	provider := newFakeOIDCProvider(t)
	defer provider.Close()

	config := OIDCConfig{
		Issuer:       provider.issuerURL,
		ClientID:     "test-client-id",
		ClientSecret: "test-client-secret",
		RedirectURL:  "http://localhost/callback",
	}

	p, err := NewOIDCProvider(context.Background(), config)
	require.NoError(t, err)

	req := httptest.NewRequest("GET", "/login", nil)
	w := httptest.NewRecorder()

	p.HandleLogin(w, req)

	resp := w.Result()
	assert.Equal(t, http.StatusFound, resp.StatusCode)

	location := resp.Header.Get("Location")
	assert.Contains(t, location, provider.issuerURL+"/auth")
	assert.Contains(t, location, "client_id=test-client-id")
	assert.Contains(t, location, "redirect_uri=http%3A%2F%2Flocalhost%2Fcallback")
	assert.Contains(t, location, "response_type=code")
	assert.Contains(t, location, "scope=openid+profile+email")

	cookies := resp.Cookies()
	require.Len(t, cookies, 1)
	assert.Equal(t, "oauth_state", cookies[0].Name)
	assert.NotEmpty(t, cookies[0].Value)
}

func TestHandleCallback(t *testing.T) {
	provider := newFakeOIDCProvider(t)
	defer provider.Close()

	config := OIDCConfig{
		Issuer:       provider.issuerURL,
		ClientID:     "test-client-id",
		ClientSecret: "test-client-secret",
		RedirectURL:  "http://localhost/callback",
	}

	p, err := NewOIDCProvider(context.Background(), config)
	require.NoError(t, err)

	t.Run("MissingCookie", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/callback?code=foo&state=bar", nil)
		w := httptest.NewRecorder()

		p.HandleCallback(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "State cookie missing")
	})

	t.Run("InvalidState", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/callback?code=foo&state=invalid", nil)
		req.AddCookie(&http.Cookie{Name: "oauth_state", Value: "valid"})
		w := httptest.NewRecorder()

		p.HandleCallback(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "State invalid")
	})

	t.Run("ExchangeError", func(t *testing.T) {
		state := "valid-state"
		req := httptest.NewRequest("GET", "/callback?code=invalid_code&state="+state, nil)
		req.AddCookie(&http.Cookie{Name: "oauth_state", Value: state})
		w := httptest.NewRecorder()

		p.HandleCallback(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "Failed to exchange token")
	})

	t.Run("Success", func(t *testing.T) {
		state := "valid-state"
		req := httptest.NewRequest("GET", "/callback?code=valid_code&state="+state, nil)
		req.AddCookie(&http.Cookie{Name: "oauth_state", Value: state})
		w := httptest.NewRecorder()

		p.HandleCallback(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp map[string]string
		err := json.NewDecoder(w.Body).Decode(&resp)
		require.NoError(t, err)

		assert.Equal(t, "Authenticated", resp["message"])
		assert.Equal(t, "test@example.com", resp["email"])
		assert.Equal(t, "test-user-id", resp["user_id"])

		// Check cookie is cleared
		cookies := w.Result().Cookies()
		found := false
		for _, c := range cookies {
			if c.Name == "oauth_state" {
				assert.Equal(t, -1, c.MaxAge)
				found = true
			}
		}
		assert.True(t, found, "oauth_state cookie should be cleared")
	})
}
