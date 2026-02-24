// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"net/http"
	"sync"

	"github.com/mcpany/core/server/pkg/logging"
)

// HTTPCORSMiddleware handles Cross-Origin Resource Sharing (CORS) for HTTP endpoints.
// It is thread-safe and supports dynamic updates of allowed origins.
type HTTPCORSMiddleware struct {
	mu              sync.RWMutex
	allowedOrigins  map[string]struct{}
	wildcardAllowed bool
}

// NewHTTPCORSMiddleware creates a new HTTPCORSMiddleware instance.
//
// Parameters:
//   - allowedOrigins ([]string): A list of allowed origins (e.g., "http://localhost:3000").
//     Pass []string{"*"} to allow all origins.
//
// Returns:
//   - (*HTTPCORSMiddleware): The initialized middleware.
//
// Errors:
//   - None.
//
// Side Effects:
//   - Initializes internal data structures for O(1) origin lookup.
func NewHTTPCORSMiddleware(allowedOrigins []string) *HTTPCORSMiddleware {
	m := &HTTPCORSMiddleware{}
	m.updateInternal(allowedOrigins)
	return m
}

// Update updates the list of allowed origins dynamically.
//
// Parameters:
//   - allowedOrigins ([]string): The new list of allowed origins.
//
// Returns:
//   - None.
//
// Errors:
//   - None.
//
// Side Effects:
//   - Replaces the internal map of allowed origins.
//   - Acquires a write lock on the middleware.
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
// Parameters:
//   - next (http.Handler): The next handler in the chain.
//
// Returns:
//   - (http.Handler): The wrapped handler that adds CORS headers.
//
// Errors:
//   - None.
//
// Side Effects:
//   - Sets Access-Control-Allow-Origin, Access-Control-Allow-Credentials, and other CORS headers on the response.
//   - Intercepts OPTIONS requests and returns 200 OK immediately.
//   - Logs warnings if a wildcard origin is used (security risk).
func (m *HTTPCORSMiddleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		// If no Origin header is present, it's not a CORS request (or it's a same-origin request).
		// Pass through without modifying headers, unless we want to enforce CORS on everything?
		// Standard CORS middleware usually passes through if no Origin.
		if origin == "" {
			next.ServeHTTP(w, r)
			return
		}

		m.mu.RLock()
		_, allowed := m.allowedOrigins[origin]
		wildcardAllowed := m.wildcardAllowed
		m.mu.RUnlock()

		if !allowed && !wildcardAllowed {
			// Origin not allowed.
			// Do NOT set CORS headers. Browser will block the response.
			// We continue processing the request, but the browser won't let the client see the response.
			// Alternatively, we could abort here. But standard behavior is to just omit headers.
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
			// Wildcard match: Return "*" and NO credentials (spec requirement)
			logging.GetLogger().Warn("CORS: Allowing wildcard origin", "origin", origin, "source", "HTTPCORSMiddleware")
			w.Header().Set("Access-Control-Allow-Origin", "*")
			// No Access-Control-Allow-Credentials for wildcard
		}

		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-API-Key, X-Requested-With, x-grpc-web, grpc-timeout, x-user-agent")
		w.Header().Set("Access-Control-Expose-Headers", "grpc-status, grpc-message, Date, Content-Length, Content-Type")

		// Handle preflight requests
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}
