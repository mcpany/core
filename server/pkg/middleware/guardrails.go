// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware
// NewGuardrailsMiddleware creates a new Guardrails middleware.
//
// Summary: Initializes the guardrails middleware for blocking malicious prompts.
//
// Parameters:
//   - config: GuardrailsConfig. The configuration for blocking patterns.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
//
// Returns:
//   - gin.HandlerFunc: The Gin middleware handler.
// Errors:
//   - triggers relevant error states on failure.
// Side Effects:
//   - updates relevant subsystem state or network conditions.
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
