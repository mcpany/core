// Copyright 2026 Author(s) of MCP Any
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

// TestDLPMiddleware_BugCommentRepro verifies that strings inside comments are NOT redacted.
// This uses the middleware stack to simulate an End-to-End flow for argument redaction.
func TestDLPMiddleware_BugCommentRepro(t *testing.T) {
	logger := logging.GetLogger()
	enabled := true
	// Configure DLP to redact "secret-123"
	cfg := &configv1.DLPConfig{
		Enabled:        &enabled,
		CustomPatterns: []string{`secret-\d+`},
	}

	mw := DLPMiddleware(cfg, logger)

	handler := func(ctx context.Context, method string, req mcp.Request) (mcp.Result, error) {
		callReq := req.(*mcp.CallToolRequest)

		// We inspect the raw bytes of arguments to see if redaction occurred in comments
		raw := string(callReq.Params.Arguments)
		t.Logf("Received args: %s", raw)

		// Verification:
		// The string "secret-123" inside the comment should NOT be redacted.
		assert.Contains(t, raw, "secret-123", "Secret in comment should be preserved")
		assert.NotContains(t, raw, "***REDACTED***", "Should not redact inside comments")

		return &mcp.CallToolResult{}, nil
	}

	wrapped := mw(handler)

	// Construct a request with a comment containing the secret, preceded by a division operator
	// which triggers the bug in WalkJSONStrings.
	// Input: `[ 1/2, /* "secret-123" */ "visible" ]`
	jsonInput := `[ 1/2, /* "secret-123" */ "visible" ]`

	req := &mcp.CallToolRequest{
		Params: &mcp.CallToolParamsRaw{
			Name: "tools/call",
			Arguments: json.RawMessage(jsonInput),
		},
	}

	_, err := wrapped(context.Background(), "tools/call", req)
	assert.NoError(t, err)
}
