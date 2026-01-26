// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"net/http"
	"net/url"
	"strings"
	"sync"
)

// CSRFMiddleware protects against Cross-Site Request Forgery.
type CSRFMiddleware struct {
	allowedOrigins map[string]bool
	allowAll       bool
	mu             sync.RWMutex
}

// NewCSRFMiddleware creates a new CSRFMiddleware.
func NewCSRFMiddleware(allowedOrigins []string) *CSRFMiddleware {
	m := &CSRFMiddleware{}
	m.Update(allowedOrigins)
	return m
}

// Update updates the allowed origins.
func (m *CSRFMiddleware) Update(allowedOrigins []string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.allowedOrigins = make(map[string]bool)
	m.allowAll = false

	for _, o := range allowedOrigins {
		if o == "*" {
			m.allowAll = true
		}
		m.allowedOrigins[o] = true
	}
}

// Handler returns the HTTP handler.
func (m *CSRFMiddleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 1. Safe Methods are allowed
		if r.Method == http.MethodGet || r.Method == http.MethodHead || r.Method == http.MethodOptions || r.Method == http.MethodTrace {
			next.ServeHTTP(w, r)
			return
		}

		// 2. Custom Auth Headers imply consent (preflighted or non-browser)
		if r.Header.Get("Authorization") != "" || r.Header.Get("X-API-Key") != "" {
			next.ServeHTTP(w, r)
			return
		}

		// 3. Check Origin / Referer
		origin := r.Header.Get("Origin")
		referer := r.Header.Get("Referer")

		if origin == "" && referer == "" {
			// No Origin/Referer implies non-browser tool (e.g. curl)
			// We allow this to support CLI tools against localhost without auth headers (if IP allowlist permits)
			next.ServeHTTP(w, r)
			return
		}

		// Validate Origin if present
		if origin != "" {
			// Check Same Origin
			if isSameOrigin(origin, r.Host) {
				next.ServeHTTP(w, r)
				return
			}
			if !m.isAllowed(origin) {
				http.Error(w, "CSRF: Invalid Origin", http.StatusForbidden)
				return
			}
		} else if referer != "" {
			// Fallback to Referer if Origin is missing
			// We need to parse the origin from the referer URL
			u, err := url.Parse(referer)
			if err != nil {
				http.Error(w, "CSRF: Invalid Referer", http.StatusForbidden)
				return
			}
			// Reconstruct origin: scheme://host
			refererOrigin := u.Scheme + "://" + u.Host

			// Check Same Origin for Referer
			if isSameOrigin(refererOrigin, r.Host) {
				next.ServeHTTP(w, r)
				return
			}

			if !m.isAllowed(refererOrigin) {
				http.Error(w, "CSRF: Invalid Referer Origin", http.StatusForbidden)
				return
			}
		}

		next.ServeHTTP(w, r)
	})
}

func (m *CSRFMiddleware) isAllowed(origin string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.allowAll {
		return true
	}
	return m.allowedOrigins[origin]
}

// isSameOrigin checks if the origin matches the host.
// origin is expected to be "scheme://host[:port]".
// host is expected to be "host[:port]".
func isSameOrigin(origin, host string) bool {
	// Strip scheme from origin
	if idx := strings.Index(origin, "://"); idx != -1 {
		originHost := origin[idx+3:]
		return strings.EqualFold(originHost, host)
	}
	return false
}
