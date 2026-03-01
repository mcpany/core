// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware_test

import (
	"context"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/mcpany/core/server/pkg/middleware"
	"github.com/mcpany/core/server/pkg/tool"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
)

type hitlMockToolManager struct {
	tool.ManagerInterface
	mockGetTool        func(toolName string) (tool.Tool, bool)
	mockGetServiceInfo func(serviceID string) (*tool.ServiceInfo, bool)
}

func (m *hitlMockToolManager) GetTool(toolName string) (tool.Tool, bool) {
	if m.mockGetTool != nil {
		return m.mockGetTool(toolName)
	}
	return nil, false
}

func (m *hitlMockToolManager) GetServiceInfo(serviceID string) (*tool.ServiceInfo, bool) {
	if m.mockGetServiceInfo != nil {
		return m.mockGetServiceInfo(serviceID)
	}
	return nil, false
}

type hitlMockTool struct {
	serviceID string
}

func (m *hitlMockTool) Tool() *v1.Tool {
	t := &v1.Tool{}
	t.SetServiceId(m.serviceID)
	return t
}

func (m *hitlMockTool) GetCacheConfig() *configv1.CacheConfig {
	return nil
}

func (m *hitlMockTool) MCPTool() *mcp.Tool {
	return nil
}

func (m *hitlMockTool) Execute(ctx context.Context, req *tool.ExecutionRequest) (any, error) {
	return nil, nil
}

func TestHITLMiddleware_Execute(t *testing.T) {
	tests := []struct {
		name          string
		toolName      string
		serviceID     string
		mockToolFound bool
		mockSvcFound  bool
		hitlApproved  bool
		expectErr     bool
		errMsg        string
	}{
		{
			name:          "Tool not found - pass through",
			toolName:      "unknown_tool",
			mockToolFound: false,
			expectErr:     false,
		},
		{
			name:          "Service not found - error",
			toolName:      "my_tool",
			serviceID:     "my_svc",
			mockToolFound: true,
			mockSvcFound:  false,
			expectErr:     true,
			errMsg:        "service info not found for service my_svc",
		},
		{
			name:          "Normal tool - pass through",
			toolName:      "my_tool",
			serviceID:     "my_svc",
			mockToolFound: true,
			mockSvcFound:  true,
			expectErr:     false,
		},
		{
			name:          "Destructive tool - suspends",
			toolName:      "destructive_action",
			serviceID:     "my_svc",
			mockToolFound: true,
			mockSvcFound:  true,
			expectErr:     true,
			errMsg:        "execution suspended for HITL approval",
		},
		{
			name:          "Destructive tool but already approved - pass through",
			toolName:      "destructive_action",
			serviceID:     "my_svc",
			mockToolFound: true,
			mockSvcFound:  true,
			hitlApproved:  true,
			expectErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tm := &hitlMockToolManager{
				mockGetTool: func(toolName string) (tool.Tool, bool) {
					if tt.mockToolFound {
						return tool.Tool(&hitlMockTool{serviceID: tt.serviceID}), true
					}
					return nil, false
				},
				mockGetServiceInfo: func(serviceID string) (*tool.ServiceInfo, bool) {
					if tt.mockSvcFound {
						return &tool.ServiceInfo{}, true
					}
					return nil, false
				},
			}

			mw := middleware.NewHITLMiddleware(tm)

			ctx := context.Background()
			if tt.hitlApproved {
				ctx = context.WithValue(ctx, "hitl_approved", true)
			}

			req := &tool.ExecutionRequest{
				ToolName: tt.toolName,
			}

			nextCalled := false
			next := func(ctx context.Context, req *tool.ExecutionRequest) (any, error) {
				nextCalled = true
				return "success", nil
			}

			res, err := mw.Execute(ctx, req, next)

			if tt.expectErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
				assert.False(t, nextCalled)
				assert.Nil(t, res)
			} else {
				assert.NoError(t, err)
				assert.True(t, nextCalled)
				assert.Equal(t, "success", res)
			}
		})
	}
}
