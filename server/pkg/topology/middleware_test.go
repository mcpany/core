// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package topology

import (
	"context"
	"errors"
	"testing"
	"time"

	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/mcpany/core/server/pkg/auth"
	"github.com/mcpany/core/server/pkg/consts"
	"github.com/mcpany/core/server/pkg/tool"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"google.golang.org/protobuf/proto"
)

// mockTool implements tool.Tool interface for testing
type mockTool struct {
	tool *v1.Tool
}

func (m *mockTool) Tool() *v1.Tool {
	return m.tool
}

func (m *mockTool) MCPTool() *mcp.Tool {
	return nil
}

func (m *mockTool) Execute(ctx context.Context, req *tool.ExecutionRequest) (any, error) {
	return nil, nil
}

func (m *mockTool) GetCacheConfig() *configv1.CacheConfig {
	return nil
}

func TestMiddleware(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tests := []struct {
		name          string
		method        string
		req           mcp.Request
		setupContext  func(context.Context) context.Context
		setupMocks    func(*tool.MockManagerInterface)
		nextHandler   mcp.MethodHandler
		expectStats   func(*testing.T, Stats)
		expectSession func(*testing.T, *Manager)
	}{
		{
			name:   "basic_activity",
			method: "tools/list",
			req:    &mcp.ListToolsRequest{},
			setupContext: func(ctx context.Context) context.Context {
				return ctx
			},
			setupMocks: func(tm *tool.MockManagerInterface) {},
			nextHandler: func(ctx context.Context, method string, req mcp.Request) (mcp.Result, error) {
				return &mcp.ListToolsResult{}, nil
			},
			expectStats: func(t *testing.T, s Stats) {
				assert.Equal(t, int64(1), s.TotalRequests)
				assert.Equal(t, float64(0), s.ErrorRate)
			},
			expectSession: func(t *testing.T, m *Manager) {
				// Should fall back to unknown
				m.mu.RLock()
				defer m.mu.RUnlock()
				_, ok := m.sessions["unknown"]
				assert.True(t, ok)
			},
		},
		{
			name:   "auth_context",
			method: "tools/list",
			req:    &mcp.ListToolsRequest{},
			setupContext: func(ctx context.Context) context.Context {
				return auth.ContextWithUser(ctx, "test-user")
			},
			setupMocks: func(tm *tool.MockManagerInterface) {},
			nextHandler: func(ctx context.Context, method string, req mcp.Request) (mcp.Result, error) {
				return &mcp.ListToolsResult{}, nil
			},
			expectStats: func(t *testing.T, s Stats) {
				assert.Equal(t, int64(1), s.TotalRequests)
			},
			expectSession: func(t *testing.T, m *Manager) {
				m.mu.RLock()
				defer m.mu.RUnlock()
				_, ok := m.sessions["user-test-user"]
				assert.True(t, ok)
			},
		},
		{
			name:   "remote_addr",
			method: "tools/list",
			req:    &mcp.ListToolsRequest{},
			setupContext: func(ctx context.Context) context.Context {
				return context.WithValue(ctx, consts.ContextKeyRemoteAddr, "127.0.0.1")
			},
			setupMocks: func(tm *tool.MockManagerInterface) {},
			nextHandler: func(ctx context.Context, method string, req mcp.Request) (mcp.Result, error) {
				return &mcp.ListToolsResult{}, nil
			},
			expectStats: func(t *testing.T, s Stats) {
				assert.Equal(t, int64(1), s.TotalRequests)
			},
			expectSession: func(t *testing.T, m *Manager) {
				m.mu.RLock()
				defer m.mu.RUnlock()
				_, ok := m.sessions["ip-127.0.0.1"]
				assert.True(t, ok)
			},
		},
		{
			name:   "tool_call",
			method: "tools/call",
			req: &mcp.CallToolRequest{
				Params: &mcp.CallToolParamsRaw{
					Name: "test-service.tool",
				},
			},
			setupContext: func(ctx context.Context) context.Context {
				return ctx
			},
			setupMocks: func(tm *tool.MockManagerInterface) {
				tm.EXPECT().GetTool("test-service.tool").Return(&mockTool{
					tool: v1.Tool_builder{
						Name:      proto.String("test-service.tool"),
						ServiceId: proto.String("test-service"),
					}.Build(),
				}, true)
			},
			nextHandler: func(ctx context.Context, method string, req mcp.Request) (mcp.Result, error) {
				return &mcp.CallToolResult{}, nil
			},
			expectStats: func(t *testing.T, s Stats) {
				assert.Equal(t, int64(1), s.TotalRequests)
			},
			expectSession: func(t *testing.T, m *Manager) {
				m.mu.RLock()
				defer m.mu.RUnlock()
				session, ok := m.sessions["unknown"]
				require.True(t, ok)
				assert.Equal(t, int64(1), session.ServiceCounts["test-service"])
			},
		},
		{
			name:   "error_handling",
			method: "tools/list",
			req:    &mcp.ListToolsRequest{},
			setupContext: func(ctx context.Context) context.Context {
				return ctx
			},
			setupMocks: func(tm *tool.MockManagerInterface) {},
			nextHandler: func(ctx context.Context, method string, req mcp.Request) (mcp.Result, error) {
				return nil, errors.New("handler error")
			},
			expectStats: func(t *testing.T, s Stats) {
				assert.Equal(t, int64(1), s.TotalRequests)
				assert.Equal(t, float64(1.0), s.ErrorRate)
			},
			expectSession: func(t *testing.T, m *Manager) {
				m.mu.RLock()
				defer m.mu.RUnlock()
				session, ok := m.sessions["unknown"]
				require.True(t, ok)
				assert.Equal(t, int64(1), session.ErrorCount)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockTM := tool.NewMockManagerInterface(ctrl)
			m := NewManager(nil, mockTM)
			defer m.Close()

			ctx := tt.setupContext(context.Background())
			tt.setupMocks(mockTM)

			middleware := m.Middleware(tt.nextHandler)
			_, _ = middleware(ctx, tt.method, tt.req)

			assert.Eventually(t, func() bool {
				stats := m.GetStats("")
				return stats.TotalRequests > 0
			}, 1*time.Second, 10*time.Millisecond, "stats should be updated")

			tt.expectStats(t, m.GetStats(""))
			if tt.expectSession != nil {
				tt.expectSession(t, m)
			}
		})
	}
}
