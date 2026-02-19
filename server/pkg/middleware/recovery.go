// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"net/http"
	"runtime/debug"

	"github.com/mcpany/core/server/pkg/logging"
)

// RecoveryMiddleware recovers from panics in the handler chain.
//
// It logs the panic with a stack trace and returns a generic 500 Internal Server Error response.
//
// Parameters:
//  - next (http.Handler): The next handler in the chain.
//
// Returns:
//  - http.Handler: The wrapped handler with recovery logic.
func RecoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				// Log the panic with stack trace
				log := logging.GetLogger()
				stack := string(debug.Stack())
				log.Error("Panic recovered", "error", err, "stack", stack, "url", r.URL.String(), "method", r.Method)

				// Return generic 500 error
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}
