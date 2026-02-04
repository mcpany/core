// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package admin

import (
	"context"
	"testing"

	pb "github.com/mcpany/core/proto/admin/v1"
	"github.com/mcpany/core/server/pkg/storage/memory"
	"github.com/mcpany/core/server/pkg/tool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"google.golang.org/protobuf/proto"
)

func TestServer_UserPreferences(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	store := memory.NewStore()
	tm := tool.NewMockManagerInterface(ctrl)
	sr := &MockServiceRegistry{}

	s := NewServer(nil, tm, sr, store, nil, nil)
	ctx := context.Background()

	// 1. Get empty preferences
	getResp, err := s.GetUserPreferences(ctx, pb.GetUserPreferencesRequest_builder{UserId: proto.String("user1")}.Build())
	require.NoError(t, err)
	assert.Empty(t, getResp.GetPreferences())

	// 2. Update preferences
	prefs := map[string]string{
		"theme": "dark",
		"layout": "grid",
	}
	updateResp, err := s.UpdateUserPreferences(ctx, pb.UpdateUserPreferencesRequest_builder{
		UserId:      proto.String("user1"),
		Preferences: prefs,
	}.Build())
	require.NoError(t, err)
	assert.Equal(t, "dark", updateResp.GetPreferences()["theme"])
	assert.Equal(t, "grid", updateResp.GetPreferences()["layout"])

	// 3. Get updated preferences
	getResp, err = s.GetUserPreferences(ctx, pb.GetUserPreferencesRequest_builder{UserId: proto.String("user1")}.Build())
	require.NoError(t, err)
	assert.Equal(t, "dark", getResp.GetPreferences()["theme"])
	assert.Equal(t, "grid", getResp.GetPreferences()["layout"])

	// 4. Update partial
	updateResp, err = s.UpdateUserPreferences(ctx, pb.UpdateUserPreferencesRequest_builder{
		UserId: proto.String("user1"),
		Preferences: map[string]string{
			"theme": "light",
		},
	}.Build())
	require.NoError(t, err)
	assert.Equal(t, "light", updateResp.GetPreferences()["theme"])
	assert.Equal(t, "grid", updateResp.GetPreferences()["layout"]) // Should persist

	// 5. Default user ID handling
	updateResp, err = s.UpdateUserPreferences(ctx, pb.UpdateUserPreferencesRequest_builder{
		Preferences: map[string]string{"foo": "bar"},
	}.Build())
	require.NoError(t, err)

	// Verify default user gets it
	getResp, err = s.GetUserPreferences(ctx, pb.GetUserPreferencesRequest_builder{UserId: proto.String("default")}.Build())
	require.NoError(t, err)
	assert.Equal(t, "bar", getResp.GetPreferences()["foo"])
}
