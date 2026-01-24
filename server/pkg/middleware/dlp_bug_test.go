// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/mcpany/core/server/pkg/logging"
	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
)

func TestDLPMiddleware_BugReproduction(t *testing.T) {
	logger := logging.GetLogger()
	enabled := true
	cfg := &configv1.DLPConfig{
		Enabled:        &enabled,
		CustomPatterns: []string{`secret`},
	}

	mw := DLPMiddleware(cfg, logger)

	t.Run("RedactArguments_CommentsBug", func(t *testing.T) {
		handler := func(ctx context.Context, method string, req mcp.Request) (mcp.Result, error) {
			callReq := req.(*mcp.CallToolRequest)
			// We check the raw bytes because Unmarshal might fail on comments regardless
			raw := string(callReq.Params.Arguments)
			// Expectation: The comment should NOT be redacted.
			// Bug behavior: It IS redacted.
			assert.Contains(t, raw, `"secret"`)
			assert.NotContains(t, raw, "***REDACTED***")
			return &mcp.CallToolResult{}, nil
		}

		wrapped := mw(handler)

		// Input with a stray slash followed by a comment containing a secret
		// The bug causes the walker to ignore the comment start and visit "secret"
		inputJSON := `/ // "secret"`
		req := &mcp.CallToolRequest{
			Params: &mcp.CallToolParamsRaw{
				Name: "tools/call",
				Arguments: json.RawMessage(inputJSON),
			},
		}

		_, _ = wrapped(context.Background(), "tools/call", req)
	})
}
