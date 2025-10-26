package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"gopkg.in/square/go-jose.v2"
	"gopkg.in/square/go-jose.v2/jwt"
)

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, "OK")
}

func timeHandler(w http.ResponseWriter, r *http.Request) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		http.Error(w, "Missing Authorization header", http.StatusUnauthorized)
		return
	}

	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		http.Error(w, "Invalid Authorization header", http.StatusUnauthorized)
		return
	}
	tokenString := parts[1]

	// In a real application, you would use a proper OIDC library to validate
	// the token. For this example, we'll just decode it and check the issuer.
	tok, err := jwt.ParseSigned(tokenString)
	if err != nil {
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}

	// This is a very basic validation. A real application should verify the
	// signature against the provider's public key.
	claims := &jwt.Claims{}
	if err := tok.UnsafeClaimsWithoutVerification(claims); err != nil {
		http.Error(w, "Invalid claims", http.StatusUnauthorized)
		return
	}

	issuerURL := os.Getenv("AUTH_ISSUER_URL")
	if issuerURL == "" {
		log.Println("AUTH_ISSUER_URL not set, skipping issuer validation")
	} else if claims.Issuer != issuerURL {
		http.Error(w, "Invalid issuer", http.StatusUnauthorized)
		return
	}

	response := map[string]string{
		"current_time": time.Now().Format(time.RFC3339),
		"timezone":     "UTC",
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

func main() {
	port := os.Getenv("HTTP_PORT")
	if port == "" {
		port = "8081"
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/health", healthHandler)
	mux.HandleFunc("/time", timeHandler)

	log.Printf("Server starting on port %s", port)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
