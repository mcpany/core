// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
)

// ssoContextKey is a custom type for context keys to avoid collisions.
type ssoContextKey string

const (
	// UserIDContextKey is the key used to store the user ID in the context.
	UserIDContextKey ssoContextKey = "UserID"
)

// SSOConfig defines the SSO configuration.
//
// Summary: Configuration options for Single Sign-On (SSO) middleware.
type SSOConfig struct {
	Enabled bool
	IDPURL  string
}

// SSOMiddleware creates a new SSO middleware using standard net/http.
//
// Summary: Middleware that enforces SSO authentication via trusted headers or bearer tokens.
//
// Parameters:
//   - config: SSOConfig. The configuration settings for SSO.
//
// Returns:
//   - func(http.Handler) http.Handler: The standard HTTP middleware handler.
//
// Side Effects:
//   - Inspects headers for authentication information.
//   - Aborts the request with 401 Unauthorized if authentication is missing or invalid.
//   - Sets "UserID" in the context on successful authentication.
func SSOMiddleware(config SSOConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !config.Enabled {
				next.ServeHTTP(w, r)
				return
			}

			// Check for Identity Header (Trusted Proxy pattern)
			userID := r.Header.Get("X-MCP-Identity")
			if userID != "" {
				ctx := context.WithValue(r.Context(), UserIDContextKey, userID)
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}

			// Check for Bearer Token
			auth := r.Header.Get("Authorization")
			if strings.HasPrefix(auth, "Bearer ") {
				// Validate token (Mock validation)
				token := strings.TrimPrefix(auth, "Bearer ")
				if token == "valid-mock-token" {
					ctx := context.WithValue(r.Context(), UserIDContextKey, "user-123")
					next.ServeHTTP(w, r.WithContext(ctx))
					return
				}
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			_ = json.NewEncoder(w).Encode(map[string]string{
				"error":     "Authentication required",
				"login_url": config.IDPURL + "/login",
			})
		})
	}
}
