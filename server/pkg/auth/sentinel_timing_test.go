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

// TestCheckBasicAuthWithUsers_Timing_Negative ensures that both invalid users and
// valid users with wrong passwords result in the same error, and that the code path
// for invalid users is exercised (by not panicking and returning the expected error).
//
// While we cannot strictly enforce constant time in a unit test without flakiness,
// this test verifies that the refactored logic is functional.
func TestCheckBasicAuthWithUsers_Timing_Negative(t *testing.T) {
	authManager := NewManager()
	require.NotNil(t, authManager)

	password := "secret123"
	hashed, _ := passhash.Password(password)

	// Create a valid user
	users := []*configv1.User{
		configv1.User_builder{
			Id:    proto.String("validuser"),
			Authentication: configv1.Authentication_builder{
				BasicAuth: configv1.BasicAuth_builder{
					Username:     proto.String("validuser"),
					PasswordHash: proto.String(hashed),
				}.Build(),
			}.Build(),
		}.Build(),
	}
	authManager.SetUsers(users)

	t.Run("invalid_username_should_fail", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		req.SetBasicAuth("invaliduser", "anypassword")

		_, err := authManager.checkBasicAuthWithUsers(context.Background(), req)
		assert.Error(t, err)
		assert.Equal(t, "invalid credentials", err.Error())
	})

	t.Run("valid_username_wrong_password_should_fail", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		req.SetBasicAuth("validuser", "wrongpassword")

		_, err := authManager.checkBasicAuthWithUsers(context.Background(), req)
		assert.Error(t, err)
		assert.Equal(t, "invalid credentials", err.Error())
	})

	t.Run("valid_username_correct_password_should_succeed", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		req.SetBasicAuth("validuser", password)

		ctx, err := authManager.checkBasicAuthWithUsers(context.Background(), req)
		assert.NoError(t, err)

		userID, ok := UserFromContext(ctx)
		assert.True(t, ok)
		assert.Equal(t, "validuser", userID)
	})

	t.Run("user_exists_but_no_basic_auth_should_fail", func(t *testing.T) {
		// Add a user without basic auth
		noAuthUser := []*configv1.User{
			configv1.User_builder{
				Id: proto.String("noauthuser"),
				// No Authentication field
			}.Build(),
		}
		authManager.SetUsers(noAuthUser)

		req := httptest.NewRequest("GET", "/", nil)
		req.SetBasicAuth("noauthuser", "anypassword")

		_, err := authManager.checkBasicAuthWithUsers(context.Background(), req)
		assert.Error(t, err)
		assert.Equal(t, "invalid credentials", err.Error())
	})
}
