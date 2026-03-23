// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"context"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
)

func TestLazyMCPMiddleware(t *testing.T) {
	config := LazyMCPConfig{
		Enabled:   true,
		Threshold: 0.85,
	}
	m := NewLazyMCPMiddleware(config)

	req := &mcp.ListToolsRequest{}

	next := func(ctx context.Context, method string, req mcp.Request) (mcp.Result, error) {
		return &mcp.ListToolsResult{
			Tools: []*mcp.Tool{
				&mcp.Tool{Name: "tool1"},
				&mcp.Tool{Name: "tool2"},
			},
		}, nil
	}

	res, err := m.Execute(context.Background(), "tools/list", req, next)
	assert.NoError(t, err)
	listRes, ok := res.(*mcp.ListToolsResult)
	assert.True(t, ok)
	assert.Equal(t, 2, len(listRes.Tools))
}
