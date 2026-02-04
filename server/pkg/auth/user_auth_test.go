// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package auth

import (
	"context"
	"net/http/httptest"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/util/passhash"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func TestUserAuthenticationWithProfiles(t *testing.T) {
	authManager := NewManager()

	password := "secret123"
	hashed, _ := passhash.Password(password)

	users := []*configv1.User{
		configv1.User_builder{
			Id:         proto.String("alice"),
			Roles:      []string{"admin"},
			ProfileIds: []string{"prod"},
			Authentication: configv1.Authentication_builder{
				BasicAuth: configv1.BasicAuth_builder{
					Username:     proto.String("alice"),
					PasswordHash: proto.String(hashed),
				}.Build(),
			}.Build(),
		}.Build(),
		configv1.User_builder{
			Id:         proto.String("bob"),
			Roles:      []string{"dev"},
			ProfileIds: []string{"dev"},
			Authentication: configv1.Authentication_builder{
				ApiKey: configv1.APIKeyAuth_builder{
					ParamName:         proto.String("X-API-Key"),
					VerificationValue: proto.String("bob-api-key"),
					In:                configv1.APIKeyAuth_HEADER.Enum(),
				}.Build(),
			}.Build(),
		}.Build(),
	}
	authManager.SetUsers(users)

	t.Run("basic_auth_sets_profile", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		req.SetBasicAuth("alice", password)

		// Authenticate with NO specific service (fallback to global/user)
		ctx, err := authManager.Authenticate(context.Background(), "", req)
		require.NoError(t, err)

		uid, ok := UserFromContext(ctx)
		assert.True(t, ok)
		assert.Equal(t, "alice", uid)

		roles, ok := RolesFromContext(ctx)
		assert.True(t, ok)
		assert.Contains(t, roles, "admin")

		pid, ok := ProfileIDFromContext(ctx)
		assert.True(t, ok)
		assert.Equal(t, "prod", pid)
	})

	t.Run("api_key_auth_sets_profile", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		req.Header.Set("X-API-Key", "bob-api-key")

		// Authenticate with NO specific service (fallback to global/user)
		ctx, err := authManager.Authenticate(context.Background(), "", req)
		require.NoError(t, err)

		uid, ok := UserFromContext(ctx)
		assert.True(t, ok)
		assert.Equal(t, "bob", uid)

		roles, ok := RolesFromContext(ctx)
		assert.True(t, ok)
		assert.Contains(t, roles, "dev")

		pid, ok := ProfileIDFromContext(ctx)
		assert.True(t, ok)
		assert.Equal(t, "dev", pid)
	})

	t.Run("api_key_auth_invalid", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		req.Header.Set("X-API-Key", "wrong-key")

		_, err := authManager.Authenticate(context.Background(), "", req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unauthorized")
	})
}
