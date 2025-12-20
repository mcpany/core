// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package mcpserver

import (
	"context"
	"testing"

	"github.com/mcpany/core/pkg/middleware"
	"github.com/mcpany/core/pkg/tool"
	v1 "github.com/mcpany/core/proto/api/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestCacheServer_ClearCache(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockManager := tool.NewMockManagerInterface(ctrl)
	cache := middleware.NewCachingMiddleware(mockManager)
	server := NewCacheServer(cache)

	req := &v1.ClearCacheRequest{}
	resp, err := server.ClearCache(context.Background(), req)

	require.NoError(t, err)
	assert.NotNil(t, resp)
}

func TestCacheServer_ClearCache_NilCache(t *testing.T) {
	server := NewCacheServer(nil)

	req := &v1.ClearCacheRequest{}
	_, err := server.ClearCache(context.Background(), req)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "caching is not enabled")
}
