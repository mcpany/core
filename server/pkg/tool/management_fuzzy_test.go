// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"context"
	"errors"
	"fmt"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	mcp_router_v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

// MockPreHook is a mock for PreCallHook
type MockPreHook struct {
	ExecutePreFunc func(ctx context.Context, req *ExecutionRequest) (Action, *ExecutionRequest, error)
}

func (m *MockPreHook) ExecutePre(ctx context.Context, req *ExecutionRequest) (Action, *ExecutionRequest, error) {
	if m.ExecutePreFunc != nil {
		return m.ExecutePreFunc(ctx, req)
	}
	return ActionAllow, nil, nil
}

// MockPostHook is a mock for PostCallHook
type MockPostHook struct {
	ExecutePostFunc func(ctx context.Context, req *ExecutionRequest, result any) (any, error)
}

func (m *MockPostHook) ExecutePost(ctx context.Context, req *ExecutionRequest, result any) (any, error) {
	if m.ExecutePostFunc != nil {
		return m.ExecutePostFunc(ctx, req, result)
	}
	return result, nil
}

func TestToolManager_FuzzyMatching_SingleMatch(t *testing.T) {
	t.Parallel()
	tm := NewManager(nil)

	// Register "get_weather"
	mockTool := &MockTool{
		ToolFunc: func() *mcp_router_v1.Tool {
			return mcp_router_v1.Tool_builder{
				ServiceId: proto.String("weather-service"),
				Name:      proto.String("get_weather"),
			}.Build()
		},
	}
	_ = tm.AddTool(mockTool)

	// Call "weather-service.get_weathr" (typo in full name)
	req := &ExecutionRequest{ToolName: "weather-service.get_weathr", ToolInputs: []byte(`{}`)}
	_, err := tm.ExecuteTool(context.Background(), req)

	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrToolNotFound)
	assert.Contains(t, err.Error(), `did you mean "weather-service.get_weather"?`)
}

func TestToolManager_FuzzyMatching_SuffixMatch(t *testing.T) {
	t.Parallel()
	tm := NewManager(nil)

	// Register "service.get_weather"
	mockTool := &MockTool{
		ToolFunc: func() *mcp_router_v1.Tool {
			return mcp_router_v1.Tool_builder{
				ServiceId: proto.String("service"),
				Name:      proto.String("get_weather"),
			}.Build()
		},
	}
	_ = tm.AddTool(mockTool)

	// Call "get_weather" (without namespace)
	// This should match as a suffix match if unique.
	req := &ExecutionRequest{ToolName: "get_weather", ToolInputs: []byte(`{}`)}
	_, err := tm.ExecuteTool(context.Background(), req)

	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrToolNotFound)
	assert.Contains(t, err.Error(), `did you mean "service.get_weather"?`)
}

func TestToolManager_FuzzyMatching_MultipleSuggestions(t *testing.T) {
	t.Parallel()
	tm := NewManager(nil)

	// Register "s1.tool" and "s2.tool"
	tool1 := &MockTool{
		ToolFunc: func() *mcp_router_v1.Tool {
			return mcp_router_v1.Tool_builder{ServiceId: proto.String("s1"), Name: proto.String("tool")}.Build()
		},
	}
	tool2 := &MockTool{
		ToolFunc: func() *mcp_router_v1.Tool {
			return mcp_router_v1.Tool_builder{ServiceId: proto.String("s2"), Name: proto.String("tool")}.Build()
		},
	}
	_ = tm.AddTool(tool1)
	_ = tm.AddTool(tool2)

	// Call "tool"
	req := &ExecutionRequest{ToolName: "tool", ToolInputs: []byte(`{}`)}
	_, err := tm.ExecuteTool(context.Background(), req)

	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrToolNotFound)
	assert.Contains(t, err.Error(), "did you mean one of:")
	assert.Contains(t, err.Error(), "s1.tool")
	assert.Contains(t, err.Error(), "s2.tool")
}

func TestToolManager_ServiceUnhealthy(t *testing.T) {
	t.Parallel()
	tm := NewManager(nil)
	serviceID := "unhealthy-service"

	// Add Service Info with Unhealthy Status
	info := &ServiceInfo{
		Name:         "Unhealthy Service",
		HealthStatus: HealthStatusUnhealthy,
		Config:       configv1.UpstreamServiceConfig_builder{Name: proto.String("unhealthy-service")}.Build(),
	}
	tm.AddServiceInfo(serviceID, info)

	// Register tool
	mockTool := &MockTool{
		ToolFunc: func() *mcp_router_v1.Tool {
			return mcp_router_v1.Tool_builder{
				ServiceId: proto.String(serviceID),
				Name:      proto.String("tool"),
			}.Build()
		},
	}
	_ = tm.AddTool(mockTool)

	// Execute Tool
	req := &ExecutionRequest{ToolName: serviceID + ".tool", ToolInputs: []byte(`{}`)}
	_, err := tm.ExecuteTool(context.Background(), req)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), fmt.Sprintf("service %s is currently unhealthy", serviceID))
}

func TestToolManager_PreHook_Failure(t *testing.T) {
	t.Parallel()
	tm := NewManager(nil)
	serviceID := "prehook-fail-service"

	expectedErr := errors.New("pre-hook failed")
	mockHook := &MockPreHook{
		ExecutePreFunc: func(ctx context.Context, req *ExecutionRequest) (Action, *ExecutionRequest, error) {
			return ActionAllow, nil, expectedErr
		},
	}

	// We need to inject this hook into the service info.
	// If info.Config is NIL, AddServiceInfo won't overwrite PreHooks, allowing manual injection.
	info := &ServiceInfo{
		Name:     "Hook Service",
		PreHooks: []PreCallHook{mockHook},
	}
	tm.AddServiceInfo(serviceID, info)

	// Register tool
	mockTool := &MockTool{
		ToolFunc: func() *mcp_router_v1.Tool {
			return mcp_router_v1.Tool_builder{
				ServiceId: proto.String(serviceID),
				Name:      proto.String("tool"),
			}.Build()
		},
	}
	_ = tm.AddTool(mockTool)

	req := &ExecutionRequest{ToolName: serviceID + ".tool", ToolInputs: []byte(`{}`)}
	_, err := tm.ExecuteTool(context.Background(), req)

	assert.Error(t, err)
	assert.Equal(t, expectedErr, err)
}

func TestToolManager_PreHook_Deny(t *testing.T) {
	t.Parallel()
	tm := NewManager(nil)
	serviceID := "prehook-deny-service"

	mockHook := &MockPreHook{
		ExecutePreFunc: func(ctx context.Context, req *ExecutionRequest) (Action, *ExecutionRequest, error) {
			return ActionDeny, nil, nil
		},
	}

	info := &ServiceInfo{
		Name:     "Hook Service",
		PreHooks: []PreCallHook{mockHook},
	}
	tm.AddServiceInfo(serviceID, info)

	mockTool := &MockTool{
		ToolFunc: func() *mcp_router_v1.Tool {
			return mcp_router_v1.Tool_builder{
				ServiceId: proto.String(serviceID),
				Name:      proto.String("tool"),
			}.Build()
		},
	}
	_ = tm.AddTool(mockTool)

	req := &ExecutionRequest{ToolName: serviceID + ".tool", ToolInputs: []byte(`{}`)}
	_, err := tm.ExecuteTool(context.Background(), req)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "tool execution denied by hook")
}

func TestToolManager_PostHook_Failure(t *testing.T) {
	t.Parallel()
	tm := NewManager(nil)
	serviceID := "posthook-fail-service"

	expectedErr := errors.New("post-hook failed")
	mockHook := &MockPostHook{
		ExecutePostFunc: func(ctx context.Context, req *ExecutionRequest, result any) (any, error) {
			return nil, expectedErr
		},
	}

	info := &ServiceInfo{
		Name:      "Hook Service",
		PostHooks: []PostCallHook{mockHook},
	}
	tm.AddServiceInfo(serviceID, info)

	mockTool := &MockTool{
		ToolFunc: func() *mcp_router_v1.Tool {
			return mcp_router_v1.Tool_builder{
				ServiceId: proto.String(serviceID),
				Name:      proto.String("tool"),
			}.Build()
		},
		ExecuteFunc: func(_ context.Context, _ *ExecutionRequest) (any, error) {
			return "success", nil
		},
	}
	_ = tm.AddTool(mockTool)

	req := &ExecutionRequest{ToolName: serviceID + ".tool", ToolInputs: []byte(`{}`)}
	_, err := tm.ExecuteTool(context.Background(), req)

	assert.Error(t, err)
	assert.Equal(t, expectedErr, err)
}
