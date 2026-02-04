// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package dashboard

import (
	"context"
	"testing"

	pb "github.com/mcpany/core/proto/api/v1"
	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/auth"
	"github.com/mcpany/core/server/pkg/storage/memory"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

func TestServer(t *testing.T) {
	// Setup
	s := memory.NewStore()
	defer s.Close()
	srv := NewServer(s)
	ctx := context.Background()

	// Create user
	userID := "test-user"
	user := configv1.User_builder{
		Id: proto.String(userID),
	}.Build()
	err := s.CreateUser(ctx, user)
	assert.NoError(t, err)

	// Auth context
	ctx = auth.ContextWithUser(ctx, userID)

	// Test Save
	layout := `[{"id":"1"}]`
	_, err = srv.SaveDashboardLayout(ctx, &pb.SaveDashboardLayoutRequest{
		LayoutJson: layout,
	})
	assert.NoError(t, err)

	// Test Get
	resp, err := srv.GetDashboardLayout(ctx, &pb.GetDashboardLayoutRequest{})
	assert.NoError(t, err)
	assert.Equal(t, layout, resp.LayoutJson)
}
