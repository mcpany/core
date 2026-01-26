// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"net/http"
	"strings"
)

// HTTPSecurityHeadersMiddleware adds security headers to HTTP responses.
//
// next is the next.
//
// Returns the result.
func HTTPSecurityHeadersMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		// X-XSS-Protection: 0 disables the browser's XSS audit.
		// Setting it to 1; mode=block is considered bad practice in modern browsers
		// as it can introduce XS-Leak vulnerabilities.
		w.Header().Set("X-XSS-Protection", "0")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		w.Header().Set("Strict-Transport-Security", "max-age=63072000; includeSubDomains; preload")
		w.Header().Set("Permissions-Policy", "geolocation=(), camera=(), microphone=(), payment=(), usb=(), vr=(), magnetometer=(), gyroscope=(), accelerometer=(), autoplay=(), clipboard-write=(), clipboard-read=(), fullscreen=()")

		// Additional Security Headers
		// COOP and COEP are commented out to avoid breaking external resource loading (e.g. Google Fonts)
		// in E2E tests or production until all upstream resources support CORP.
		// w.Header().Set("Cross-Origin-Opener-Policy", "same-origin")
		// w.Header().Set("Cross-Origin-Embedder-Policy", "require-corp")
		w.Header().Set("Cross-Origin-Resource-Policy", "same-origin")
		w.Header().Set("X-Permitted-Cross-Domain-Policies", "none")

		// Prevent information leakage about the server
		w.Header().Del("Server")

		// Check if the request is for the UI
		if strings.HasPrefix(r.URL.Path, "/ui/") {
			// UI-specific headers
			// Matches ui/src/middleware.ts requirements:
			// - script-src: Allows unsafe-eval (Monaco), unsafe-inline (Next.js), cdn.jsdelivr.net (Monaco)
			// - style-src: Allows unsafe-inline (Next.js), cdn.jsdelivr.net
			// - connect-src: Allows cdn.jsdelivr.net
			// - worker-src: Allows blob: (Monaco)
			// - frame-ancestors: none
			w.Header().Set("Content-Security-Policy", "default-src 'self'; script-src 'self' 'unsafe-eval' 'unsafe-inline' https://cdn.jsdelivr.net; style-src 'self' 'unsafe-inline' https://fonts.googleapis.com https://cdn.jsdelivr.net; img-src 'self' data: https:; font-src 'self' data: https://fonts.gstatic.com; connect-src 'self' https://cdn.jsdelivr.net; worker-src 'self' blob:; object-src 'none'; base-uri 'self'; frame-ancestors 'none'; form-action 'self'; upgrade-insecure-requests")
			w.Header().Set("X-Frame-Options", "DENY")

			// For UI, we allow caching (do not set strict no-cache headers).
			// http.FileServer handles ETag/Last-Modified.
		} else {
			// API/Default headers (Strict)
			// - object-src 'none': Blocks plugins like Flash/Java.
			// - base-uri 'self': Prevents base tag hijacking.
			// - img-src 'self' data: https: : Allows images from self, data URIs, and HTTPS sources.
			// - frame-ancestors 'self': Prevents clickjacking by only allowing framing from the same origin.
			w.Header().Set("Content-Security-Policy", "default-src 'self'; style-src 'self' 'unsafe-inline' https://fonts.googleapis.com; font-src 'self' https://fonts.gstatic.com; script-src 'self'; connect-src 'self'; img-src 'self' data: https:; object-src 'none'; base-uri 'self'; frame-ancestors 'self'; form-action 'self'; upgrade-insecure-requests")
			w.Header().Set("X-Frame-Options", "SAMEORIGIN")

			// Prevent caching of sensitive data
			w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate, proxy-revalidate")
			w.Header().Set("Pragma", "no-cache")
			w.Header().Set("Expires", "0")
		}

		next.ServeHTTP(w, r)
	})
}
