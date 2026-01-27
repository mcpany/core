// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

// Package main implements a mock OAuth2 server for testing purposes.
package main

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/go-jose/go-jose/v4"
)

var (
	port = flag.Int("port", 8080, "Port to listen on")
)

// OAuthServer represents a mock OAuth2 server.
type OAuthServer struct {
	privateKey *rsa.PrivateKey
	signer     jose.Signer
}

func main() {
	flag.Parse()

	// Generate RSA key for signing tokens
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		log.Fatalf("Failed to generate private key: %v", err)
	}

	// Create signer
	signer, err := jose.NewSigner(jose.SigningKey{Algorithm: jose.RS256, Key: privateKey}, nil)
	if err != nil {
		log.Fatalf("Failed to create signer: %v", err)
	}

	server := &OAuthServer{
		privateKey: privateKey,
		signer:     signer,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("OK"))
	})
	mux.HandleFunc("/.well-known/openid-configuration", server.handleDiscovery)
	mux.HandleFunc("/jwks.json", server.handleJWKS)
	mux.HandleFunc("/token", server.handleToken)

	s := &http.Server{
		Addr:              fmt.Sprintf(":%d", *port),
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}
	log.Printf("Starting OAuth server on port %d", *port)
	log.Fatal(s.ListenAndServe())
}

func (s *OAuthServer) handleDiscovery(w http.ResponseWriter, r *http.Request) {
	baseURL := fmt.Sprintf("http://%s", r.Host)
	config := map[string]interface{}{
		"issuer":                 baseURL,
		"authorization_endpoint": baseURL + "/authorize",
		"token_endpoint":         baseURL + "/token",
		"jwks_uri":               baseURL + "/jwks.json",
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
		Key:       &s.privateKey.PublicKey,
		KeyID:     "test-key-id",
		Algorithm: string(jose.RS256),
		Use:       "sig",
	}

	jwks := jose.JSONWebKeySet{
		Keys: []jose.JSONWebKey{jwk},
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(jwks)
}

func (s *OAuthServer) handleToken(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	// Simple mock: always return a token
	// In a real test, verify client_id, client_secret, grant_type, etc.

	// Create JWT
	claims := map[string]interface{}{
		"sub":   "test-user",
		"iss":   fmt.Sprintf("http://%s", r.Host),
		"aud":   "test-client",
		"exp":   time.Now().Add(1 * time.Hour).Unix(),
		"iat":   time.Now().Unix(),
		"email": "test@example.com",
	}

	payload, err := json.Marshal(claims)
	if err != nil {
		http.Error(w, "Failed to marshal claims", http.StatusInternalServerError)
		return
	}

	object, err := s.signer.Sign(payload)
	if err != nil {
		http.Error(w, "Failed to sign token", http.StatusInternalServerError)
		return
	}

	token, err := object.CompactSerialize()
	if err != nil {
		http.Error(w, "Failed to serialize token", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	// Using fmt.Fprintf instead of w.Write for string formatting if needed, but here simple JSON
	// The linter complained about unchecked error on fmt.Fprintf
	// Actually we are returning JSON.

	resp := map[string]interface{}{
		"access_token": token,
		"token_type":   "Bearer",
		"expires_in":   3600,
		"id_token":     token, // For OIDC
	}
	_ = json.NewEncoder(w).Encode(resp)
}
