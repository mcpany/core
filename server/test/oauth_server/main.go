// Package main provides a mock OAuth 2.0 Identity Provider for testing purposes.

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
	"github.com/golang-jwt/jwt/v5"
)

var (
	port = flag.Int("port", 8085, "Port to listen on")
)

// OAuthServer mocks an OAuth 2.0 Identity Provider.
type OAuthServer struct {
	// PrivateKey is the RSA private key used for signing tokens.
	PrivateKey *rsa.PrivateKey
	// BaseURL is the base URL of the mock server.
	BaseURL string
}

func main() {
	flag.Parse()

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		log.Fatalf("Failed to generate key: %v", err)
	}

	server := &OAuthServer{
		PrivateKey: privateKey,
		BaseURL:    fmt.Sprintf("http://localhost:%d", *port),
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/.well-known/openid-configuration", server.handleDiscovery)
	mux.HandleFunc("/jwks", server.handleJWKS)
	mux.HandleFunc("/auth", server.handleAuth)
	mux.HandleFunc("/token", server.handleToken)
	mux.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	})

	log.Printf("Starting Mock OAuth Server on port %d...", *port)
	httpServer := &http.Server{
		Addr:              fmt.Sprintf(":%d", *port),
		Handler:           mux,
		ReadHeaderTimeout: 3 * time.Second,
	}
	log.Fatal(httpServer.ListenAndServe())
}

func (s *OAuthServer) handleDiscovery(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	config := map[string]interface{}{
		"issuer":                 s.BaseURL,
		"jwks_uri":               s.BaseURL + "/jwks",
		"authorization_endpoint": s.BaseURL + "/auth",
		"token_endpoint":         s.BaseURL + "/token",
	}
	_ = json.NewEncoder(w).Encode(config)
}

func (s *OAuthServer) handleJWKS(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	jwk := jose.JSONWebKey{Key: &s.PrivateKey.PublicKey, Algorithm: "RS256", Use: "sig"}
	jwks := map[string]interface{}{
		"keys": []interface{}{jwk},
	}
	_ = json.NewEncoder(w).Encode(jwks)
}

func (s *OAuthServer) handleAuth(w http.ResponseWriter, r *http.Request) {
	state := r.URL.Query().Get("state")
	redirectURI := r.URL.Query().Get("redirect_uri")
	if redirectURI == "" {
		http.Error(w, "missing redirect_uri", http.StatusBadRequest)
		return
	}

	// Create a simple login page for interaction
	w.Header().Set("Content-Type", "text/html")
	_, _ = fmt.Fprintf(w, `
		<html>
			<body>
				<h1>Mock OAuth Login</h1>
				<p>Click below to approve.</p>
				<form action="%s" method="get">
					<input type="hidden" name="code" value="mock_auth_code_123" />
					<input type="hidden" name="state" value="%s" />
					<button type="submit" id="approve-btn">Approve</button>
				</form>
			</body>
		</html>
	`, redirectURI, state)
}

func (s *OAuthServer) handleToken(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
		"iss": s.BaseURL,
		"sub": "test-user",
		"aud": "mcp-any-client",
		"exp": time.Now().Add(time.Hour).Unix(),
		"scope": "read write",
	})

	signedToken, err := token.SignedString(s.PrivateKey)
	if err != nil {
		http.Error(w, "failed to sign token", http.StatusInternalServerError)
		return
	}

	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"access_token": signedToken,
		"id_token":     signedToken,
		"token_type":   "Bearer",
		"expires_in":   3600,
		"scope":        "read write",
	})
}
