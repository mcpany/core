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

// GuardrailsConfig represents the public GuardrailsConfig entity.
//
// Summary: Defines the structured data model representing a config.
//
// Parameters:
//   - None.
//
// Returns:
//   - None.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
type GuardrailsConfig struct {
	BlockedPhrases []string
}

// NewGuardrailsMiddleware serves as a public interface for interacting with NewGuardrailsMiddleware.
//
// Summary: Constructs and returns an initialized guardrails middleware ready for consumption.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the successfully computed domain model or execution state.
//
// Errors:
//   - No explicit errors are thrown by this operation.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
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
