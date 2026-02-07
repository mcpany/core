// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// SSOConfig defines the SSO configuration.
type SSOConfig struct {
	Enabled bool
	IDPURL  string
}

// SSOMiddleware creates a new SSO middleware.
//
// Summary: Middleware that handles Single Sign-On (SSO) authentication.
//
// Parameters:
//   - config: SSOConfig. The configuration for SSO.
//
// Returns:
//   - gin.HandlerFunc: The Gin middleware function.
func SSOMiddleware(config SSOConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !config.Enabled {
			c.Next()
			return
		}

		// Check for Identity Header (Trusted Proxy pattern)
		userID := c.GetHeader("X-MCP-Identity")
		if userID != "" {
			c.Set("UserID", userID)
			c.Next()
			return
		}

		// Check for Bearer Token
		auth := c.GetHeader("Authorization")
		if strings.HasPrefix(auth, "Bearer ") {
			// Validate token (Mock validation)
			token := strings.TrimPrefix(auth, "Bearer ")
			if token == "valid-mock-token" {
				c.Set("UserID", "user-123")
				c.Next()
				return
			}
		}

		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"error":     "Authentication required",
			"login_url": config.IDPURL + "/login",
		})
	}
}
