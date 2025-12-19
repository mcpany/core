// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package admin

import (
	"context"

	"github.com/mcpany/core/pkg/middleware"
	pb "github.com/mcpany/api/proto/admin/v1"
)

// Server implements the AdminServiceServer interface.
type Server struct {
	pb.UnimplementedAdminServiceServer
	cache *middleware.CachingMiddleware
}

// NewServer creates a new Admin Server.
func NewServer(cache *middleware.CachingMiddleware) *Server {
	return &Server{cache: cache}
}

// ClearCache clears the cache.
func (s *Server) ClearCache(ctx context.Context, _ *pb.ClearCacheRequest) (*pb.ClearCacheResponse, error) {
	if err := s.cache.Clear(ctx); err != nil {
		return nil, err
	}
	return &pb.ClearCacheResponse{}, nil
}
