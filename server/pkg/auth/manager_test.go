// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package auth

import (
	"context"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/storage/memory"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func TestManager_SettersAndGetters(t *testing.T) {
	am := NewManager()

	// Test SetStorage
	store := memory.NewStore()
	am.SetStorage(store)
	// We can't verify private field easily without reflection or exposing getter, or testing behavior.
	// InitiateOAuth fails if storage is nil, succeeds if set.
	ctx := context.Background()
	svc := configv1.UpstreamServiceConfig_builder{
		Name: proto.String("svc"),
		UpstreamAuth: configv1.Authentication_builder{
			Oauth2: configv1.OAuth2Auth_builder{
				ClientId:         configv1.SecretValue_builder{PlainText: proto.String("id")}.Build(),
				ClientSecret:     configv1.SecretValue_builder{PlainText: proto.String("secret")}.Build(),
				AuthorizationUrl: proto.String("http://auth"),
				TokenUrl:         proto.String("http://token"),
			}.Build(),
		}.Build(),
	}.Build()
	require.NoError(t, store.SaveService(ctx, svc))

	_, _, err := am.InitiateOAuth(ctx, "u", "svc", "", "http://cb")
	assert.NoError(t, err)

	// Test SetUsers and GetUser
	users := []*configv1.User{
		configv1.User_builder{
			Id: proto.String("user1"),
			Authentication: configv1.Authentication_builder{
				BasicAuth: configv1.BasicAuth_builder{
					PasswordHash: proto.String("hash"),
				}.Build(),
			}.Build(),
		}.Build(),
	}
	am.SetUsers(users)

	u, ok := am.GetUser("user1")
	require.True(t, ok)
	assert.NotNil(t, u)
	assert.Equal(t, "user1", u.GetId())

	u2, ok := am.GetUser("unknown")
	assert.False(t, ok)
	assert.Nil(t, u2)
}
