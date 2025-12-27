// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"net/http"
)

// HTTPCORSMiddleware handles CORS for HTTP endpoints.
type HTTPCORSMiddleware struct {
	allowedOrigins []string
}

// NewHTTPCORSMiddleware creates a new HTTPCORSMiddleware.
// If allowedOrigins is empty, it defaults to allowing nothing (or behaving like standard Same-Origin).
// To allow all, pass []string{"*"}.
func NewHTTPCORSMiddleware(allowedOrigins []string) *HTTPCORSMiddleware {
	return &HTTPCORSMiddleware{
		allowedOrigins: allowedOrigins,
	}
}

// Handler wraps an http.Handler with CORS logic.
func (m *HTTPCORSMiddleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin == "" {
			// Not a CORS request
			next.ServeHTTP(w, r)
			return
		}

		allowed := false
		wildcardAllowed := false

		for _, o := range m.allowedOrigins {
			if o == "*" {
				wildcardAllowed = true
				// Continue checking for specific match which might override wildcard behavior if we wanted,
				// but here we just note wildcard is present.
			}
			if o == origin {
				allowed = true
				break
			}
		}

		if !allowed && wildcardAllowed {
			// Wildcard match
			w.Header().Set("Access-Control-Allow-Origin", "*")
			// Do NOT set Access-Control-Allow-Credentials for wildcard
			// w.Header().Set("Access-Control-Allow-Credentials", "true")
		} else if allowed {
			// Specific match
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Vary", "Origin")
			w.Header().Set("Access-Control-Allow-Credentials", "true")
		} else {
			// CORS check failed
			next.ServeHTTP(w, r)
			return
		}

		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-API-Key, X-Requested-With")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}
