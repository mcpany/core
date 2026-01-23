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

func TestManager_BasicAuth_Security(t *testing.T) {
	// Setup
	password := "correct-password"
	hashedPassword, err := passhash.Password(password)
	require.NoError(t, err)

	user := &configv1.User{
		Id: proto.String("valid-user"),
		Authentication: &configv1.Authentication{
			AuthMethod: &configv1.Authentication_BasicAuth{
				BasicAuth: &configv1.BasicAuth{
					PasswordHash: proto.String(hashedPassword),
				},
			},
		},
	}

	am := NewManager()
	am.SetUsers([]*configv1.User{user})

	// Helper to make request
	checkAuth := func(username, pwd string) error {
		req := httptest.NewRequest("GET", "/", nil)
		req.SetBasicAuth(username, pwd)
		// We use a service ID that doesn't have a specific authenticator,
		// triggering the fallback to checkBasicAuthWithUsers
		_, err := am.Authenticate(context.Background(), "some-service", req)
		return err
	}

	t.Run("valid_user_valid_password", func(t *testing.T) {
		err := checkAuth("valid-user", "correct-password")
		assert.NoError(t, err)
	})

	t.Run("valid_user_invalid_password", func(t *testing.T) {
		err := checkAuth("valid-user", "wrong-password")
		assert.Error(t, err)
		// We don't check exact error message as it might be generic "unauthorized"
	})

	t.Run("invalid_user_any_password", func(t *testing.T) {
		err := checkAuth("invalid-user", "any-password")
		assert.Error(t, err)
	})
}
