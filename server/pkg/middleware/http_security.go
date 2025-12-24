// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware

import "net/http"

// HTTPSecurityHeadersMiddleware adds security headers to HTTP responses.
func HTTPSecurityHeadersMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "SAMEORIGIN")
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")

		// Defense in Depth: Content Security Policy
		// We allow 'unsafe-inline' for styles because the simple UI uses them.
		w.Header().Set("Content-Security-Policy", "default-src 'self'; script-src 'self'; style-src 'self' 'unsafe-inline'; img-src 'self' data:; font-src 'self'; frame-ancestors 'self'; object-src 'none'; base-uri 'self'; form-action 'self';")

		// Force HTTPS behavior (even if currently on HTTP, for future proofing/proxies)
		w.Header().Set("Strict-Transport-Security", "max-age=63072000; includeSubDomains")

		next.ServeHTTP(w, r)
	})
}
