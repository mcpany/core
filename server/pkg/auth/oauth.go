// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package auth

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/coreos/go-oidc/v3/oidc"
)

// OAuth2Authenticator implements the Authenticator interface for OAuth2-based.
//
// Summary: implements the Authenticator interface for OAuth2-based.
type OAuth2Authenticator struct {
	verifier  *oidc.IDTokenVerifier
	audiences []string
}

// NewOAuth2Authenticator creates a new OAuth2Authenticator with the provided.
//
// Summary: creates a new OAuth2Authenticator with the provided.
//
// Parameters:
//   - ctx: context.Context. The context for the operation.
//   - config: *OAuth2Config. The config.
//
// Returns:
//   - *OAuth2Authenticator: The *OAuth2Authenticator.
//   - error: An error if the operation fails.
//
// Throws/Errors:
//   Returns an error if the operation fails.
func NewOAuth2Authenticator(ctx context.Context, config *OAuth2Config) (*OAuth2Authenticator, error) {
	provider, err := oidc.NewProvider(ctx, config.IssuerURL)
	if err != nil {
		return nil, fmt.Errorf("failed to create OIDC provider: %w", err)
	}

	oidcConfig := &oidc.Config{}
	audiences := config.Audiences

	// Backward compatibility
	if len(audiences) == 0 && config.Audience != "" {
		audiences = []string{config.Audience}
	}

	if len(audiences) == 1 {
		oidcConfig.ClientID = audiences[0]
	} else if len(audiences) > 1 {
		// If multiple audiences are allowed, we skip the ClientID check in the verifier
		// and perform it manually in Authenticate.
		oidcConfig.SkipClientIDCheck = true
	}

	return &OAuth2Authenticator{
		verifier:  provider.Verifier(oidcConfig),
		audiences: audiences,
	}, nil
}

// Authenticate validates the JWT from the Authorization header of the request.
//
// Summary: validates the JWT from the Authorization header of the request.
//
// Parameters:
//   - ctx: context.Context. The context for the operation.
//   - r: *http.Request. The r.
//
// Returns:
//   - context.Context: The context.Context.
//   - error: An error if the operation fails.
//
// Throws/Errors:
//   Returns an error if the operation fails.
func (a *OAuth2Authenticator) Authenticate(ctx context.Context, r *http.Request) (context.Context, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return ctx, fmt.Errorf("unauthorized")
	}

	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		return ctx, fmt.Errorf("unauthorized")
	}
	token := parts[1]

	idToken, err := a.verifier.Verify(ctx, token)
	if err != nil {
		return ctx, fmt.Errorf("unauthorized")
	}

	// Manual audience check if multiple audiences are configured
	if len(a.audiences) > 1 {
		matched := false
		for _, aud := range idToken.Audience {
			for _, allowedAud := range a.audiences {
				if aud == allowedAud {
					matched = true
					break
				}
			}
			if matched {
				break
			}
		}
		if !matched {
			return ctx, fmt.Errorf("unauthorized: audience mismatch")
		}
	}

	var claims struct {
		Email         string `json:"email"`
		EmailVerified bool   `json:"email_verified"`
		// Add other claims as needed
	}
	if err := idToken.Claims(&claims); err != nil {
		return ctx, fmt.Errorf("unauthorized")
	}

	if !claims.EmailVerified {
		return ctx, fmt.Errorf("unauthorized")
	}

	return context.WithValue(ctx, UserContextKey, claims.Email), nil
}
