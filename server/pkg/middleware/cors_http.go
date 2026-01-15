// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"net/http"
	"sync"

	"github.com/mcpany/core/server/pkg/logging"
)

// HTTPCORSMiddleware handles CORS for HTTP endpoints.
// It is thread-safe and supports dynamic updates.
type HTTPCORSMiddleware struct {
	mu             sync.RWMutex
	allowedOrigins []string
}

// NewHTTPCORSMiddleware creates a new HTTPCORSMiddleware.
// If allowedOrigins is empty, it defaults to allowing nothing (or behaving like standard Same-Origin).
// To allow all, pass []string{"*"}.
func NewHTTPCORSMiddleware(allowedOrigins []string) *HTTPCORSMiddleware {
	for _, o := range allowedOrigins {
		if o == "*" {
			logging.GetLogger().Warn("⚠️  CORS configured with wildcard origin '*'. This allows any website to access your API. Ensure this is intended.")
		}
	}
	return &HTTPCORSMiddleware{
		allowedOrigins: allowedOrigins,
	}
}

// Update updates the allowed origins.
func (m *HTTPCORSMiddleware) Update(allowedOrigins []string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, o := range allowedOrigins {
		if o == "*" {
			logging.GetLogger().Warn("⚠️  CORS updated with wildcard origin '*'. This allows any website to access your API.")
		}
	}
	m.allowedOrigins = allowedOrigins
}

// Handler wraps an http.Handler with CORS logic.
func (m *HTTPCORSMiddleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		m.mu.RLock()
		origins := m.allowedOrigins
		m.mu.RUnlock()

		origin := r.Header.Get("Origin")
		if origin == "" {
			// Not a CORS request
			next.ServeHTTP(w, r)
			return
		}

		allowed := false
		wildcardAllowed := false

		for _, o := range origins {
			if o == origin {
				allowed = true
				break
			}
			if o == "*" {
				wildcardAllowed = true
				// Don't break, keep looking for exact match
			}
		}

		if !allowed && !wildcardAllowed {
			// CORS check failed
			next.ServeHTTP(w, r)
			return
		}

		// Set CORS headers
		if allowed {
			// Exact match: Reflect origin and allow credentials
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			w.Header().Set("Vary", "Origin")
		} else {
			// Wildcard match: Return "*" and NO credentials
			logging.GetLogger().Debug("CORS: Allowing wildcard origin", "origin", origin, "source", "HTTPCORSMiddleware")
			w.Header().Set("Access-Control-Allow-Origin", "*")
			// No Access-Control-Allow-Credentials
		}

		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-API-Key, X-Requested-With, x-grpc-web, grpc-timeout, x-user-agent")
		w.Header().Set("Access-Control-Expose-Headers", "grpc-status, grpc-message, Date, Content-Length, Content-Type")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}
