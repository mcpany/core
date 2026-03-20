// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
)

type ssoContextKey string

const ssoUserIDKey ssoContextKey = "UserID"

// GetUserIDFromContext retrieves the user ID from the request context.
func GetUserIDFromContext(ctx context.Context) string {
	if val, ok := ctx.Value(ssoUserIDKey).(string); ok {
		return val
	}
	return ""
}

// SSOConfig defines the SSO configuration.
//
// Summary: Configuration options for Single Sign-On (SSO) middleware.
type SSOConfig struct {
	Enabled bool
	IDPURL  string
}

// SSOMiddleware creates a new SSO middleware.
//
// Summary: Middleware that enforces SSO authentication via trusted headers or bearer tokens.
//
// Parameters:
//   - config: SSOConfig. The configuration settings for SSO.
//
// Returns:
//   - func(http.Handler) http.Handler: The net/http middleware handler.
//
// Side Effects:
//   - Inspects headers for authentication information.
//   - Aborts the request with 401 Unauthorized if authentication is missing or invalid.
//   - Sets User in the context on successful authentication.
func SSOMiddleware(config SSOConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !config.Enabled {
				next.ServeHTTP(w, r)
				return
			}

			var userID string

			// Check for Identity Header (Trusted Proxy pattern)
			identityHeader := r.Header.Get("X-MCP-Identity")
			if identityHeader != "" {
				userID = identityHeader
			} else {
				// Check for Bearer Token
				authHeader := r.Header.Get("Authorization")
				if strings.HasPrefix(authHeader, "Bearer ") {
					// Validate token (Mock validation)
					token := strings.TrimPrefix(authHeader, "Bearer ")
					if token == "valid-mock-token" {
						userID = "user-123"
					}
				}
			}

			if userID != "" {
				ctx := context.WithValue(r.Context(), ssoUserIDKey, userID)
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)

			resp := map[string]string{
				"error":     "Authentication required",
				"login_url": config.IDPURL + "/login",
			}
			json.NewEncoder(w).Encode(resp)
		})
	}
}
