// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package auth

import (
	"context"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
)

func TestRBACEnforcer_HasRole(t *testing.T) {
	enforcer := NewRBACEnforcer()

	t.Run("nil user", func(t *testing.T) {
		assert.False(t, enforcer.HasRole(nil, "admin"))
	})

	t.Run("user without roles", func(t *testing.T) {
		user := configv1.User_builder{
			Roles: []string{},
		}.Build()
		assert.False(t, enforcer.HasRole(user, "admin"))
	})

	t.Run("user with different role", func(t *testing.T) {
		user := configv1.User_builder{
			Roles: []string{"user"},
		}.Build()
		assert.False(t, enforcer.HasRole(user, "admin"))
	})

	t.Run("user with required role", func(t *testing.T) {
		user := configv1.User_builder{
			Roles: []string{"user", "admin"},
		}.Build()
		assert.True(t, enforcer.HasRole(user, "admin"))
	})
}

func TestRBACEnforcer_HasAnyRole(t *testing.T) {
	enforcer := NewRBACEnforcer()

	t.Run("nil user", func(t *testing.T) {
		assert.False(t, enforcer.HasAnyRole(nil, []string{"admin"}))
	})

	t.Run("user without matching roles", func(t *testing.T) {
		user := configv1.User_builder{
			Roles: []string{"user"},
		}.Build()
		assert.False(t, enforcer.HasAnyRole(user, []string{"admin", "superuser"}))
	})

	t.Run("user with one matching role", func(t *testing.T) {
		user := configv1.User_builder{
			Roles: []string{"admin"},
		}.Build()
		assert.True(t, enforcer.HasAnyRole(user, []string{"admin", "superuser"}))
	})

	t.Run("empty required roles", func(t *testing.T) {
		user := configv1.User_builder{
			Roles: []string{"admin"},
		}.Build()
		assert.False(t, enforcer.HasAnyRole(user, []string{}))
	})
}

func TestContextWithRoles(t *testing.T) {
	ctx := context.Background()
	roles := []string{"admin", "user"}
	ctxWithRoles := ContextWithRoles(ctx, roles)

	retrievedRoles, ok := RolesFromContext(ctxWithRoles)
	assert.True(t, ok)
	assert.Equal(t, roles, retrievedRoles)

	_, ok = RolesFromContext(ctx)
	assert.False(t, ok)
}

func TestRBACEnforcer_HasRoleInContext(t *testing.T) {
	enforcer := NewRBACEnforcer()
	ctx := context.Background()

	t.Run("no roles in context", func(t *testing.T) {
		assert.False(t, enforcer.HasRoleInContext(ctx, "admin"))
	})

	t.Run("role not present", func(t *testing.T) {
		ctxWithRoles := ContextWithRoles(ctx, []string{"user"})
		assert.False(t, enforcer.HasRoleInContext(ctxWithRoles, "admin"))
	})

	t.Run("role present", func(t *testing.T) {
		ctxWithRoles := ContextWithRoles(ctx, []string{"admin", "user"})
		assert.True(t, enforcer.HasRoleInContext(ctxWithRoles, "admin"))
	})
}
