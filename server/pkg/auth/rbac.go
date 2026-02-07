// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package auth

import (
	"context"
	"slices"

	configv1 "github.com/mcpany/core/proto/config/v1"
)

// RolesContextKey is the context key for the user roles.
const RolesContextKey authContextKey = "user_roles"

// ContextWithRoles returns a new context with the user roles.
//
// Summary: Embeds the user's roles into the context for downstream access control checks.
//
// Parameters:
//   - ctx: context.Context. The parent context.
//   - roles: []string. A list of role names associated with the user.
//
// Returns:
//   - context.Context: A new context containing the roles.
func ContextWithRoles(ctx context.Context, roles []string) context.Context {
	return context.WithValue(ctx, RolesContextKey, roles)
}

// RolesFromContext returns the user roles from the context.
//
// Summary: Retrieves the list of roles assigned to the user from the context.
//
// Parameters:
//   - ctx: context.Context. The context to search.
//
// Returns:
//   - []string: The list of roles.
//   - bool: True if roles were found in the context.
func RolesFromContext(ctx context.Context) ([]string, bool) {
	val, ok := ctx.Value(RolesContextKey).([]string)
	return val, ok
}

// RBACEnforcer handles Role-Based Access Control checks.
//
// Summary: Provides methods to enforce role-based access policies on users.
type RBACEnforcer struct {
}

// NewRBACEnforcer creates a new RBACEnforcer.
//
// Summary: Initializes a new RBACEnforcer instance.
//
// Returns:
//   - *RBACEnforcer: A pointer to the new enforcer.
func NewRBACEnforcer() *RBACEnforcer {
	return &RBACEnforcer{}
}

// HasRole checks if the given user has the specified role.
//
// Summary: Verifies if a user possesses a specific role.
//
// Parameters:
//   - user: *configv1.User. The user to check.
//   - role: string. The role name to verify.
//
// Returns:
//   - bool: True if the user has the role, false otherwise (including if user is nil).
func (e *RBACEnforcer) HasRole(user *configv1.User, role string) bool {
	if user == nil {
		return false
	}
	return slices.Contains(user.GetRoles(), role)
}

// HasAnyRole checks if the user has at least one of the specified roles.
//
// Summary: Verifies if a user possesses any one of the provided roles.
//
// Parameters:
//   - user: *configv1.User. The user to check.
//   - roles: []string. A list of role names.
//
// Returns:
//   - bool: True if the user has at least one of the roles.
func (e *RBACEnforcer) HasAnyRole(user *configv1.User, roles []string) bool {
	if user == nil {
		return false
	}
	for _, role := range roles {
		if slices.Contains(user.GetRoles(), role) {
			return true
		}
	}
	return false
}

// HasRoleInContext checks if the context contains the specified role.
//
// Summary: Verifies if the authenticated user in the context possesses a specific role.
//
// Parameters:
//   - ctx: context.Context. The request context containing user roles.
//   - role: string. The role name to verify.
//
// Returns:
//   - bool: True if the role is present in the context.
func (e *RBACEnforcer) HasRoleInContext(ctx context.Context, role string) bool {
	roles, ok := RolesFromContext(ctx)
	if !ok {
		return false
	}
	return slices.Contains(roles, role)
}
