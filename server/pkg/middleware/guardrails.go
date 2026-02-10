// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"bytes"
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// GuardrailsConfig defines patterns to block.
//
// Summary: Configuration for Guardrails.
//
// Fields:
//   - BlockedPhrases: []string. List of phrases to block.
type GuardrailsConfig struct {
	BlockedPhrases []string
}

// NewGuardrailsMiddleware creates a new Guardrails middleware.
//
// Summary: Middleware for blocking malicious inputs.
//
// Parameters:
//   - config: GuardrailsConfig. The configuration.
//
// Returns:
//   - gin.HandlerFunc: The Gin middleware handler.
func NewGuardrailsMiddleware(config GuardrailsConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Only check POST requests (likely prompt submissions)
		if c.Request.Method != "POST" {
			c.Next()
			return
		}

		// Read body
		bodyBytes, err := io.ReadAll(c.Request.Body)
		if err != nil {
			c.Next()
			return
		}

		// Restore body
		c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

		// Check for blocked phrases
		// Convert to lower case for case-insensitive matching logic MVP
		bodyLower := strings.ToLower(string(bodyBytes))

		for _, phrase := range config.BlockedPhrases {
			if strings.Contains(bodyLower, strings.ToLower(phrase)) {
				c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
					"error":  "Prompt Injection Detected: Request blocked by validation policy.",
					"policy": "no-jailbreak",
				})
				return
			}
		}

		c.Next()
	}
}
