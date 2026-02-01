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
// Parameters:
//   - ctx: The context for the request.
//   - roles: The roles to add.
//
// Returns:
//   - context.Context: The new context with the roles.
func ContextWithRoles(ctx context.Context, roles []string) context.Context {
	return context.WithValue(ctx, RolesContextKey, roles)
}

// RolesFromContext returns the user roles from the context.
//
// Parameters:
//   - ctx: The context for the request.
//
// Returns:
//   - []string: The user roles.
//   - bool: True if the roles were found in the context.
func RolesFromContext(ctx context.Context) ([]string, bool) {
	val, ok := ctx.Value(RolesContextKey).([]string)
	return val, ok
}

// RBACEnforcer handles Role-Based Access Control checks.
type RBACEnforcer struct {
}

// NewRBACEnforcer creates a new RBACEnforcer.
//
// Returns:
//   - *RBACEnforcer: A new RBACEnforcer instance.
func NewRBACEnforcer() *RBACEnforcer {
	return &RBACEnforcer{}
}

// HasRole checks if the given user has the specified role.
//
// Parameters:
//   - user: The user to check.
//   - role: The role to check for.
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
// Parameters:
//   - user: The user to check.
//   - roles: The roles to check for.
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
// Parameters:
//   - ctx: The context for the request.
//   - role: The role to check for.
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
