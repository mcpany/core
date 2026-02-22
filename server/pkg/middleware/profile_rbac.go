// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"fmt"
	"net/http"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/auth"
	"github.com/mcpany/core/server/pkg/logging"
)

// ProfileRBACMiddleware creates a middleware that enforces role requirements defined in active profiles.
//
// Summary: Restricts access to the server based on the required roles of active profiles.
//
// Parameters:
//   - profiles: []*configv1.ProfileDefinition. The list of active profile definitions.
//
// Returns:
//   - func(http.Handler) http.Handler: The middleware function.
func ProfileRBACMiddleware(profiles []*configv1.ProfileDefinition) func(http.Handler) http.Handler {
	// Pre-compute the set of required roles
	requiredRoles := make(map[string]struct{})
	hasRequirements := false

	for _, p := range profiles {
		for _, role := range p.GetRequiredRoles() {
			requiredRoles[role] = struct{}{}
			hasRequirements = true
		}
	}

	if !hasRequirements {
		// Optimization: No requirements, return pass-through middleware
		return func(next http.Handler) http.Handler {
			return next
		}
	}

	logging.GetLogger().Info("Enforcing Profile RBAC", "required_roles", requiredRoles)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			// Get user roles from context
			userRoles, ok := auth.RolesFromContext(ctx)
			if !ok {
				// No roles found in context.
				// If requirements exist, and user has no roles, access denied.
				// Note: Anonymous users might not have roles.
				http.Error(w, "Forbidden: Access requires specific roles", http.StatusForbidden)
				return
			}

			// Check if user has at least one of the required roles
			// Logic: If the server requires [A, B], and user has [A], access granted.
			// This interprets the requirement as "User must have one of the allowed roles".
			allowed := false
			for _, userRole := range userRoles {
				if _, found := requiredRoles[userRole]; found {
					allowed = true
					break
				}
			}

			if !allowed {
				http.Error(w, fmt.Sprintf("Forbidden: User roles %v do not satisfy profile requirements", userRoles), http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
