// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package index

import (
	"context"
	"testing"

	v1 "github.com/mcpany/core/proto/api/v1"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

func TestManager(t *testing.T) {
	m := NewManager()
	ctx := context.Background()

	// Seed
	tools := []*v1.IndexedTool{
		{Name: proto.String("Tool A"), Description: proto.String("Description A"), Tags: []string{"tag1"}},
		{Name: proto.String("Tool B"), Description: proto.String("Description B"), Tags: []string{"tag2"}},
	}
	count := m.Seed(ctx, tools, true)
	assert.Equal(t, int32(2), count)

	// Search All
	results, total := m.Search(ctx, "", 1, 10)
	assert.Equal(t, int32(2), total)
	assert.Len(t, results, 2)

	// Search Query
	results, total = m.Search(ctx, "Tool A", 1, 10)
	assert.Equal(t, int32(1), total)
	assert.Len(t, results, 1)
	assert.Equal(t, "Tool A", results[0].GetName())

	// Stats
	stats := m.GetStats(ctx)
	assert.Equal(t, int32(2), stats.GetTotalTools())
	assert.Equal(t, int32(2), stats.GetTotalSearches()) // 2 searches above
	assert.Equal(t, int32(2), stats.GetHits())          // 2 searches had hits
	assert.Equal(t, int32(0), stats.GetMisses())
}
