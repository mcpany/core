// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"context"
	"errors"
	"testing"

	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

// MockHook implements PreCallHook and PostCallHook
type MockHook struct {
	Action Action
	Err    error
}

func (h *MockHook) ExecutePre(ctx context.Context, req *ExecutionRequest) (Action, *ExecutionRequest, error) {
	return h.Action, nil, h.Err
}

func (h *MockHook) ExecutePost(ctx context.Context, req *ExecutionRequest, result any) (any, error) {
	return result, h.Err
}

// MockMiddlewareExtra
type MockMiddlewareExtra struct {
	Called bool
}

func (m *MockMiddlewareExtra) Execute(ctx context.Context, req *ExecutionRequest, next ExecutionFunc) (any, error) {
	m.Called = true
	return next(ctx, req)
}

func TestManager_ExecuteTool_Hooks(t *testing.T) {
	manager := NewManager(nil)

	// Add a service with a hook
	serviceID := "s1"
	info := &ServiceInfo{
		Name:         "s1",
		HealthStatus: "healthy",
		PreHooks:     []PreCallHook{},
	}
	manager.AddServiceInfo(serviceID, info)

	// Add tool
	mockTool := &MockTool{
		ToolFunc: func() *v1.Tool {
			return &v1.Tool{
				Name:      proto.String("tool1"),
				ServiceId: proto.String("s1"),
			}
		},
		ExecuteFunc: func(ctx context.Context, req *ExecutionRequest) (any, error) {
			return "ok", nil
		},
	}
	manager.AddTool(mockTool)

	req := &ExecutionRequest{ToolName: "s1.tool1"}

	// Test Deny Hook
	info.PreHooks = []PreCallHook{&MockHook{Action: ActionDeny}}
	manager.AddServiceInfo(serviceID, info) // Re-add to update

	_, err := manager.ExecuteTool(context.Background(), req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "denied by hook")

	// Test Error Hook
	info.PreHooks = []PreCallHook{&MockHook{Err: errors.New("hook failed")}}
	manager.AddServiceInfo(serviceID, info)

	_, err = manager.ExecuteTool(context.Background(), req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "hook failed")

	// Test Post Hook Error
	info.PreHooks = nil
	info.PostHooks = []PostCallHook{&MockHook{Err: errors.New("post fail")}}
	manager.AddServiceInfo(serviceID, info)

	_, err = manager.ExecuteTool(context.Background(), req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "post fail")
}

func TestManager_ExecuteTool_Unhealthy(t *testing.T) {
	manager := NewManager(nil)
	serviceID := "s1"
	info := &ServiceInfo{
		Name:         "s1",
		HealthStatus: "unhealthy",
	}
	manager.AddServiceInfo(serviceID, info)

	mockTool := &MockTool{
		ToolFunc: func() *v1.Tool {
			return &v1.Tool{
				Name:      proto.String("tool1"),
				ServiceId: proto.String("s1"),
			}
		},
	}
	manager.AddTool(mockTool)

	req := &ExecutionRequest{ToolName: "s1.tool1"}
	_, err := manager.ExecuteTool(context.Background(), req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "service s1 is currently unhealthy")
}

func TestManager_ExecuteTool_Middleware(t *testing.T) {
	manager := NewManager(nil)

	mw := &MockMiddlewareExtra{}
	manager.AddMiddleware(mw)

	// Add tool
	mockTool := &MockTool{
		ToolFunc: func() *v1.Tool {
			return &v1.Tool{
				Name:      proto.String("tool1"),
				ServiceId: proto.String("s1"),
			}
		},
		ExecuteFunc: func(ctx context.Context, req *ExecutionRequest) (any, error) {
			return "ok", nil
		},
	}
	manager.AddTool(mockTool)
	// Need service info?
	manager.AddServiceInfo("s1", &ServiceInfo{HealthStatus: "healthy"})

	req := &ExecutionRequest{ToolName: "s1.tool1"}
	_, err := manager.ExecuteTool(context.Background(), req)
	assert.NoError(t, err)
	assert.True(t, mw.Called)
}
