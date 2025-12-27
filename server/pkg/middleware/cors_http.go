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
		allowAny := false
		for _, o := range m.allowedOrigins {
			if o == "*" {
				allowed = true
				allowAny = true
				break
			}
			if o == origin {
				allowed = true
				break
			}
		}

		if !allowed {
			// CORS check failed
			next.ServeHTTP(w, r)
			return
		}

		// Set CORS headers
		if allowAny {
			// If "*" is configured, we set Origin to "*" and do NOT allow credentials.
			// This is the secure configuration for public APIs.
			// If you need credentials, you must explicitly list the origin.
			w.Header().Set("Access-Control-Allow-Origin", "*")
			// Remove Credentials header if it was set (though it shouldn't be yet)
			w.Header().Del("Access-Control-Allow-Credentials")
		} else if allowed {
			// Specific origin match
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Vary", "Origin")
			w.Header().Set("Access-Control-Allow-Credentials", "true")
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
