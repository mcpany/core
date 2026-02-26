// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package rest

import (
	"context"

	apiv1 "github.com/mcpany/core/proto/api/v1"
	"github.com/mcpany/core/server/pkg/index"
	"google.golang.org/protobuf/proto"
)

// IndexServer implements the IndexService API.
//
// Summary: Server implementation for the Index Service.
type IndexServer struct {
	manager *index.Manager
}

// NewIndexServer creates a new IndexServer.
//
// Summary: Initializes a new IndexServer.
//
// Parameters:
//   - manager (*index.Manager): The index manager instance.
//
// Returns:
//   - (*IndexServer): The initialized server instance.
func NewIndexServer(manager *index.Manager) *IndexServer {
	return &IndexServer{manager: manager}
}

// Search searches the index for tools matching the query.
func (s *IndexServer) Search(ctx context.Context, req *apiv1.SearchIndexRequest) (*apiv1.SearchIndexResponse, error) {
	tools, total := s.manager.Search(ctx, req.GetQuery(), req.GetPage(), req.GetLimit())
	return &apiv1.SearchIndexResponse{
		Tools: tools,
		Total: proto.Int32(total),
	}, nil
}

// Seed seeds the index with tool definitions.
func (s *IndexServer) Seed(ctx context.Context, req *apiv1.SeedIndexRequest) (*apiv1.SeedIndexResponse, error) {
	count := s.manager.Seed(ctx, req.GetTools(), req.GetClear())
	return &apiv1.SeedIndexResponse{
		Count: proto.Int32(count),
	}, nil
}

// GetStats returns statistics about the index usage.
func (s *IndexServer) GetStats(ctx context.Context, _ *apiv1.GetIndexStatsRequest) (*apiv1.GetIndexStatsResponse, error) {
	return s.manager.GetStats(ctx), nil
}
