// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package topology

import (
	"context"
	"errors"
	"testing"
	"time"

	mcp_router_v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/mcpany/core/server/pkg/auth"
	"github.com/mcpany/core/server/pkg/consts"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

func TestMiddleware_Comprehensive(t *testing.T) {
	// Define test cases
	tests := []struct {
		name           string
		method         string
		req            mcp.Request
		nextResult     mcp.Result
		nextError      error
		setupMock      func(mockTM *MockToolManager, mockRegistry *MockServiceRegistry)
		setupContext   func() context.Context
		expectedError  bool
		verifyActivity func(t *testing.T, session *SessionStats)
	}{
		{
			name:   "Standard Tool Call - Text Content",
			method: "tools/call",
			req: &mcp.CallToolRequest{
				Params: &mcp.CallToolParamsRaw{
					Name: "test-tool",
				},
			},
			nextResult: &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: "Hello World"}, // Length 11
				},
			},
			setupMock: func(mockTM *MockToolManager, mockRegistry *MockServiceRegistry) {
				mockTool := new(MockTool)
				mockTool.On("Tool").Return(mcp_router_v1.Tool_builder{
					Name:      proto.String("test-tool"),
					ServiceId: proto.String("svc-1"),
				}.Build())
				// Allow GetTool to be called
				mockTM.On("GetTool", "test-tool").Return(mockTool, true).Once()
			},
			setupContext: func() context.Context {
				return auth.ContextWithUser(context.Background(), "user-1")
			},
			verifyActivity: func(t *testing.T, s *SessionStats) {
				assert.Equal(t, "user-user-1", s.ID)
				assert.Equal(t, int64(11), s.TotalBytes)
				assert.Equal(t, int64(1), s.RequestCount)
				assert.Equal(t, int64(0), s.ErrorCount)
				// Check Service Stats
				assert.Equal(t, int64(1), s.ServiceCounts["svc-1"])
			},
		},
		{
			name:   "Standard Tool Call - Image Content",
			method: "tools/call",
			req: &mcp.CallToolRequest{
				Params: &mcp.CallToolParamsRaw{
					Name: "image-tool",
				},
			},
			nextResult: &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.ImageContent{Data: []byte("base64data")}, // Length 10
				},
			},
			setupMock: func(mockTM *MockToolManager, mockRegistry *MockServiceRegistry) {
				mockTool := new(MockTool)
				mockTool.On("Tool").Return(mcp_router_v1.Tool_builder{
					Name:      proto.String("image-tool"),
					ServiceId: proto.String("svc-img"),
				}.Build())
				mockTM.On("GetTool", "image-tool").Return(mockTool, true).Once()
			},
			setupContext: func() context.Context {
				return context.WithValue(context.Background(), consts.ContextKeyRemoteAddr, "192.168.1.1")
			},
			verifyActivity: func(t *testing.T, s *SessionStats) {
				assert.Equal(t, "ip-192.168.1.1", s.ID)
				assert.Equal(t, int64(10), s.TotalBytes)
				assert.Equal(t, int64(1), s.ServiceCounts["svc-img"])
			},
		},
		{
			name:   "Standard Tool Call - Embedded Resource",
			method: "tools/call",
			req: &mcp.CallToolRequest{
				Params: &mcp.CallToolParamsRaw{
					Name: "embed-tool",
				},
			},
			nextResult: &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.EmbeddedResource{
						Resource: &mcp.ResourceContents{
							Text: "some text",      // 9
							Blob: []byte("blob"), // 4
						},
					},
				},
			},
			setupMock: func(mockTM *MockToolManager, mockRegistry *MockServiceRegistry) {
				mockTool := new(MockTool)
				mockTool.On("Tool").Return(mcp_router_v1.Tool_builder{
					Name:      proto.String("embed-tool"),
					ServiceId: proto.String("svc-embed"),
				}.Build())
				mockTM.On("GetTool", "embed-tool").Return(mockTool, true).Once()
			},
			setupContext: func() context.Context {
				return context.Background() // Unknown session
			},
			verifyActivity: func(t *testing.T, s *SessionStats) {
				assert.Equal(t, "unknown", s.ID)
				assert.Equal(t, int64(13), s.TotalBytes) // 9 + 4
			},
		},
		{
			name:   "Read Resource Result",
			method: "resources/read",
			req:    &mcp.ReadResourceRequest{},
			nextResult: &mcp.ReadResourceResult{
				Contents: []*mcp.ResourceContents{
					{Text: "resource text"},     // 13
					{Blob: []byte("data")},      // 4
				},
			},
			setupMock: func(mockTM *MockToolManager, mockRegistry *MockServiceRegistry) {}, // No tool lookup for resources/read in middleware logic
			setupContext: func() context.Context {
				return context.Background()
			},
			verifyActivity: func(t *testing.T, s *SessionStats) {
				assert.Equal(t, int64(17), s.TotalBytes) // 13 + 4
				// ServiceID is empty for non-tool calls in current logic
				assert.Empty(t, s.ServiceCounts)
			},
		},
		{
			name:   "Error from Next Handler",
			method: "tools/call",
			req: &mcp.CallToolRequest{
				Params: &mcp.CallToolParamsRaw{
					Name: "error-tool",
				},
			},
			nextError: errors.New("something went wrong"),
			expectedError: true,
			setupMock: func(mockTM *MockToolManager, mockRegistry *MockServiceRegistry) {
				mockTool := new(MockTool)
				mockTool.On("Tool").Return(mcp_router_v1.Tool_builder{
					Name:      proto.String("error-tool"),
					ServiceId: proto.String("svc-err"),
				}.Build())
				mockTM.On("GetTool", "error-tool").Return(mockTool, true).Once()
			},
			setupContext: func() context.Context {
				return context.Background()
			},
			verifyActivity: func(t *testing.T, s *SessionStats) {
				assert.Equal(t, int64(1), s.ErrorCount)
				assert.Equal(t, int64(1), s.ServiceErrors["svc-err"])
			},
		},
		{
			name:   "Unknown Tool",
			method: "tools/call",
			req: &mcp.CallToolRequest{
				Params: &mcp.CallToolParamsRaw{
					Name: "unknown-tool",
				},
			},
			nextResult: &mcp.CallToolResult{},
			setupMock: func(mockTM *MockToolManager, mockRegistry *MockServiceRegistry) {
				mockTM.On("GetTool", "unknown-tool").Return(nil, false).Once()
			},
			setupContext: func() context.Context {
				return context.Background()
			},
			verifyActivity: func(t *testing.T, s *SessionStats) {
				// ServiceID should be empty
				assert.Empty(t, s.ServiceCounts)
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Setup Mocks per test case
			mockRegistry := new(MockServiceRegistry)
			mockTM := new(MockToolManager)

			if tc.setupMock != nil {
				tc.setupMock(mockTM, mockRegistry)
			}

			// Create Manager per test case
			m := NewManager(mockRegistry, mockTM)
			defer m.Close()

			// Construct Middleware
			nextHandler := func(ctx context.Context, method string, req mcp.Request) (mcp.Result, error) {
				if tc.nextError != nil {
					return nil, tc.nextError
				}
				return tc.nextResult, nil
			}

			wrapped := m.Middleware(nextHandler)
			ctx := tc.setupContext()

			// Execute
			_, err := wrapped(ctx, tc.method, tc.req)

			if tc.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			// Wait for async activity recording
			sessionID := "unknown"
			if uid, ok := auth.UserFromContext(ctx); ok {
				sessionID = "user-" + uid
			} else if ip, ok := ctx.Value(consts.ContextKeyRemoteAddr).(string); ok {
				sessionID = "ip-" + ip
			}

			assert.Eventually(t, func() bool {
				m.mu.RLock()
				defer m.mu.RUnlock()
				_, exists := m.sessions[sessionID]
				return exists
			}, 1*time.Second, 10*time.Millisecond, "Session stats not recorded")

			m.mu.RLock()
			stats := m.sessions[sessionID]
			m.mu.RUnlock()

			if tc.verifyActivity != nil {
				tc.verifyActivity(t, stats)
			}
		})
	}
}
