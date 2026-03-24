package middleware

import (
	"context"
	"testing"
	"time"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func TestNewESBMiddleware(t *testing.T) {
	// Test nil config
	m1 := NewESBMiddleware(nil)
	assert.True(t, m1.enabled)

	// Test config with enabled
	m2 := NewESBMiddleware(configv1.Middleware_builder{Disabled: proto.Bool(false)}.Build())
	assert.True(t, m2.enabled)

	// Test config with disabled
	m3 := NewESBMiddleware(configv1.Middleware_builder{Disabled: proto.Bool(true)}.Build())
	assert.False(t, m3.enabled)
}

func TestESBMiddleware_Execute(t *testing.T) {
	tests := []struct {
		name          string
		enabled       bool
		req           mcp.Request
		setupCtx      func() context.Context
		expectedError bool
		expectedMsg   string
	}{
		{
			name:    "Disabled middleware passes through",
			enabled: false,
			req:     &mcp.CallToolRequest{},
			setupCtx: func() context.Context {
				return context.Background()
			},
			expectedError: false,
		},
		{
			name:    "Non-tool request passes through",
			enabled: true,
			req:     &mcp.ListToolsRequest{},
			setupCtx: func() context.Context {
				return context.Background()
			},
			expectedError: false,
		},
		{
			name:    "Missing mission intent",
			enabled: true,
			req:     &mcp.CallToolRequest{},
			setupCtx: func() context.Context {
				return context.Background()
			},
			expectedError: true,
			expectedMsg:   "Missing x-mission-intent",
		},
		{
			name:    "Missing entanglement shard",
			enabled: true,
			req:     &mcp.CallToolRequest{},
			setupCtx: func() context.Context {
				ctx := context.WithValue(context.Background(), missionIntentKey, "intent-123")
				return ctx
			},
			expectedError: true,
			expectedMsg:   "Missing x-entanglement-shard",
		},
		{
			name:    "Valid request with typed keys",
			enabled: true,
			req:     &mcp.CallToolRequest{},
			setupCtx: func() context.Context {
				ctx := context.WithValue(context.Background(), missionIntentKey, "intent-123")
				ctx = context.WithValue(ctx, entanglementShardKey, "shard-456")
				return ctx
			},
			expectedError: false,
		},
		{
			name:    "Valid request with string fallback keys",
			enabled: true,
			req:     &mcp.CallToolRequest{},
			setupCtx: func() context.Context {
				ctx := context.WithValue(context.Background(), "x-mission-intent", "intent-123")
				ctx = context.WithValue(ctx, "x-entanglement-shard", "shard-456")
				return ctx
			},
			expectedError: false,
		},
		{
			name:    "Empty string values should fail (intent)",
			enabled: true,
			req:     &mcp.CallToolRequest{},
			setupCtx: func() context.Context {
				ctx := context.WithValue(context.Background(), missionIntentKey, "")
				return ctx
			},
			expectedError: true,
			expectedMsg:   "Missing x-mission-intent",
		},
		{
			name:    "Empty string values should fail (shard)",
			enabled: true,
			req:     &mcp.CallToolRequest{},
			setupCtx: func() context.Context {
				ctx := context.WithValue(context.Background(), missionIntentKey, "valid")
				ctx = context.WithValue(ctx, entanglementShardKey, "")
				return ctx
			},
			expectedError: true,
			expectedMsg:   "Missing x-entanglement-shard",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &ESBMiddleware{enabled: tt.enabled}

			// Mock next handler
			nextCalled := false
			next := func(ctx context.Context, method string, req mcp.Request) (mcp.Result, error) {
				nextCalled = true
				return &mcp.CallToolResult{
					Content: []mcp.Content{
						&mcp.TextContent{Text: "success"},
					},
				}, nil
			}

			ctx := tt.setupCtx()

			start := time.Now()
			res, err := m.Execute(ctx, "tools/call", tt.req, next)
			duration := time.Since(start)

			require.NoError(t, err) // Execute itself doesn't return err, it returns CallToolResult with IsError=true

			if tt.expectedError {
				assert.False(t, nextCalled, "Next handler should not be called on error")
				callRes, ok := res.(*mcp.CallToolResult)
				require.True(t, ok)
				assert.True(t, callRes.IsError)
				assert.Len(t, callRes.Content, 1)
				assert.Contains(t, callRes.Content[0].(*mcp.TextContent).Text, tt.expectedMsg)
			} else {
				assert.True(t, nextCalled, "Next handler should be called on success")

				// Verify TSJ timing if enabled and is CallToolRequest
				if tt.enabled {
					if _, ok := tt.req.(*mcp.CallToolRequest); ok {
						assert.GreaterOrEqual(t, duration, 5*time.Millisecond, "TSJ jitter should be at least 5ms")
					}
				}
			}
		})
	}
}

func TestESBMiddleware_InjectTSJ(t *testing.T) {
	m := &ESBMiddleware{enabled: true}

	// Multiple runs to ensure randomness and boundaries
	for i := 0; i < 5; i++ {
		start := time.Now()
		m.injectTSJ()
		duration := time.Since(start)

		// It should take at least 5ms, and practically not more than 100ms (50ms jitter + overhead)
		assert.GreaterOrEqual(t, duration, 5*time.Millisecond)
		assert.LessOrEqual(t, duration, 200*time.Millisecond)
	}
}
