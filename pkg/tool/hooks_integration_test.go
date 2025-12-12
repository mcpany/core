// Copyright 2025 Author(s) of MCP Any
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"context"
	"testing"

	"github.com/mcpany/core/pkg/bus"
	busproto "github.com/mcpany/core/proto/bus"
	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ptrIntegration is a helper to get a pointer to a value.
func ptrIntegration[T any](v T) *T {
	return &v
}

func TestToolManager_ExecuteTool_WithHooks(t *testing.T) {
	// Setup ToolManager
	busProvider, err := bus.NewBusProvider(&busproto.MessageBus{})
	require.NoError(t, err)
	tm := NewToolManager(busProvider)

	// Define Tool
	toolName := "my-tool"
	serviceID := "service-1"

	// Create local copies for pointers
	tn := toolName
	sid := serviceID
	protoTool := &v1.Tool{
		Name:      &tn,
		ServiceId: &sid,
	}

	// ToolID is conventionally serviceID.toolName (sanitized)
	toolID := serviceID + "." + toolName

	// 1. Test PreCallHook (Policy) Deny
	t.Run("PreCallHook_Deny", func(t *testing.T) {
		mockTool := &MockTool{
			ToolFunc: func() *v1.Tool { return protoTool },
			// ExecuteFunc should not be called if hook denies
		}

		tm.AddTool(mockTool)
		tm.AddServiceInfo(serviceID, &ServiceInfo{
			Config: &configv1.UpstreamServiceConfig{
				PreCallHooks: []*configv1.CallHook{
					{
						HookConfig: &configv1.CallHook_CallPolicy{
							CallPolicy: &configv1.CallPolicy{
								DefaultAction: ptrIntegration(configv1.CallPolicy_DENY),
							},
						},
					},
				},
			},
		})

		req := &ExecutionRequest{ToolName: toolID}
		_, err := tm.ExecuteTool(context.Background(), req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "denied")
	})

	// 2. Test PostCallHook (TextTruncation)
	t.Run("PostCallHook_Truncate", func(t *testing.T) {
		mockTool := &MockTool{
			ToolFunc: func() *v1.Tool { return protoTool },
			ExecuteFunc: func(ctx context.Context, req *ExecutionRequest) (any, error) {
				return "very long response", nil
			},
		}

		tm.AddTool(mockTool)
		tm.AddServiceInfo(serviceID, &ServiceInfo{
			Config: &configv1.UpstreamServiceConfig{
				PostCallHooks: []*configv1.CallHook{
					{
						HookConfig: &configv1.CallHook_TextTruncation{
							TextTruncation: &configv1.TextTruncationConfig{
								MaxChars: ptrIntegration(int32(5)),
							},
						},
					},
				},
			},
		})

		req := &ExecutionRequest{ToolName: toolID}
		res, err := tm.ExecuteTool(context.Background(), req)
		require.NoError(t, err)
		assert.Equal(t, "very ...", res)
	})

	// 3. Test CallPolicy (Legacy)
	t.Run("LegacyCallPolicy_Deny", func(t *testing.T) {
		mockTool := &MockTool{
			ToolFunc: func() *v1.Tool { return protoTool },
		}

		tm.AddTool(mockTool)
		tm.AddServiceInfo(serviceID, &ServiceInfo{
			Config: &configv1.UpstreamServiceConfig{
				CallPolicy: &configv1.CallPolicy{
					DefaultAction: ptrIntegration(configv1.CallPolicy_DENY),
				},
			},
		})

		req := &ExecutionRequest{ToolName: toolID}
		_, err := tm.ExecuteTool(context.Background(), req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "denied")
	})

	// 4. Test Tool Management (Get, List, Clear)
	t.Run("ToolManagement", func(t *testing.T) {
		// Add duplicate check? AddTool overwrites?
		mockTool := &MockTool{
			ToolFunc: func() *v1.Tool { return protoTool },
		}
		tm.AddTool(mockTool)

		tool, ok := tm.GetTool(toolID)
		assert.True(t, ok)
		assert.Equal(t, mockTool, tool)

		tools := tm.ListTools()
		assert.Len(t, tools, 1)
		assert.Equal(t, mockTool, tools[0])

		tm.ClearToolsForService(serviceID)
		tool, ok = tm.GetTool(toolID)
		assert.False(t, ok)
		assert.Nil(t, tool)

		tools = tm.ListTools()
		assert.Empty(t, tools)
	})

	// 5. Test Middleware
	t.Run("Middleware", func(t *testing.T) {
		middlewareCalled := false
		mw := &MockMiddleware{
			ExecuteFunc: func(ctx context.Context, req *ExecutionRequest, next ToolExecutionFunc) (any, error) {
				middlewareCalled = true
				return next(ctx, req)
			},
		}

		tmWithMw := NewToolManager(busProvider)
		tmWithMw.AddMiddleware(mw)

		mockToolMw := &MockTool{
			ToolFunc: func() *v1.Tool { return protoTool },
			ExecuteFunc: func(ctx context.Context, req *ExecutionRequest) (any, error) {
				return "ok", nil
			},
		}
		tmWithMw.AddTool(mockToolMw)

		req := &ExecutionRequest{ToolName: toolID}
		res, err := tmWithMw.ExecuteTool(context.Background(), req)
		require.NoError(t, err)
		assert.Equal(t, "ok", res)
		assert.True(t, middlewareCalled)
	})
}

type MockMiddleware struct {
	ExecuteFunc func(ctx context.Context, req *ExecutionRequest, next ToolExecutionFunc) (any, error)
}

func (m *MockMiddleware) Execute(
	ctx context.Context,
	req *ExecutionRequest,
	next ToolExecutionFunc,
) (any, error) {
	if m.ExecuteFunc != nil {
		return m.ExecuteFunc(ctx, req, next)
	}
	return next(ctx, req)
}
