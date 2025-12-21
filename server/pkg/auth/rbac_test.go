// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package auth

import (
	"context"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
)

func TestContextWithRoles(t *testing.T) {
	ctx := context.Background()
	roles := []string{"admin", "user"}

	newCtx := ContextWithRoles(ctx, roles)

	retrievedRoles, ok := RolesFromContext(newCtx)
	assert.True(t, ok)
	assert.Equal(t, roles, retrievedRoles)
}

func TestRolesFromContext_Empty(t *testing.T) {
	ctx := context.Background()
	roles, ok := RolesFromContext(ctx)
	assert.False(t, ok)
	assert.Nil(t, roles)
}

func TestRBACEnforcer_HasRole(t *testing.T) {
	enforcer := NewRBACEnforcer()

	t.Run("nil_user", func(t *testing.T) {
		assert.False(t, enforcer.HasRole(nil, "admin"))
	})

	t.Run("user_has_role", func(t *testing.T) {
		user := &configv1.User{
			Roles: []string{"admin", "user"},
		}
		assert.True(t, enforcer.HasRole(user, "admin"))
	})

	t.Run("user_missing_role", func(t *testing.T) {
		user := &configv1.User{
			Roles: []string{"user"},
		}
		assert.False(t, enforcer.HasRole(user, "admin"))
	})
}

func TestRBACEnforcer_HasAnyRole(t *testing.T) {
	enforcer := NewRBACEnforcer()

	t.Run("nil_user", func(t *testing.T) {
		assert.False(t, enforcer.HasAnyRole(nil, []string{"admin"}))
	})

	t.Run("user_has_one_role", func(t *testing.T) {
		user := &configv1.User{
			Roles: []string{"user"},
		}
		assert.True(t, enforcer.HasAnyRole(user, []string{"admin", "user"}))
	})

	t.Run("user_has_none_role", func(t *testing.T) {
		user := &configv1.User{
			Roles: []string{"guest"},
		}
		assert.False(t, enforcer.HasAnyRole(user, []string{"admin", "user"}))
	})
}

func TestRBACEnforcer_HasRoleInContext(t *testing.T) {
	enforcer := NewRBACEnforcer()
	ctx := context.Background()

	t.Run("no_roles_in_context", func(t *testing.T) {
		assert.False(t, enforcer.HasRoleInContext(ctx, "admin"))
	})

	t.Run("has_role_in_context", func(t *testing.T) {
		roles := []string{"admin", "user"}
		ctxWithRoles := ContextWithRoles(ctx, roles)
		assert.True(t, enforcer.HasRoleInContext(ctxWithRoles, "admin"))
	})

	t.Run("missing_role_in_context", func(t *testing.T) {
		roles := []string{"user"}
		ctxWithRoles := ContextWithRoles(ctx, roles)
		assert.False(t, enforcer.HasRoleInContext(ctxWithRoles, "admin"))
	})
}
