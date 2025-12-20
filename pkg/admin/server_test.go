// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package admin

import (
	"context"
	"testing"

	"github.com/mcpany/core/pkg/middleware"
	"github.com/mcpany/core/pkg/tool"
	pb "github.com/mcpany/core/proto/admin/v1"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestNewServer(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockManager := tool.NewMockManagerInterface(ctrl)
	cache := middleware.NewCachingMiddleware(mockManager)
	server := NewServer(cache)
	assert.NotNil(t, server)
}

func TestClearCache(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockManager := tool.NewMockManagerInterface(ctrl)

	// We use a real CachingMiddleware here since it is a struct and cannot be mocked easily.
	// The default implementation uses an in-memory cache which should succeed.
	cache := middleware.NewCachingMiddleware(mockManager)
	server := NewServer(cache)

	req := &pb.ClearCacheRequest{}
	resp, err := server.ClearCache(context.Background(), req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
}
