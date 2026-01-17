// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strings"
)

// GuardrailsConfig defines patterns to block.
type GuardrailsConfig struct {
	BlockedPhrases []string
}

// NewGuardrailsMiddleware creates a new Guardrails middleware.
func NewGuardrailsMiddleware(config GuardrailsConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Only check POST requests (likely prompt submissions)
			if r.Method != http.MethodPost {
				next.ServeHTTP(w, r)
				return
			}

			// Read body
			bodyBytes, err := io.ReadAll(r.Body)
			if err != nil {
				next.ServeHTTP(w, r)
				return
			}

			// Restore body
			r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

			// Check for blocked phrases
			// Convert to lower case for case-insensitive matching logic MVP
			bodyLower := strings.ToLower(string(bodyBytes))

			for _, phrase := range config.BlockedPhrases {
				if strings.Contains(bodyLower, strings.ToLower(phrase)) {
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusBadRequest)
					_ = json.NewEncoder(w).Encode(map[string]string{
						"error":  "Prompt Injection Detected: Request blocked by validation policy.",
						"policy": "no-jailbreak",
					})
					return
				}
			}

			next.ServeHTTP(w, r)
		})
	}
}
