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

// TestCheckBasicAuthWithUsers_TimingMitigation verifies that the authentication logic
// correctly handles valid and invalid users, ensuring that the timing attack mitigation
// (dummy hash verification) doesn't break functional correctness.
func TestCheckBasicAuthWithUsers_TimingMitigation(t *testing.T) {
	// Setup
	authManager := NewManager()
	password := "securePass123"
	hashed, err := passhash.Password(password)
	require.NoError(t, err)

	users := []*configv1.User{
		configv1.User_builder{
			Id: proto.String("validUser"),
			Authentication: configv1.Authentication_builder{
				BasicAuth: configv1.BasicAuth_builder{
					Username:     proto.String("validUser"),
					PasswordHash: proto.String(hashed),
				}.Build(),
			}.Build(),
		}.Build(),
	}
	authManager.SetUsers(users)

	t.Run("ValidUser_CorrectPassword", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		req.SetBasicAuth("validUser", password)
		ctx, err := authManager.checkBasicAuthWithUsers(context.Background(), req)
		assert.NoError(t, err)

		userID, ok := UserFromContext(ctx)
		assert.True(t, ok)
		assert.Equal(t, "validUser", userID)
	})

	t.Run("ValidUser_WrongPassword", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		req.SetBasicAuth("validUser", "wrongPass")
		_, err := authManager.checkBasicAuthWithUsers(context.Background(), req)
		assert.Error(t, err)
		assert.Equal(t, "invalid credentials", err.Error())
	})

	t.Run("NonExistentUser", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		req.SetBasicAuth("ghost", "anyPass")

		// This should trigger the dummy hash path
		_, err := authManager.checkBasicAuthWithUsers(context.Background(), req)
		assert.Error(t, err)
		assert.Equal(t, "invalid credentials", err.Error())
	})

	t.Run("ValidUser_NoBasicAuth", func(t *testing.T) {
		// User exists but has no basic auth configured
		noAuthUser := []*configv1.User{
			configv1.User_builder{
				Id: proto.String("oauthUser"),
				// No Authentication field
			}.Build(),
		}
		authManager.SetUsers(noAuthUser)

		req := httptest.NewRequest("GET", "/", nil)
		req.SetBasicAuth("oauthUser", "anyPass")

		// This should also trigger the dummy hash path
		_, err := authManager.checkBasicAuthWithUsers(context.Background(), req)
		assert.Error(t, err)
		assert.Equal(t, "invalid credentials", err.Error())
	})
}
