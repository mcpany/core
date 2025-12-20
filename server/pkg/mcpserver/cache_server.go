// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package mcpserver

import (
	"context"

	"github.com/mcpany/core/pkg/middleware"
	v1 "github.com/mcpany/core/proto/api/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// CacheServer implements the CacheServiceServer interface.
type CacheServer struct {
	v1.UnimplementedCacheServiceServer
	cache *middleware.CachingMiddleware
}

// NewCacheServer creates a new CacheServer.
func NewCacheServer(cache *middleware.CachingMiddleware) *CacheServer {
	return &CacheServer{cache: cache}
}

// ClearCache clears the cache.
func (s *CacheServer) ClearCache(ctx context.Context, _ *v1.ClearCacheRequest) (*v1.ClearCacheResponse, error) {
	if s.cache == nil {
		return nil, status.Error(codes.FailedPrecondition, "caching is not enabled")
	}
	if err := s.cache.Clear(ctx); err != nil {
		return nil, err
	}
	return &v1.ClearCacheResponse{}, nil
}
