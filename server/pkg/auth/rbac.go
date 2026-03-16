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
// Summary: RolesContextKey is the context key for the user roles.
//
// Summary: RolesContextKey is the context key for the user roles.
// ContextWithRoles returns a new context with the user roles. ctx is the context for the request. roles is the roles. Returns the result.
//
// Summary: ContextWithRoles returns a new context with the user roles. ctx is the context for the request. roles is the roles. Returns the result.
//
// Parameters:
//   - ctx (context.Context): The cancellation and deadline context.
//   - roles ([]string): The textual representation of roles.
//
// Returns:
//   - context.Context: The resulting object or data structure.
//
// Errors:
//   - None.
//
// Side Effects:
//   - May modify internal state or perform external network calls.
//   - None.
//
// Side Effects:
//   - May modify internal state or perform external network calls.
// RolesFromContext returns the user roles from the context. ctx is the context for the request. Returns the result. Returns true if successful.
//
// Summary: RolesFromContext returns the user roles from the context. ctx is the context for the request. Returns the result. Returns true if successful.
//
// Parameters:
//   - ctx (context.Context): The cancellation and deadline context.
//
// Returns:
//   - []string: The resulting text.
//   - bool: True if successful or valid, false otherwise.
//
// Errors:
//   - None.
//
// Side Effects:
//   - May modify internal state or perform external network calls.
//
// Errors:
//   - None.
//
// Side Effects:
// RBACEnforcer handles Role-Based Access Control checks.
//
// Summary: RBACEnforcer handles Role-Based Access Control checks.
func RolesFromContext(ctx context.Context) ([]string, bool) {
	val, ok := ctx.Value(RolesContextKey).([]string)
	return val, ok
// NewRBACEnforcer creates a new RBACEnforcer. Returns the result.
//
// Summary: NewRBACEnforcer creates a new RBACEnforcer. Returns the result.
//
// Parameters:
//   - None.
//
// Returns:
//   - *RBACEnforcer: The resulting object or data structure.
//
// Errors:
//   - None.
//
// Side Effects:
//   - May modify internal state or perform external network calls.
//   - None.
//
// Returns:
//   - *RBACEnforcer: The resulting object or data structure.
// HasRole checks if the given user has the specified role. user is the user. role is the role. Returns true if successful.
//
// Summary: HasRole checks if the given user has the specified role. user is the user. role is the role. Returns true if successful.
//
// Parameters:
//   - user (*configv1.User): The provided user data.
//   - role (string): The textual representation of role.
//
// Returns:
//   - bool: True if successful or valid, false otherwise.
//
// Errors:
//   - None.
//
// Side Effects:
//   - May modify internal state or perform external network calls.
// Parameters:
//   - user (*configv1.User): The provided user data.
//   - role (string): The textual representation of role.
//
// Returns:
//   - bool: True if successful or valid, false otherwise.
//
// HasAnyRole checks if the user has at least one of the specified roles. user is the user. roles is the roles. Returns true if successful.
//
// Summary: HasAnyRole checks if the user has at least one of the specified roles. user is the user. roles is the roles. Returns true if successful.
//
// Parameters:
//   - user (*configv1.User): The provided user data.
//   - roles ([]string): The textual representation of roles.
//
// Returns:
//   - bool: True if successful or valid, false otherwise.
//
// Errors:
//   - None.
//
// Side Effects:
//   - May modify internal state or perform external network calls.
// Summary: HasAnyRole checks if the user has at least one of the specified roles. user is the user. roles is the roles. Returns true if successful.
//
// Parameters:
//   - user (*configv1.User): The provided user data.
//   - roles ([]string): The textual representation of roles.
//
// Returns:
//   - bool: True if successful or valid, false otherwise.
//
// Errors:
//   - None.
//
// HasRoleInContext checks if the context contains the specified role. ctx is the context for the request. role is the role. Returns true if successful.
//
// Summary: HasRoleInContext checks if the context contains the specified role. ctx is the context for the request. role is the role. Returns true if successful.
//
// Parameters:
//   - ctx (context.Context): The cancellation and deadline context.
//   - role (string): The textual representation of role.
//
// Returns:
//   - bool: True if successful or valid, false otherwise.
//
// Errors:
//   - None.
//
// Side Effects:
//   - May modify internal state or perform external network calls.
// HasRoleInContext checks if the context contains the specified role. ctx is the context for the request. role is the role. Returns true if successful.
//
// Summary: HasRoleInContext checks if the context contains the specified role. ctx is the context for the request. role is the role. Returns true if successful.
//
// Parameters:
//   - ctx (context.Context): The cancellation and deadline context.
//   - role (string): The textual representation of role.
//
// Returns:
//   - bool: True if successful or valid, false otherwise.
//
// Errors:
//   - None.
//
// Side Effects:
//   - May modify internal state or perform external network calls.
func (e *RBACEnforcer) HasRoleInContext(ctx context.Context, role string) bool {
	roles, ok := RolesFromContext(ctx)
	if !ok {
		return false
	}
	return slices.Contains(roles, role)
}
