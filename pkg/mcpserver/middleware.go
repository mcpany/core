// Copyright (C) 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package mcpserver

import (
	"crypto/subtle"
	"net/http"
)

// AuthenticationMiddleware is a middleware that checks for a valid API key in the request header.
func AuthenticationMiddleware(apiKey string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if apiKey == "" {
				next.ServeHTTP(w, r)
				return
			}

			providedKey := r.Header.Get("X-API-Key")
			if subtle.ConstantTimeCompare([]byte(providedKey), []byte(apiKey)) != 1 {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
