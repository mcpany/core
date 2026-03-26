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
// Summary: Context key used for storing and retrieving user roles.
const RolesContextKey authContextKey = "user_roles"

// ContextWithRoles returns a new context with the user roles.
//
// Summary: Executes ContextWithRoles operation.
//
// Parameters:
//   - ctx: context.Context. The context for the request.
//   - roles: []string. The roles to add to the context.
//
// Returns:
//   - context.Context: The new context with roles.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func ContextWithRoles(ctx context.Context, roles []string) context.Context {
	return context.WithValue(ctx, RolesContextKey, roles)
}

// RolesFromContext returns the user roles from the context.
//
// Summary: Executes RolesFromContext operation.
//
// Parameters:
//   - ctx: context.Context. The context for the request.
//
// Returns:
//   - []string: The user roles.
//   - bool: True if roles were found, false otherwise.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func RolesFromContext(ctx context.Context) ([]string, bool) {
	val, ok := ctx.Value(RolesContextKey).([]string)
	return val, ok
}

// RBACEnforcer handles Role-Based Access Control checks.
//
// Summary: Component for enforcing Role-Based Access Control (RBAC) policies.
type RBACEnforcer struct {
}

// NewRBACEnforcer creates a new RBACEnforcer.
//
// Summary: Initializes NewRBACEnforcer operation.
//
// Returns:
//   - *RBACEnforcer: The initialized enforcer.
//
// Parameters:
//   - None.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func NewRBACEnforcer() *RBACEnforcer {
	return &RBACEnforcer{}
}

// HasRole checks if the given user has the specified role.
//
// Summary: Checks HasRole operation.
//
// Parameters:
//   - user: *configv1.User. The user to check.
//   - role: string. The role to look for.
//
// Returns:
//   - bool: True if the user has the role, false otherwise.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func (e *RBACEnforcer) HasRole(user *configv1.User, role string) bool {
	if user == nil {
		return false
	}
	return slices.Contains(user.GetRoles(), role)
}

// HasAnyRole checks if the user has at least one of the specified roles.
//
// Summary: Checks HasAnyRole operation.
//
// Parameters:
//   - user: *configv1.User. The user to check.
//   - roles: []string. The roles to check against.
//
// Returns:
//   - bool: True if the user has any of the roles, false otherwise.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
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
// Summary: Checks HasRoleInContext operation.
//
// Parameters:
//   - ctx: context.Context. The request context.
//   - role: string. The role to check for.
//
// Returns:
//   - bool: True if the role is found in the context, false otherwise.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func (e *RBACEnforcer) HasRoleInContext(ctx context.Context, role string) bool {
	roles, ok := RolesFromContext(ctx)
	if !ok {
		return false
	}
	return slices.Contains(roles, role)
}
