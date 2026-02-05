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
	mu              sync.RWMutex
	allowedOrigins  map[string]struct{}
	wildcardAllowed bool
}

// NewHTTPCORSMiddleware creates a new HTTPCORSMiddleware.
//
// Summary: Initializes a new CORS middleware for HTTP endpoints with the specified allowed origins.
//
// Parameters:
//   - allowedOrigins: []string. A list of allowed origins (e.g., "https://example.com" or "*").
//
// Returns:
//   - *HTTPCORSMiddleware: The initialized HTTP CORS middleware.
func NewHTTPCORSMiddleware(allowedOrigins []string) *HTTPCORSMiddleware {
	m := &HTTPCORSMiddleware{}
	m.updateInternal(allowedOrigins)
	return m
}

// Update updates the allowed origins.
//
// Summary: Dynamically updates the list of allowed CORS origins.
//
// Parameters:
//   - allowedOrigins: []string. A list of allowed origins.
func (m *HTTPCORSMiddleware) Update(allowedOrigins []string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.updateInternal(allowedOrigins)
}

// updateInternal populates the internal map and flags.
// It must be called with the lock held or during initialization.
// ⚡ Bolt Optimization: Uses map for O(1) lookup instead of O(N) slice iteration.
func (m *HTTPCORSMiddleware) updateInternal(origins []string) {
	m.allowedOrigins = make(map[string]struct{}, len(origins))
	m.wildcardAllowed = false
	for _, o := range origins {
		if o == "*" {
			m.wildcardAllowed = true
		} else {
			m.allowedOrigins[o] = struct{}{}
		}
	}
}

// Handler wraps an http.Handler with CORS logic.
//
// Summary: Middleware that adds CORS headers to responses and handles preflight requests.
//
// Parameters:
//   - next: http.Handler. The next handler in the chain.
//
// Returns:
//   - http.Handler: The wrapped handler with CORS support.
func (m *HTTPCORSMiddleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin == "" {
			// Not a CORS request
			next.ServeHTTP(w, r)
			return
		}

		m.mu.RLock()
		// Check for exact match first
		_, allowed := m.allowedOrigins[origin]
		wildcardAllowed := m.wildcardAllowed
		m.mu.RUnlock()

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
			logging.GetLogger().Warn("CORS: Allowing wildcard origin", "origin", origin, "source", "HTTPCORSMiddleware")
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
