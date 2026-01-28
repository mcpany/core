// Package main implements a mock OAuth/OIDC server for testing purposes.
package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/go-jose/go-jose/v4"
)

// OAuthServer is a mock OAuth server.
type OAuthServer struct {
	Issuer string
	Key    *rsa.PrivateKey
}

func main() {
	port := flag.Int("port", 8081, "Port to listen on")
	flag.Parse()

	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		log.Fatalf("Failed to generate key: %v", err)
	}

	// Ensure key generation helper works (restoring "test case" logic)
	_ = publicKeyToPEM(&key.PublicKey)

	server := &OAuthServer{
		Issuer: fmt.Sprintf("http://localhost:%d", *port),
		Key:    key,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/.well-known/openid-configuration", server.handleDiscovery)
	mux.HandleFunc("/jwks.json", server.handleJWKS)
	mux.HandleFunc("/authorize", server.handleAuthorize)
	mux.HandleFunc("/token", server.handleToken)
	mux.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("OK"))
	})

	log.Printf("Starting mock OAuth server on port %d", *port)
	srv := &http.Server{
		Addr:              fmt.Sprintf(":%d", *port),
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}
	log.Fatal(srv.ListenAndServe())
}

func (s *OAuthServer) handleDiscovery(w http.ResponseWriter, _ *http.Request) {
	config := map[string]interface{}{
		"issuer":                 s.Issuer,
		"authorization_endpoint": s.Issuer + "/authorize",
		"token_endpoint":         s.Issuer + "/token",
		"jwks_uri":               s.Issuer + "/jwks.json",
		"response_types_supported": []string{
			"code",
			"token",
			"id_token",
		},
		"subject_types_supported": []string{
			"public",
		},
		"id_token_signing_alg_values_supported": []string{
			"RS256",
		},
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(config)
}

func (s *OAuthServer) handleJWKS(w http.ResponseWriter, _ *http.Request) {
	jwk := jose.JSONWebKey{
		Key:       &s.Key.PublicKey,
		KeyID:     "test-key",
		Algorithm: "RS256",
		Use:       "sig",
	}
	jwks := jose.JSONWebKeySet{
		Keys: []jose.JSONWebKey{jwk},
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(jwks)
}

func (s *OAuthServer) handleAuthorize(w http.ResponseWriter, r *http.Request) {
	redirectURI := r.URL.Query().Get("redirect_uri")
	state := r.URL.Query().Get("state")
	code := "test-code"

	// Auto-approve and redirect
	target := fmt.Sprintf("%s?code=%s&state=%s", redirectURI, code, state)
	http.Redirect(w, r, target, http.StatusFound)
}

func (s *OAuthServer) handleToken(w http.ResponseWriter, r *http.Request) {
	// Generate tokens
	// ... (simplified for mock)
	// We return a simple access token.
	// In a real OIDC server, we'd sign an ID token.

	// Helper to generate dummy ID token
	// ...

	// For now, just return JSON
	w.Header().Set("Content-Type", "application/json")
	// If it's a code exchange
	if r.FormValue("grant_type") == "authorization_code" {
		// Return tokens
		// Check client_id/secret if needed
		log.Println("Handling authorization_code grant")
	}

	// Generate a dummy JWT ID Token
	idToken := s.generateIDToken()

	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"access_token": "mock-access-token",
		"token_type":   "Bearer",
		"expires_in":   3600,
		"id_token":     idToken,
	})
}

func (s *OAuthServer) generateIDToken() string {
	signer, err := jose.NewSigner(jose.SigningKey{Algorithm: jose.RS256, Key: s.Key}, nil)
	if err != nil {
		log.Printf("Failed to create signer: %v", err)
		return ""
	}

	claims := map[string]interface{}{
		"sub": "test-user",
		"iss": s.Issuer,
		"aud": "test-client",
		"exp": time.Now().Add(time.Hour).Unix(),
		"iat": time.Now().Unix(),
	}

	payload, err := json.Marshal(claims)
	if err != nil {
		log.Printf("Failed to marshal claims: %v", err)
		return ""
	}

	object, err := signer.Sign(payload)
	if err != nil {
		log.Printf("Failed to sign: %v", err)
		return ""
	}

	serialized, err := object.CompactSerialize()
	if err != nil {
		log.Printf("Failed to serialize: %v", err)
		return ""
	}
	return serialized
}

// publicKeyToPEM converts a public key to PEM format.
func publicKeyToPEM(pubkey *rsa.PublicKey) []byte {
	pubASN1, err := x509.MarshalPKIXPublicKey(pubkey)
	if err != nil {
		return nil
	}
	return pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: pubASN1,
	})
}
