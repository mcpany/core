package middleware

import (
	"context"
	"encoding/json"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/logging"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
)

func TestDLPMiddleware(t *testing.T) {
	logger := logging.GetLogger()
	enabled := true
	cfg := configv1.DLPConfig_builder{
		Enabled:        &enabled,
		CustomPatterns: []string{`secret-\d+`},
	}.Build()

	mw := DLPMiddleware(cfg, logger)

	t.Run("RedactArguments", func(t *testing.T) {
		handler := func(ctx context.Context, method string, req mcp.Request) (mcp.Result, error) {
			// Check if arguments were redacted
			callReq := req.(*mcp.CallToolRequest)

			var args map[string]interface{}
			err := json.Unmarshal(callReq.Params.Arguments, &args)
			assert.NoError(t, err)

			email := args["email"].(string)
			assert.Equal(t, "***REDACTED***", email)

			nested := args["nested"].(map[string]interface{})
			custom := nested["custom"].(string)
			assert.Equal(t, "***REDACTED***", custom)
			return &mcp.CallToolResult{}, nil
		}

		wrapped := mw(handler)

		req := &mcp.CallToolRequest{
			Params: &mcp.CallToolParamsRaw{
				Name: "tools/call",
				Arguments: json.RawMessage(`{
					"email": "user@example.com",
					"normal": "hello",
					"nested": {
						"custom": "secret-123"
					}
				}`),
			},
		}

		_, err := wrapped(context.Background(), "tools/call", req)
		assert.NoError(t, err)
	})

	t.Run("OptimizationPath", func(t *testing.T) {
		// Test the optimization path (no custom patterns)
		cfg := configv1.DLPConfig_builder{
			Enabled: &enabled,
			// No CustomPatterns
		}.Build()
		mwOpt := DLPMiddleware(cfg, logger)

		handler := func(ctx context.Context, method string, req mcp.Request) (mcp.Result, error) {
			callReq := req.(*mcp.CallToolRequest)
			var args map[string]interface{}
			err := json.Unmarshal(callReq.Params.Arguments, &args)
			assert.NoError(t, err)

			// Email should be redacted (contains @)
			email := args["email"].(string)
			assert.Equal(t, "***REDACTED***", email)

			// Safe string should NOT be redacted (and should skip unmarshal in optimization)
			safe := args["safe"].(string)
			assert.Equal(t, "safe string", safe)

			return &mcp.CallToolResult{}, nil
		}

		wrapped := mwOpt(handler)
		req := &mcp.CallToolRequest{
			Params: &mcp.CallToolParamsRaw{
				Arguments: json.RawMessage(`{
					"email": "user@example.com",
					"safe": "safe string"
				}`),
			},
		}
		_, err := wrapped(context.Background(), "tools/call", req)
		assert.NoError(t, err)
	})

	t.Run("RedactResult", func(t *testing.T) {
		handler := func(ctx context.Context, method string, req mcp.Request) (mcp.Result, error) {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: "Here is a credit card: 1234 5678 1234 5678"},
				},
			}, nil
		}

		wrapped := mw(handler)

		res, err := wrapped(context.Background(), "tools/call", &mcp.CallToolRequest{})
		assert.NoError(t, err)

		callRes := res.(*mcp.CallToolResult)
		text := callRes.Content[0].(*mcp.TextContent).Text
		assert.Contains(t, text, "***REDACTED***")
		assert.NotContains(t, text, "1234 5678")
	})

	t.Run("Disabled", func(t *testing.T) {
		disabled := false
		cfg := configv1.DLPConfig_builder{
			Enabled: &disabled,
		}.Build()
		mwDisabled := DLPMiddleware(cfg, logger)

		handler := func(ctx context.Context, method string, req mcp.Request) (mcp.Result, error) {
			callReq := req.(*mcp.CallToolRequest)

			var args map[string]interface{}
			err := json.Unmarshal(callReq.Params.Arguments, &args)
			assert.NoError(t, err)

			assert.Equal(t, "user@example.com", args["email"])
			return &mcp.CallToolResult{}, nil
		}

		wrapped := mwDisabled(handler)
		req := &mcp.CallToolRequest{
			Params: &mcp.CallToolParamsRaw{
				Arguments: json.RawMessage(`{"email": "user@example.com"}`),
			},
		}
		_, err := wrapped(context.Background(), "tools/call", req)
		assert.NoError(t, err)
	})
}
