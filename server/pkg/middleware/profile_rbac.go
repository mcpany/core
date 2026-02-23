// Copyright 2026 Author(s) of MCP Any
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
// Logic: A user is allowed if they satisfy the requirements of AT LEAST ONE active profile.
// - If a profile has no RequiredRoles, it allows everyone (Public).
// - If a profile has RequiredRoles, the user must have at least one of them to satisfy that profile.
//
// Parameters:
//   - profiles: []*configv1.ProfileDefinition. The list of active profile definitions.
//
// Returns:
//   - func(http.Handler) http.Handler: The middleware function.
func ProfileRBACMiddleware(profiles []*configv1.ProfileDefinition) func(http.Handler) http.Handler {
	// Optimization: If any profile has NO requirements, then the server is effectively public.
	// We can skip the middleware entirely in that case.
	for _, p := range profiles {
		if len(p.GetRequiredRoles()) == 0 {
			logging.GetLogger().Info("Profile RBAC: Found unrestricted profile, allowing public access", "profile", p.GetName())
			return func(next http.Handler) http.Handler {
				return next
			}
		}
	}

	// Optimization: If no profiles are active, we default to allow (or should we fail closed? usually empty config = allow)
	if len(profiles) == 0 {
		return func(next http.Handler) http.Handler {
			return next
		}
	}

	logging.GetLogger().Info("Enforcing Profile RBAC: All active profiles have role requirements")

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			// Get user roles from context
			userRoles, ok := auth.RolesFromContext(ctx)
			if !ok {
				userRoles = []string{} // Empty roles for anonymous/unauthenticated
			}

			// Check if user satisfies ANY of the active profiles
			satisfied := false
			for _, p := range profiles {
				reqs := p.GetRequiredRoles()
				// This should not happen due to optimization above, but safe to check
				if len(reqs) == 0 {
					satisfied = true
					break
				}

				// Check match
				for _, userRole := range userRoles {
					for _, reqRole := range reqs {
						if userRole == reqRole {
							satisfied = true
							break
						}
					}
					if satisfied {
						break
					}
				}
				if satisfied {
					break
				}
			}

			if !satisfied {
				http.Error(w, fmt.Sprintf("Forbidden: User roles %v do not satisfy any active profile requirements", userRoles), http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
