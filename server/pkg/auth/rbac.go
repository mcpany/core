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

// ContextWithRoles provides contextwithroles functionality.
//
// Summary: ContextWithRoles.
//
// Parameters.
//   - ctx: The parameter.
//   - roles: The parameter.
//
// Returns.
//   - result: The result.
func ContextWithRoles(ctx context.Context, roles []string) context.Context {
	return context.WithValue(ctx, RolesContextKey, roles)
}

// RolesFromContext provides rolesfromcontext functionality.
//
// Summary: RolesFromContext.
//
// Parameters.
//   - ctx: The parameter.
//   - bool: The parameter.
//
// Returns.
//   - None.
func RolesFromContext(ctx context.Context) ([]string, bool) {
	val, ok := ctx.Value(RolesContextKey).([]string)
	return val, ok
}

// RBACEnforcer handles Role-Based Access Control checks.
//
// Summary: Represents a RBACEnforcer.
type RBACEnforcer struct {
}

// NewRBACEnforcer provides newrbacenforcer functionality.
//
// Summary: NewRBACEnforcer.
//
// Parameters.
//   - None.
//
// Returns.
//   - result: The result.
func NewRBACEnforcer() *RBACEnforcer {
	return &RBACEnforcer{}
}

// HasRole provides hasrole functionality.
//
// Summary: HasRole.
//
// Parameters.
//   - user: The parameter.
//   - role: The parameter.
//
// Returns.
//   - result: The result.
func (e *RBACEnforcer) HasRole(user *configv1.User, role string) bool {
	if user == nil {
		return false
	}
	return slices.Contains(user.GetRoles(), role)
}

// HasAnyRole provides hasanyrole functionality.
//
// Summary: HasAnyRole.
//
// Parameters.
//   - user: The parameter.
//   - roles: The parameter.
//
// Returns.
//   - result: The result.
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

// HasRoleInContext provides hasroleincontext functionality.
//
// Summary: HasRoleInContext.
//
// Parameters.
//   - ctx: The parameter.
//   - role: The parameter.
//
// Returns.
//   - result: The result.
func (e *RBACEnforcer) HasRoleInContext(ctx context.Context, role string) bool {
	roles, ok := RolesFromContext(ctx)
	if !ok {
		return false
	}
	return slices.Contains(roles, role)
}
