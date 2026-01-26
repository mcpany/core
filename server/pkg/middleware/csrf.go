// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"

	"github.com/mcpany/core/server/pkg/logging"
)

// CSRFMiddleware protects against Cross-Site Request Forgery attacks.
type CSRFMiddleware struct {
	allowedOrigins map[string]bool
	mu             sync.RWMutex
}

// NewCSRFMiddleware creates a new CSRFMiddleware.
func NewCSRFMiddleware(allowedOrigins []string) *CSRFMiddleware {
	m := &CSRFMiddleware{
		allowedOrigins: make(map[string]bool),
	}
	m.Update(allowedOrigins)
	return m
}

// Update updates the allowed origins.
func (m *CSRFMiddleware) Update(origins []string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.allowedOrigins = make(map[string]bool)
	for _, o := range origins {
		m.allowedOrigins[strings.ToLower(o)] = true
	}
}

// Handler returns the HTTP handler.
func (m *CSRFMiddleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 1. Safe Methods are always allowed
		if r.Method == http.MethodGet ||
			r.Method == http.MethodHead ||
			r.Method == http.MethodOptions ||
			r.Method == http.MethodTrace {
			next.ServeHTTP(w, r)
			return
		}

		// 2. Custom Headers indicate non-simple request (preflighted) or intentional API access
		if r.Header.Get("X-API-Key") != "" ||
			r.Header.Get("X-Requested-With") != "" ||
			r.Header.Get("X-MCP-Any-CSRF") != "" ||
			strings.HasPrefix(r.Header.Get("Authorization"), "Bearer ") {
			next.ServeHTTP(w, r)
			return
		}

		// 3. Content-Type: application/json
		// Although simple requests can't send this without preflight (mostly),
		// we trust it as a signal of API intent.
		// Note: We check the start because it might include charset.
		contentType := strings.ToLower(r.Header.Get("Content-Type"))
		if strings.HasPrefix(contentType, "application/json") {
			next.ServeHTTP(w, r)
			return
		}

		// 4. Origin/Referer Verification
		// If we are here, it's a state-changing request without custom headers and not JSON.
		// This could be a form submission or a simple fetch/xhr.
		origin := r.Header.Get("Origin")
		referer := r.Header.Get("Referer")

		if origin == "" && referer == "" {
			// If both are missing, it's likely not a browser, or privacy tools are stripping headers.
			// In a strict mode we might block, but for now we log and allow?
			// Blocking is safer for CSRF. Non-browser tools usually set headers if required.
			// But curl doesn't set Origin.
			// If it's curl, it likely doesn't have cookies/basic-auth cached from a browser session.
			// So CSRF risk is low if we assume CSRF targets browser sessions.
			// Let's allow if no Origin/Referer, assuming it's a CLI/script.
			// But attacker can suppress Referer? Origin is harder to suppress in browser.
			// Modern browsers send Origin for POST.
			next.ServeHTTP(w, r)
			return
		}

		// Check Origin
		if origin != "" {
			if !m.isOriginAllowed(origin) {
				logging.GetLogger().Warn("CSRF blocked: Origin not allowed", "origin", origin, "path", r.URL.Path)
				http.Error(w, "Forbidden: CSRF Origin Check Failed", http.StatusForbidden)
				return
			}
		} else if referer != "" {
			// Check Referer
			u, err := url.Parse(referer)
			if err != nil {
				logging.GetLogger().Warn("CSRF blocked: Invalid Referer", "referer", referer, "error", err)
				http.Error(w, "Forbidden: CSRF Referer Check Failed", http.StatusForbidden)
				return
			}
			// Reconstruct origin from referer (scheme://host)
			refOrigin := fmt.Sprintf("%s://%s", u.Scheme, u.Host)
			if !m.isOriginAllowed(refOrigin) {
				logging.GetLogger().Warn("CSRF blocked: Referer Origin not allowed", "referer", referer, "extracted_origin", refOrigin)
				http.Error(w, "Forbidden: CSRF Referer Check Failed", http.StatusForbidden)
				return
			}
		}

		next.ServeHTTP(w, r)
	})
}

func (m *CSRFMiddleware) isOriginAllowed(origin string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	origin = strings.ToLower(origin)
	if m.allowedOrigins[origin] {
		return true
	}

	// Also allow exact match on localhost variants if strictly configured?
	// Usually allowedOrigins contains them.
	// But if the user didn't configure anything, allowedOrigins might be empty?
	// If empty, we should arguably fail or allow same-origin?
	// But we don't know "same-origin" easily without knowing the server's own public URL.
	// However, we can check if it matches "null" (local files, blocked usually) or generic checks.

	// If list is empty, maybe we should be permissive or restrictive?
	// Default to restrictive if list provided?
	// If the user provided specific origins, they expect enforcement.
	return false
}
