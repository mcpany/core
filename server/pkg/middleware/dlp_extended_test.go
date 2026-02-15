package middleware_test

import (
	"context"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/logging"
	"github.com/mcpany/core/server/pkg/middleware"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func TestDLPMiddleware_GetPrompt(t *testing.T) {
	config := configv1.DLPConfig_builder{
		Enabled: proto.Bool(true),
	}.Build()
	logger := logging.GetLogger()
	mw := middleware.DLPMiddleware(config, logger)

	handler := func(_ context.Context, _ string, req mcp.Request) (mcp.Result, error) {
		// Verify request arguments are redacted
		if r, ok := req.(*mcp.GetPromptRequest); ok {
			args := r.Params.Arguments
			if args["email"] != "***REDACTED***" {
				return nil, nil // Fail if not redacted (handled by assertion below)
			}
		}

		// Return a result with PII
		return &mcp.GetPromptResult{
			Messages: []*mcp.PromptMessage{
				{
					Content: &mcp.TextContent{
						Text: "Here is the email: test@example.com",
					},
				},
			},
		}, nil
	}

	wrapped := mw(handler)

	req := &mcp.GetPromptRequest{
		Params: &mcp.GetPromptParams{
			Name: "test-prompt",
			Arguments: map[string]string{
				"email": "user@example.com",
			},
		},
	}

	res, err := wrapped(context.Background(), "prompts/get", req)
	require.NoError(t, err)

	// Verify request redaction
	assert.Equal(t, "***REDACTED***", req.Params.Arguments["email"])

	// Verify result redaction
	pRes, ok := res.(*mcp.GetPromptResult)
	require.True(t, ok)
	require.Len(t, pRes.Messages, 1)
	textContent, ok := pRes.Messages[0].Content.(*mcp.TextContent)
	require.True(t, ok)
	assert.Equal(t, "Here is the email: ***REDACTED***", textContent.Text)
}

func TestDLPMiddleware_ReadResource(t *testing.T) {
	config := configv1.DLPConfig_builder{
		Enabled: proto.Bool(true),
	}.Build()
	logger := logging.GetLogger()
	mw := middleware.DLPMiddleware(config, logger)

	handler := func(_ context.Context, _ string, req mcp.Request) (mcp.Result, error) {
		return &mcp.ReadResourceResult{
			Contents: []*mcp.ResourceContents{
				{
					URI:  "file:///test",
					Text: "Found SSN: 123-45-6789 in file",
				},
			},
		}, nil
	}

	wrapped := mw(handler)

	req := &mcp.ReadResourceRequest{
		Params: &mcp.ReadResourceParams{
			URI: "file:///test",
		},
	}

	res, err := wrapped(context.Background(), "resources/read", req)
	require.NoError(t, err)

	// Verify result redaction
	rRes, ok := res.(*mcp.ReadResourceResult)
	require.True(t, ok)
	require.Len(t, rRes.Contents, 1)
	assert.Equal(t, "Found SSN: ***REDACTED*** in file", rRes.Contents[0].Text)
}
