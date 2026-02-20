// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package auth

import (
	"context"
	"slices"

	configv1 "github.com/mcpany/core/proto/config/v1"
)

// RolesContextKey is the context key for the user roles.
//
// Summary: Context key for storing user roles.
const RolesContextKey authContextKey = "user_roles"

// ContextWithRoles returns a new context with the user roles.
//
// Summary: Embeds user roles into the context.
//
// Parameters:
//   - ctx: context.Context. The context for the request.
//   - roles: []string. The list of roles to store.
//
// Returns:
//   - context.Context: A new context containing the roles.
func ContextWithRoles(ctx context.Context, roles []string) context.Context {
	return context.WithValue(ctx, RolesContextKey, roles)
}

// RolesFromContext returns the user roles from the context.
//
// Summary: Retrieves user roles from the context.
//
// Parameters:
//   - ctx: context.Context. The context for the request.
//
// Returns:
//   - []string: The list of roles.
//   - bool: True if found.
func RolesFromContext(ctx context.Context) ([]string, bool) {
	val, ok := ctx.Value(RolesContextKey).([]string)
	return val, ok
}

// RBACEnforcer handles Role-Based Access Control checks.
//
// Summary: Enforces RBAC policies.
type RBACEnforcer struct {
}

// NewRBACEnforcer creates a new RBACEnforcer.
//
// Summary: Initializes a new RBACEnforcer.
//
// Returns:
//   - *RBACEnforcer: The new instance.
func NewRBACEnforcer() *RBACEnforcer {
	return &RBACEnforcer{}
}

// HasRole checks if the given user has the specified role.
//
// Summary: Checks if a user has a specific role.
//
// Parameters:
//   - user: *configv1.User. The user to check.
//   - role: string. The role to check for.
//
// Returns:
//   - bool: True if the user has the role.
func (e *RBACEnforcer) HasRole(user *configv1.User, role string) bool {
	if user == nil {
		return false
	}
	return slices.Contains(user.GetRoles(), role)
}

// HasAnyRole checks if the user has at least one of the specified roles.
//
// Summary: Checks if a user has any of the specified roles.
//
// Parameters:
//   - user: *configv1.User. The user to check.
//   - roles: []string. The list of roles to check.
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
// Summary: Checks if the context contains a specific role.
//
// Parameters:
//   - ctx: context.Context. The context to check.
//   - role: string. The role to check for.
//
// Returns:
//   - bool: True if the context contains the role.
func (e *RBACEnforcer) HasRoleInContext(ctx context.Context, role string) bool {
	roles, ok := RolesFromContext(ctx)
	if !ok {
		return false
	}
	return slices.Contains(roles, role)
}
