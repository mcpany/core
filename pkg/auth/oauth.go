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
	"fmt"
	"net/http"
	"strings"

	"github.com/coreos/go-oidc/v3/oidc"
)

// OAuth2Authenticator implements the Authenticator interface for OAuth2-based
// authentication using OpenID Connect (OIDC). It validates JWTs (JSON Web
// Tokens) presented in the HTTP Authorization header.
type OAuth2Authenticator struct {
	verifier *oidc.IDTokenVerifier
}

// NewOAuth2Authenticator creates a new OAuth2Authenticator with the provided
// configuration. It initializes the OIDC provider and creates a verifier for
// validating ID tokens.
//
// ctx is the context for the OIDC provider initialization.
// config holds the OAuth2 configuration, including the issuer URL and client ID.
//
// It returns a new OAuth2Authenticator or an error if the OIDC provider cannot
// be initialized.
func NewOAuth2Authenticator(ctx context.Context, config *OAuth2Config) (*OAuth2Authenticator, error) {
	provider, err := oidc.NewProvider(ctx, config.IssuerURL)
	if err != nil {
		return nil, fmt.Errorf("failed to create OIDC provider: %w", err)
	}

	oidcConfig := &oidc.Config{
		ClientID: config.Audience,
	}

	return &OAuth2Authenticator{
		verifier: provider.Verifier(oidcConfig),
	}, nil
}

// Authenticate validates the JWT from the Authorization header of the request.
// It checks for a "Bearer" token and verifies its signature, expiration, and
// claims against the OIDC provider.
//
// ctx is the request context.
// r is the HTTP request to authenticate.
//
// It returns the context with the user's identity on success, or an error if
// authentication fails.
func (a *OAuth2Authenticator) Authenticate(ctx context.Context, r *http.Request) (context.Context, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return ctx, fmt.Errorf("missing Authorization header")
	}

	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		return ctx, fmt.Errorf("invalid Authorization header format")
	}
	token := parts[1]

	idToken, err := a.verifier.Verify(ctx, token)
	if err != nil {
		return ctx, fmt.Errorf("failed to verify token: %w", err)
	}

	var claims struct {
		Email string `json:"email"`
		// Add other claims as needed
	}
	if err := idToken.Claims(&claims); err != nil {
		return ctx, fmt.Errorf("failed to extract claims: %w", err)
	}

	return context.WithValue(ctx, "user", claims.Email), nil
}
