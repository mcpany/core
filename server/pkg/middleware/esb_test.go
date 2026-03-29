package middleware

import (
	"context"
	"testing"
	"time"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestESBMiddleware_Execute(t *testing.T) {
	// Dummy handler to simulate the "next" function in the middleware chain.
	// It returns a predefined successful result.
	expectedResult := &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: "success"},
		},
	}
	nextHandler := func(ctx context.Context, method string, req mcp.Request) (mcp.Result, error) {
		return expectedResult, nil
	}

	tests := []struct {
		name            string
		config          *configv1.Middleware
		req             mcp.Request
		setupCtx        func() context.Context
		expectedError   bool
		expectedMessage string
		expectNext      bool
	}{
		{
			name:   "Disabled middleware",
			config: &configv1.Middleware{Disabled: true},
			req:    &mcp.CallToolRequest{},
			setupCtx: func() context.Context {
				return context.Background()
			},
			expectedError: false,
			expectNext:    true,
		},
		{
			name:   "Non-CallTool request passes through",
			config: &configv1.Middleware{Disabled: false},
			req:    &mcp.InitializeRequest{}, // Not a CallToolRequest
			setupCtx: func() context.Context {
				return context.Background()
			},
			expectedError: false,
			expectNext:    true,
		},
		{
			name:   "CallTool request with valid strongly typed context keys",
			config: &configv1.Middleware{Disabled: false},
			req:    &mcp.CallToolRequest{},
			setupCtx: func() context.Context {
				ctx := context.WithValue(context.Background(), missionIntentKey, "intent123")
				ctx = context.WithValue(ctx, entanglementShardKey, "shard456")
				return ctx
			},
			expectedError: false,
			expectNext:    true,
		},
		{
			name:   "CallTool request with valid fallback string context keys",
			config: &configv1.Middleware{Disabled: false},
			req:    &mcp.CallToolRequest{},
			setupCtx: func() context.Context {
				ctx := context.WithValue(context.Background(), "x-mission-intent", "intent123")
				ctx = context.WithValue(ctx, "x-entanglement-shard", "shard456")
				return ctx
			},
			expectedError: false,
			expectNext:    true,
		},
		{
			name:   "CallTool request missing intent (strongly typed empty string)",
			config: &configv1.Middleware{Disabled: false},
			req:    &mcp.CallToolRequest{},
			setupCtx: func() context.Context {
				ctx := context.WithValue(context.Background(), missionIntentKey, "")
				ctx = context.WithValue(ctx, entanglementShardKey, "shard456")
				return ctx
			},
			expectedError:   true,
			expectedMessage: "ESB Error: Missing x-mission-intent header/context.",
			expectNext:      false,
		},
		{
			name:   "CallTool request missing intent (completely missing)",
			config: &configv1.Middleware{Disabled: false},
			req:    &mcp.CallToolRequest{},
			setupCtx: func() context.Context {
				ctx := context.WithValue(context.Background(), entanglementShardKey, "shard456")
				return ctx
			},
			expectedError:   true,
			expectedMessage: "ESB Error: Missing x-mission-intent header/context.",
			expectNext:      false,
		},
		{
			name:   "CallTool request missing shard (strongly typed empty string)",
			config: &configv1.Middleware{Disabled: false},
			req:    &mcp.CallToolRequest{},
			setupCtx: func() context.Context {
				ctx := context.WithValue(context.Background(), missionIntentKey, "intent123")
				ctx = context.WithValue(ctx, entanglementShardKey, "")
				return ctx
			},
			expectedError:   true,
			expectedMessage: "ESB Error: Missing x-entanglement-shard header/context.",
			expectNext:      false,
		},
		{
			name:   "CallTool request missing shard (completely missing)",
			config: &configv1.Middleware{Disabled: false},
			req:    &mcp.CallToolRequest{},
			setupCtx: func() context.Context {
				ctx := context.WithValue(context.Background(), missionIntentKey, "intent123")
				return ctx
			},
			expectedError:   true,
			expectedMessage: "ESB Error: Missing x-entanglement-shard header/context.",
			expectNext:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewESBMiddleware(tt.config)
			ctx := tt.setupCtx()

			start := time.Now()
			result, err := m.Execute(ctx, "method", tt.req, nextHandler)
			duration := time.Since(start)

			require.NoError(t, err)

			if tt.expectNext {
				assert.Equal(t, expectedResult, result)
				// If middleware was enabled and it was a CallToolRequest, we expect Jitter
				if !m.enabled {
					return
				}
				if _, ok := tt.req.(*mcp.CallToolRequest); ok {
					assert.True(t, duration >= 4*time.Millisecond, "Expected jitter to take at least ~5ms, took %v", duration)
				}
			} else {
				// We expect it to short-circuit and return an error result
				callResult, ok := result.(*mcp.CallToolResult)
				require.True(t, ok, "Expected result to be *mcp.CallToolResult")
				assert.True(t, callResult.IsError)
				require.Len(t, callResult.Content, 1)

				textContent, ok := callResult.Content[0].(*mcp.TextContent)
				require.True(t, ok, "Expected content to be *mcp.TextContent")
				assert.Equal(t, tt.expectedMessage, textContent.Text)
			}
		})
	}
}

func TestESBMiddleware_New(t *testing.T) {
	t.Run("nil config", func(t *testing.T) {
		m := NewESBMiddleware(nil)
		assert.True(t, m.enabled, "Expected middleware to be enabled by default with nil config")
	})

	t.Run("config with disabled true", func(t *testing.T) {
		m := NewESBMiddleware(&configv1.Middleware{Disabled: true})
		assert.False(t, m.enabled, "Expected middleware to be disabled")
	})

	t.Run("config with disabled false", func(t *testing.T) {
		m := NewESBMiddleware(&configv1.Middleware{Disabled: false})
		assert.True(t, m.enabled, "Expected middleware to be enabled")
	})
}
