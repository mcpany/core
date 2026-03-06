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

// GuardrailsConfig - Auto-generated documentation.
//
// Summary: GuardrailsConfig defines patterns to block.
//
// Fields:
//   - Various fields for GuardrailsConfig.
type GuardrailsConfig struct {
	BlockedPhrases []string
}

// NewGuardrailsMiddleware creates a new Guardrails middleware. Summary: Initializes the guardrails middleware for blocking malicious prompts. Parameters: - config: GuardrailsConfig. The configuration for blocking patterns. Returns: - gin.HandlerFunc: The Gin middleware handler.
//
// Summary: NewGuardrailsMiddleware creates a new Guardrails middleware. Summary: Initializes the guardrails middleware for blocking malicious prompts. Parameters: - config: GuardrailsConfig. The configuration for blocking patterns. Returns: - gin.HandlerFunc: The Gin middleware handler.
//
// Parameters:
//   - config (GuardrailsConfig): The configuration settings to be applied.
//
// Returns:
//   - (gin.HandlerFunc): The resulting gin.HandlerFunc object containing the requested data.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
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
