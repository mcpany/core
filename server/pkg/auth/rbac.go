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
// Summary: returns a new context with the user roles.
//
// Parameters:
//   - ctx: context.Context. The context for the operation.
//   - roles: []string. The roles.
//
// Returns:
//   - context.Context: The context.Context.
func ContextWithRoles(ctx context.Context, roles []string) context.Context {
	return context.WithValue(ctx, RolesContextKey, roles)
}

// RolesFromContext returns the user roles from the context.
//
// Summary: returns the user roles from the context.
//
// Parameters:
//   - ctx: context.Context. The context for the operation.
//
// Returns:
//   - []string: The []string.
//   - bool: The bool.
func RolesFromContext(ctx context.Context) ([]string, bool) {
	val, ok := ctx.Value(RolesContextKey).([]string)
	return val, ok
}

// RBACEnforcer handles Role-Based Access Control checks.
//
// Summary: handles Role-Based Access Control checks.
type RBACEnforcer struct {
}

// NewRBACEnforcer creates a new RBACEnforcer.
//
// Summary: creates a new RBACEnforcer.
//
// Parameters:
//   None.
//
// Returns:
//   - *RBACEnforcer: The *RBACEnforcer.
func NewRBACEnforcer() *RBACEnforcer {
	return &RBACEnforcer{}
}

// HasRole checks if the given user has the specified role.
//
// Summary: checks if the given user has the specified role.
//
// Parameters:
//   - user: *configv1.User. The user.
//   - role: string. The role.
//
// Returns:
//   - bool: The bool.
func (e *RBACEnforcer) HasRole(user *configv1.User, role string) bool {
	if user == nil {
		return false
	}
	return slices.Contains(user.GetRoles(), role)
}

// HasAnyRole checks if the user has at least one of the specified roles.
//
// Summary: checks if the user has at least one of the specified roles.
//
// Parameters:
//   - user: *configv1.User. The user.
//   - roles: []string. The roles.
//
// Returns:
//   - bool: The bool.
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
// Summary: checks if the context contains the specified role.
//
// Parameters:
//   - ctx: context.Context. The context for the operation.
//   - role: string. The role.
//
// Returns:
//   - bool: The bool.
func (e *RBACEnforcer) HasRoleInContext(ctx context.Context, role string) bool {
	roles, ok := RolesFromContext(ctx)
	if !ok {
		return false
	}
	return slices.Contains(roles, role)
}
