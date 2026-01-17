// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware

import "net/http"

// HTTPSecurityHeadersMiddleware adds security headers to HTTP responses.
//
// next is the next.
//
// Returns the result.
func HTTPSecurityHeadersMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "SAMEORIGIN")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		// Enhanced CSP:
		// - object-src 'none': Blocks plugins like Flash/Java.
		// - base-uri 'self': Prevents base tag hijacking.
		// - img-src 'self' data: https: : Allows images from self, data URIs, and HTTPS sources.
		// - frame-ancestors 'self': Prevents clickjacking by only allowing framing from the same origin.
		w.Header().Set("Content-Security-Policy", "default-src 'self'; style-src 'self' 'unsafe-inline' https://fonts.googleapis.com; font-src 'self' https://fonts.gstatic.com; script-src 'self'; connect-src 'self'; img-src 'self' data: https:; object-src 'none'; base-uri 'self'; frame-ancestors 'self'; form-action 'self'; upgrade-insecure-requests")
		w.Header().Set("Strict-Transport-Security", "max-age=63072000; includeSubDomains")
		w.Header().Set("Permissions-Policy", "geolocation=(), camera=(), microphone=(), payment=(), usb=(), vr=()")
		// Prevent caching of sensitive data
		w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate, proxy-revalidate")
		w.Header().Set("Pragma", "no-cache")
		w.Header().Set("Expires", "0")
		// Prevent information leakage about the server
		w.Header().Set("Server", "")
		next.ServeHTTP(w, r)
	})
}
