// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"context"
	"strings"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/storage/memory"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func TestHashUserPassword_SecretValue(t *testing.T) {
	app := NewApplication()
	_ = app
	store := memory.NewStore()

	// 1. Create a user with password in SecretValue (PlainText) using the builder
	user := configv1.User_builder{
		Id: proto.String("user-secret"),
		Authentication: configv1.Authentication_builder{
			BasicAuth: configv1.BasicAuth_builder{
				Password: configv1.SecretValue_builder{
					PlainText: proto.String("secret123"),
				}.Build(),
			}.Build(),
		}.Build(),
	}.Build()

	// 2. Call hashUserPassword
	err := hashUserPassword(context.Background(), user, store)
	require.NoError(t, err)

	// 3. Verify
	basicAuth := user.GetAuthentication().GetBasicAuth()
	// PasswordHash should be set and valid
	assert.NotEmpty(t, basicAuth.GetPasswordHash())
	assert.True(t, strings.HasPrefix(basicAuth.GetPasswordHash(), "$2"), "should be a bcrypt hash: "+basicAuth.GetPasswordHash())

	// Password field should be cleared (nil)
	assert.Nil(t, basicAuth.GetPassword())
}
