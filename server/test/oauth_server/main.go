// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

// Package main implements a mock OAuth server for testing.
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

// OAuthServer mock.
type OAuthServer struct {
	Issuer string
	Key    *rsa.PrivateKey
}

func main() {
	port := flag.Int("port", 8081, "Port to listen on")
	flag.Parse()

	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		log.Fatal(err)
	}

	server := &OAuthServer{
		Issuer: fmt.Sprintf("http://localhost:%d", *port),
		Key:    key,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	})
	mux.HandleFunc("/.well-known/openid-configuration", server.handleDiscovery)
	mux.HandleFunc("/jwks.json", server.handleJWKS)
	mux.HandleFunc("/authorize", server.handleAuthorize)
	mux.HandleFunc("/token", server.handleToken)

	log.Printf("Starting mock OAuth server on port %d", *port)
	serverInstance := &http.Server{
		Addr:              fmt.Sprintf(":%d", *port),
		Handler:           mux,
		ReadHeaderTimeout: 3 * time.Second,
	}
	log.Fatal(serverInstance.ListenAndServe())
}

func (s *OAuthServer) handleDiscovery(w http.ResponseWriter, _ *http.Request) {
	config := map[string]interface{}{
		"issuer":                 s.Issuer,
		"authorization_endpoint": s.Issuer + "/authorize",
		"token_endpoint":         s.Issuer + "/token",
		"jwks_uri":               s.Issuer + "/jwks.json",
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(config); err != nil {
		log.Printf("Failed to encode discovery config: %v", err)
	}
}

func (s *OAuthServer) handleJWKS(w http.ResponseWriter, _ *http.Request) {
	jwk := jose.JSONWebKey{
		Key:       &s.Key.PublicKey,
		KeyID:     "key1",
		Algorithm: string(jose.RS256),
		Use:       "sig",
	}
	jwks := jose.JSONWebKeySet{
		Keys: []jose.JSONWebKey{jwk},
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(jwks); err != nil {
		log.Printf("Failed to encode JWKS: %v", err)
	}
}

func (s *OAuthServer) handleAuthorize(w http.ResponseWriter, r *http.Request) {
	redirectURI := r.URL.Query().Get("redirect_uri")
	state := r.URL.Query().Get("state")
	code := "mock-code"

	// Auto-redirect
	http.Redirect(w, r, fmt.Sprintf("%s?code=%s&state=%s", redirectURI, code, state), http.StatusFound)
}

func (s *OAuthServer) handleToken(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]interface{}{
		"access_token": "mock-access-token",
		"token_type":   "Bearer",
		"expires_in":   3600,
		"id_token":     "mock-id-token", // Simplified
	}); err != nil {
		log.Printf("Failed to encode token response: %v", err)
	}
}
