// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package admin

import (
	"context"
	"testing"

	"github.com/mcpany/core/pkg/middleware"
	pb "github.com/mcpany/core/proto/admin/v1"
	"github.com/stretchr/testify/assert"
)

func TestServer_ClearCache(t *testing.T) {
	// Create a real caching middleware (tool manager can be nil as it's not used in Clear)
	cacheMiddleware := middleware.NewCachingMiddleware(nil)
	server := NewServer(cacheMiddleware)

	resp, err := server.ClearCache(context.Background(), &pb.ClearCacheRequest{})
	assert.NoError(t, err)
	assert.NotNil(t, resp)
}
