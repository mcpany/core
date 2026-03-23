// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

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

// TestESBMiddleware_Execute validates the ESBMiddleware logic, specifically its ability to enforce intent checking and jitter.
//
// Summary: Tests the main ESB middleware execution.
//
// Parameters:
//   - t: *testing.T. The testing context.
//
// Side Effects:
//   - Injects temporal jitter which delays execution.
func TestESBMiddleware_Execute(t *testing.T) {
	t.Parallel()

	disabled := false
	enabledConfig := &configv1.Middleware{Disabled: &disabled}
	disabledConfig := &configv1.Middleware{Disabled: func() *bool { b := true; return &b }()}

	tests := []struct {
		name              string
		config            *configv1.Middleware
		req               mcp.Request
		setupContext      func() context.Context
		expectedError     bool
		expectedResultErr bool
		expectedErrorText string
		expectNextCalled  bool
		minJitter         time.Duration
	}{
		{
			name:   "Middleware Disabled",
			config: disabledConfig,
			req:    &mcp.CallToolRequest{},
			setupContext: func() context.Context {
				return context.Background()
			},
			expectedError:     false,
			expectedResultErr: false,
			expectNextCalled:  true,
		},
		{
			name:   "Not a Tool Call",
			config: enabledConfig,
			req:    &mcp.ListToolsRequest{},
			setupContext: func() context.Context {
				return context.Background()
			},
			expectedError:     false,
			expectedResultErr: false,
			expectNextCalled:  true,
		},
		{
			name:   "Missing Both Headers",
			config: enabledConfig,
			req:    &mcp.CallToolRequest{},
			setupContext: func() context.Context {
				return context.Background()
			},
			expectedError:     false,
			expectedResultErr: true,
			expectedErrorText: "Missing x-mission-intent",
			expectNextCalled:  false,
		},
		{
			name:   "Missing Shard Header",
			config: enabledConfig,
			req:    &mcp.CallToolRequest{},
			setupContext: func() context.Context {
				ctx := context.Background()
				return context.WithValue(ctx, missionIntentKey, "intent123")
			},
			expectedError:     false,
			expectedResultErr: true,
			expectedErrorText: "Missing x-entanglement-shard",
			expectNextCalled:  false,
		},
		{
			name:   "Fallback String Context for Intent",
			config: enabledConfig,
			req:    &mcp.CallToolRequest{},
			setupContext: func() context.Context {
				ctx := context.Background()
				return context.WithValue(ctx, "x-mission-intent", "intent123")
			},
			expectedError:     false,
			expectedResultErr: true, // Still missing shard
			expectedErrorText: "Missing x-entanglement-shard",
			expectNextCalled:  false,
		},
		{
			name:   "Fallback String Context for Shard",
			config: enabledConfig,
			req:    &mcp.CallToolRequest{},
			setupContext: func() context.Context {
				ctx := context.Background()
				ctx = context.WithValue(ctx, missionIntentKey, "intent123")
				return context.WithValue(ctx, "x-entanglement-shard", "shard456")
			},
			expectedError:     false,
			expectedResultErr: false,
			expectNextCalled:  true,
			minJitter:         5 * time.Millisecond,
		},
		{
			name:   "Success Happy Path",
			config: enabledConfig,
			req:    &mcp.CallToolRequest{},
			setupContext: func() context.Context {
				ctx := context.Background()
				ctx = context.WithValue(ctx, missionIntentKey, "intent123")
				ctx = context.WithValue(ctx, entanglementShardKey, "shard456")
				return ctx
			},
			expectedError:     false,
			expectedResultErr: false,
			expectNextCalled:  true,
			minJitter:         5 * time.Millisecond,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			middleware := NewESBMiddleware(tt.config)
			ctx := tt.setupContext()

			nextCalled := false
			next := func(ctx context.Context, method string, request mcp.Request) (mcp.Result, error) {
				nextCalled = true
				return nil, nil
			}

			start := time.Now()
			res, err := middleware.Execute(ctx, "tools/call", tt.req, next)
			elapsed := time.Since(start)

			if tt.expectedError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			if tt.expectedResultErr {
				require.NotNil(t, res)
				callRes, ok := res.(*mcp.CallToolResult)
				require.True(t, ok)
				assert.True(t, callRes.IsError)
				if tt.expectedErrorText != "" {
					assert.Contains(t, callRes.Content[0].(*mcp.TextContent).Text, tt.expectedErrorText)
				}
			}

			assert.Equal(t, tt.expectNextCalled, nextCalled)

			if tt.minJitter > 0 {
				assert.True(t, elapsed >= tt.minJitter, "Expected minimum jitter of %v, got %v", tt.minJitter, elapsed)
			}
		})
	}
}
