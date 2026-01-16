// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"context"
	"strings"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

func TestContextOptimizerMiddleware(t *testing.T) {
	// Create optimizer with max chars 10
	config := &configv1.ContextOptimizerConfig{
		MaxChars: proto.Int32(10),
	}
	opt := NewContextOptimizer(config)

	// Mock next handler
	longText := "This text is definitely longer than 10 characters"
	shortText := "Short"

	// Case 1: Long Response
	nextLong := func(ctx context.Context, method string, req mcp.Request) (mcp.Result, error) {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{
					Text: longText,
				},
			},
		}, nil
	}

	handler := opt.Middleware(nextLong)
	res, err := handler(context.Background(), "tools/call", &mcp.CallToolRequest{})
	assert.NoError(t, err)
	ctr, ok := res.(*mcp.CallToolResult)
	assert.True(t, ok)
	text := ctr.Content[0].(*mcp.TextContent).Text
	assert.True(t, strings.Contains(text, "TRUNCATED"), "Should contain truncated message")
	assert.True(t, len(text) < len(longText), "Should be truncated")

	// Case 2: Short Response
	nextShort := func(ctx context.Context, method string, req mcp.Request) (mcp.Result, error) {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{
					Text: shortText,
				},
			},
		}, nil
	}

	handler = opt.Middleware(nextShort)
	res, err = handler(context.Background(), "tools/call", &mcp.CallToolRequest{})
	assert.NoError(t, err)
	ctr, ok = res.(*mcp.CallToolResult)
	assert.True(t, ok)
	text = ctr.Content[0].(*mcp.TextContent).Text
	assert.Equal(t, shortText, text)
}
