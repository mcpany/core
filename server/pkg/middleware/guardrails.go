// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"bytes"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
)

// GuardrailsConfig defines patterns to block.
//
// Summary: Configuration for the guardrails middleware.
type GuardrailsConfig struct {
	BlockedPhrases []string
}

// NewGuardrailsMiddleware creates a new Guardrails middleware.
//
// Summary: Initializes the guardrails middleware for blocking malicious prompts.
//
// Parameters:
//   - config: GuardrailsConfig. The configuration for blocking patterns.
//
// Returns:
//   - gin.HandlerFunc: The Gin middleware handler.
func NewGuardrailsMiddleware(config GuardrailsConfig) gin.HandlerFunc {
	// ⚡ BOLT: Pre-compute lowercase byte slices to avoid repetitive conversion and allocations.
	// Randomized Selection from Top 5 High-Impact Targets
	blockedPhrasesLower := make([][]byte, len(config.BlockedPhrases))
	for i, phrase := range config.BlockedPhrases {
		blockedPhrasesLower[i] = bytes.ToLower([]byte(phrase))
	}

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
		// ⚡ BOLT: Use bytes.ToLower to avoid string allocation.
		// Randomized Selection from Top 5 High-Impact Targets
		bodyLower := bytes.ToLower(bodyBytes)

		for _, phraseLower := range blockedPhrasesLower {
			if bytes.Contains(bodyLower, phraseLower) {
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
