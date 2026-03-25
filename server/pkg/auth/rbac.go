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
// Summary: Represents a RolesContextKey.
const RolesContextKey authContextKey = "user_roles"

// ContextWithRoles contextWithRoles context with roles.
//
// Summary: ContextWithRoles context with roles.
//
// Parameters: - None.
//   - ctx (context.Context): The context for the request.
//   - roles ([]string): The roles.
//
// Returns: - None.
//   - context.Context: The result.
func ContextWithRoles(ctx context.Context, roles []string) context.Context {
	return context.WithValue(ctx, RolesContextKey, roles)
}

// RolesFromContext rolesFromContext roles from context.
//
// Summary: RolesFromContext roles from context.
//
// Parameters: - None.
//   - ctx (context.Context): The context for the request.
//
// Returns: - None.
//   - []string: The result.
//   - bool: The result.
func RolesFromContext(ctx context.Context) ([]string, bool) {
	val, ok := ctx.Value(RolesContextKey).([]string)
	return val, ok
}

// RBACEnforcer handles Role-Based Access Control checks.
//
// Summary: Represents a RBACEnforcer.
type RBACEnforcer struct {
}

// NewRBACEnforcer creates a new rbac enforcer.
//
// Summary: Creates a new rbac enforcer.
//
// Parameters: - None.
//   - None.
//
// Returns: - None.
//   - *RBACEnforcer: The result.
func NewRBACEnforcer() *RBACEnforcer {
	return &RBACEnforcer{}
}

// HasRole hasRole has role.
//
// Summary: HasRole has role.
//
// Parameters: - None.
//   - user (*configv1.User): The user.
//   - role (string): The role.
//
// Returns: - None.
//   - bool: The result.
func (e *RBACEnforcer) HasRole(user *configv1.User, role string) bool {
	if user == nil {
		return false
	}
	return slices.Contains(user.GetRoles(), role)
}

// HasAnyRole hasAnyRole has any role.
//
// Summary: HasAnyRole has any role.
//
// Parameters: - None.
//   - user (*configv1.User): The user.
//   - roles ([]string): The roles.
//
// Returns: - None.
//   - bool: The result.
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

// HasRoleInContext hasRoleInContext has role in context.
//
// Summary: HasRoleInContext has role in context.
//
// Parameters: - None.
//   - ctx (context.Context): The context for the request.
//   - role (string): The role.
//
// Returns: - None.
//   - bool: The result.
func (e *RBACEnforcer) HasRoleInContext(ctx context.Context, role string) bool {
	roles, ok := RolesFromContext(ctx)
	if !ok {
		return false
	}
	return slices.Contains(roles, role)
}
