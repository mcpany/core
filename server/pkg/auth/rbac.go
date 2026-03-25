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

// ContextWithRoles returns a new context with the user roles. ctx is the context for the request. roles is the roles. Returns the result.
//
// Parameters: - None.
//   - ctx (context.Context): The context for the request.
//   - roles ([]string): The roles parameter.
//
// Returns: - None.
//   - context.Context: The resulting context.Context.
//
// Errors: - None.
//   - None.
//
// Side Effects: - None.
//   - None.
//
// Summary: Executes ContextWithRoles operation.
//
// Parameters: - None.
//
// Returns: - None.
//
// Errors: - None.
//
// Side Effects: - None.
//   - None.
func ContextWithRoles(ctx context.Context, roles []string) context.Context {
	return context.WithValue(ctx, RolesContextKey, roles)
}

// RolesFromContext returns the user roles from the context. ctx is the context for the request. Returns the result. Returns true if successful.
//
// Parameters: - None.
//   - ctx (context.Context): The context for the request.
//
// Returns: - None.
//   - []string: The resulting []string.
//   - bool: True if successful, false otherwise.
//
// Errors: - None.
//   - None.
//
// Side Effects: - None.
//   - None.
//
// Summary: Executes RolesFromContext operation.
//
// Parameters: - None.
//
// Returns: - None.
//
// Errors: - None.
//
// Side Effects: - None.
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

// NewRBACEnforcer creates a new RBACEnforcer. Returns the result.
//
// Parameters: - None.
//   - None.
//
// Returns: - None.
//   - *RBACEnforcer: The resulting *RBACEnforcer.
//
// Errors: - None.
//   - None.
//
// Side Effects: - None.
//   - None.
//
// Summary: Initializes NewRBACEnforcer operation.
//
// Parameters: - None.
//
// Returns: - None.
//
// Errors: - None.
//
// Side Effects: - None.
//   - None.
func NewRBACEnforcer() *RBACEnforcer {
	return &RBACEnforcer{}
}

// HasRole checks if the given user has the specified role. user is the user. role is the role. Returns true if successful.
//
// Parameters: - None.
//   - user (*configv1.User): The user parameter.
//   - role (string): The role parameter.
//
// Returns: - None.
//   - bool: True if successful, false otherwise.
//
// Errors: - None.
//   - None.
//
// Side Effects: - None.
//   - None.
//
// Summary: Checks HasRole operation.
//
// Parameters: - None.
//
// Returns: - None.
//
// Errors: - None.
//
// Side Effects: - None.
//   - None.
func (e *RBACEnforcer) HasRole(user *configv1.User, role string) bool {
	if user == nil {
		return false
	}
	return slices.Contains(user.GetRoles(), role)
}

// HasAnyRole checks if the user has at least one of the specified roles. user is the user. roles is the roles. Returns true if successful.
//
// Parameters: - None.
//   - user (*configv1.User): The user parameter.
//   - roles ([]string): The roles parameter.
//
// Returns: - None.
//   - bool: True if successful, false otherwise.
//
// Errors: - None.
//   - None.
//
// Side Effects: - None.
//   - None.
//
// Summary: Checks HasAnyRole operation.
//
// Parameters: - None.
//
// Returns: - None.
//
// Errors: - None.
//
// Side Effects: - None.
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

// HasRoleInContext checks if the context contains the specified role. ctx is the context for the request. role is the role. Returns true if successful.
//
// Parameters: - None.
//   - ctx (context.Context): The context for the request.
//   - role (string): The role parameter.
//
// Returns: - None.
//   - bool: True if successful, false otherwise.
//
// Errors: - None.
//   - None.
//
// Side Effects: - None.
//   - None.
//
// Summary: Checks HasRoleInContext operation.
//
// Parameters: - None.
//
// Returns: - None.
//
// Errors: - None.
//
// Side Effects: - None.
//   - None.
func (e *RBACEnforcer) HasRoleInContext(ctx context.Context, role string) bool {
	roles, ok := RolesFromContext(ctx)
	if !ok {
		return false
	}
	return slices.Contains(roles, role)
}
