package middleware

import (
	"fmt"
	"net/http"

	"github.com/mcpany/core/server/pkg/auth"
	configv1 "github.com/mcpany/core/proto/config/v1"
)

// RBACMiddleware provides middleware for Role-Based Access Control.
type RBACMiddleware struct {
	enforcer *auth.RBACEnforcer
}

// NewRBACMiddleware creates a new RBACMiddleware.
//
// Returns the result.
func NewRBACMiddleware() *RBACMiddleware {
	return &RBACMiddleware{
		enforcer: auth.NewRBACEnforcer(),
	}
}

// RequireRole returns an HTTP middleware that requires the user to have the specified role.
// It assumes that the user roles are already populated in the request context (e.g., by an authentication middleware).
func (m *RBACMiddleware) RequireRole(role string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Check if role is present in context
			if m.enforcer.HasRoleInContext(r.Context(), role) {
				next.ServeHTTP(w, r)
				return
			}

			// If roles are not in context, we might be able to check user object if present
			// However, standard pattern is to use ContextWithRoles.

			// Fallback: check if "user" is in context and use that
			// This depends on how auth middleware populates context.
			// Assuming auth.ContextWithUser or similar puts a *configv1.User or user ID.
			// But server/pkg/app/server.go uses auth.ContextWithRoles.

			http.Error(w, fmt.Sprintf("Forbidden: requires role %s", role), http.StatusForbidden)
		})
	}
}

// RequireAnyRole returns an HTTP middleware that requires the user to have at least one of the specified roles.
//
// roles is the roles.
//
// Returns the result.
func (m *RBACMiddleware) RequireAnyRole(roles ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			// Check directly against context roles
			userRoles, ok := auth.RolesFromContext(ctx)
			if ok {
				// We manually check here because Enforcer.HasAnyRole takes *User
				// But we can add a helper to Enforcer or do it here.
				for _, reqRole := range roles {
					for _, userRole := range userRoles {
						if reqRole == userRole {
							next.ServeHTTP(w, r)
							return
						}
					}
				}
			}

			http.Error(w, fmt.Sprintf("Forbidden: requires one of roles %v", roles), http.StatusForbidden)
		})
	}
}

// EnforcePolicy allows passing a custom policy function.
//
// _ is an unused parameter.
//
// Returns the result.
func (m *RBACMiddleware) EnforcePolicy(_ func(user *configv1.User) bool) func(http.Handler) http.Handler {
	return func(_ http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			// We need the user object here.
			// Assuming it's in context? `server.go` doesn't seem to put the full User struct in a standard key that is exported.
			// It puts UID.
			// For now, we'll stick to Role-based checks which use RolesFromContext.
			http.Error(w, "Policy enforcement not implemented (requires user object in context)", http.StatusNotImplemented)
		})
	}
}
