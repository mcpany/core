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

// ContextWithRoles returns a new context with the user roles embedded.
//
// Parameters:
//   - ctx: context.Context. The context to extend.
//   - roles: []string. The list of roles to store.
//
// Returns:
//   - context.Context: A new context containing the roles.
func ContextWithRoles(ctx context.Context, roles []string) context.Context {
	return context.WithValue(ctx, RolesContextKey, roles)
}

// RolesFromContext retrieves the user roles from the context.
//
// Parameters:
//   - ctx: context.Context. The context to search.
//
// Returns:
//   - []string: The list of roles if found.
//   - bool: True if roles exist, false otherwise.
func RolesFromContext(ctx context.Context) ([]string, bool) {
	val, ok := ctx.Value(RolesContextKey).([]string)
	return val, ok
}

// RBACEnforcer handles Role-Based Access Control checks.
type RBACEnforcer struct {
}

// NewRBACEnforcer creates a new RBACEnforcer instance.
//
// Parameters:
//   None.
//
// Returns:
//   - *RBACEnforcer: The initialized enforcer.
func NewRBACEnforcer() *RBACEnforcer {
	return &RBACEnforcer{}
}

// HasRole checks if the user possesses the specified role.
//
// Parameters:
//   - user: *configv1.User. The user configuration.
//   - role: string. The role to check.
//
// Returns:
//   - bool: True if the user has the role, false otherwise.
func (e *RBACEnforcer) HasRole(user *configv1.User, role string) bool {
	if user == nil {
		return false
	}
	return slices.Contains(user.GetRoles(), role)
}

// HasAnyRole checks if the user possesses at least one of the specified roles.
//
// Parameters:
//   - user: *configv1.User. The user configuration.
//   - roles: []string. The list of roles to check against.
//
// Returns:
//   - bool: True if the user has any of the roles, false otherwise.
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
// Parameters:
//   - ctx: context.Context. The context to search.
//   - role: string. The role to check.
//
// Returns:
//   - bool: True if the role is present in the context, false otherwise.
func (e *RBACEnforcer) HasRoleInContext(ctx context.Context, role string) bool {
	roles, ok := RolesFromContext(ctx)
	if !ok {
		return false
	}
	return slices.Contains(roles, role)
}
