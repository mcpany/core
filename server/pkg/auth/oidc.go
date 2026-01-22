// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"net/http"

	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"
)

// OIDCConfig holds the configuration for the OIDC provider.
type OIDCConfig struct {
	Issuer       string
	ClientID     string
	ClientSecret string
	RedirectURL  string
}

// OIDCProvider handles OIDC authentication flow.
type OIDCProvider struct {
	config       OIDCConfig
	provider     *oidc.Provider
	verifier     *oidc.IDTokenVerifier
	oauth2Config oauth2.Config
}

// NewOIDCProvider creates a new OIDCProvider.
//
// ctx is the context for the request.
// config holds the configuration settings.
//
// Returns the result.
// Returns an error if the operation fails.
func NewOIDCProvider(ctx context.Context, config OIDCConfig) (*OIDCProvider, error) {
	provider, err := oidc.NewProvider(ctx, config.Issuer)
	if err != nil {
		return nil, err
	}

	verifier := provider.Verifier(&oidc.Config{ClientID: config.ClientID})

	oauth2Config := oauth2.Config{
		ClientID:     config.ClientID,
		ClientSecret: config.ClientSecret,
		RedirectURL:  config.RedirectURL,
		Endpoint:     provider.Endpoint(),
		Scopes:       []string{oidc.ScopeOpenID, "profile", "email"},
	}

	return &OIDCProvider{
		config:       config,
		provider:     provider,
		verifier:     verifier,
		oauth2Config: oauth2Config,
	}, nil
}

// HandleLogin initiates the OIDC login flow.
//
// w is the HTTP response writer.
// r is the HTTP request.
func (p *OIDCProvider) HandleLogin(w http.ResponseWriter, r *http.Request) {
	state, err := generateRandomState()
	if err != nil {
		http.Error(w, "Failed to generate state", http.StatusInternalServerError)
		return
	}

	// In a real app, store state in a secure cookie to verify later
	isSecure := r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https"
	http.SetCookie(w, &http.Cookie{
		Name:     "oauth_state",
		Value:    state,
		HttpOnly: true,
		Secure:   isSecure,
		SameSite: http.SameSiteLaxMode,
		Path:     "/",
		MaxAge:   300, // 5 minutes
	})

	http.Redirect(w, r, p.oauth2Config.AuthCodeURL(state), http.StatusFound)
}

// HandleCallback handles the OIDC provider callback.
//
// w is the HTTP response writer.
// r is the HTTP request.
func (p *OIDCProvider) HandleCallback(w http.ResponseWriter, r *http.Request) {
	stateCookie, err := r.Cookie("oauth_state")
	if err != nil {
		http.Error(w, "State cookie missing", http.StatusBadRequest)
		return
	}

	if r.URL.Query().Get("state") != stateCookie.Value {
		http.Error(w, "State invalid", http.StatusBadRequest)
		return
	}

	// Delete state cookie to prevent reuse and cleanup
	http.SetCookie(w, &http.Cookie{
		Name:     "oauth_state",
		Value:    "",
		HttpOnly: true,
		Secure:   r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https",
		SameSite: http.SameSiteLaxMode,
		Path:     "/",
		MaxAge:   -1,
	})

	oauth2Token, err := p.oauth2Config.Exchange(r.Context(), r.URL.Query().Get("code"))
	if err != nil {
		http.Error(w, "Failed to exchange token: "+err.Error(), http.StatusInternalServerError)
		return
	}

	rawIDToken, ok := oauth2Token.Extra("id_token").(string)
	if !ok {
		http.Error(w, "No id_token field in oauth2 token", http.StatusInternalServerError)
		return
	}

	idToken, err := p.verifier.Verify(r.Context(), rawIDToken)
	if err != nil {
		http.Error(w, "Failed to verify ID Token: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Extract claims
	var claims struct {
		Email string `json:"email"`
		Sub   string `json:"sub"`
	}
	if err := idToken.Claims(&claims); err != nil {
		http.Error(w, "Failed to parse claims: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Create session (in memory or JWT) - Simplified for MVP
	// We'll set a cookie with the user's email signed or just plain for now (INSECURE for prod, use session store)
	// For this MVP, we will assume we have a session manager.
	// We'll return a simple JSON response with the user info.

	resp := map[string]string{
		"message": "Authenticated",
		"email":   claims.Email,
		"user_id": claims.Sub,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

func generateRandomState() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}
