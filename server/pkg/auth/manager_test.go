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
	svc := &configv1.UpstreamServiceConfig{
		Name: proto.String("svc"),
		UpstreamAuth: &configv1.Authentication{
			AuthMethod: &configv1.Authentication_Oauth2{
				Oauth2: &configv1.OAuth2Auth{
					ClientId:         &configv1.SecretValue{Value: &configv1.SecretValue_PlainText{PlainText: "id"}},
					ClientSecret:     &configv1.SecretValue{Value: &configv1.SecretValue_PlainText{PlainText: "secret"}},
					AuthorizationUrl: proto.String("http://auth"),
					TokenUrl:         proto.String("http://token"),
				},
			},
		},
	}
	require.NoError(t, store.SaveService(ctx, svc))

	_, _, err := am.InitiateOAuth(ctx, "u", "svc", "", "http://cb")
	assert.NoError(t, err)

	// Test SetUsers and GetUser
	users := []*configv1.User{
		{
			Id: proto.String("user1"),
			Authentication: &configv1.Authentication{
				AuthMethod: &configv1.Authentication_BasicAuth{
					BasicAuth: &configv1.BasicAuth{
						PasswordHash: proto.String("hash"),
					},
				},
			},
		},
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
