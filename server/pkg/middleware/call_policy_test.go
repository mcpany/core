// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware_test

import (
	"context"
	"encoding/json"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/mcpany/core/server/pkg/middleware"
	"github.com/mcpany/core/server/pkg/tool"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/protobuf/proto"
)

// Helper to handle builders or direct structs
// Since we had issues with builders, we use direct structs.

type callPolicyMockTool struct {
	toolProto *v1.Tool
	mock.Mock
}

func (m *callPolicyMockTool) Execute(ctx context.Context, req *tool.ExecutionRequest) (any, error) {
	args := m.Called(ctx, req)
	return args.Get(0), args.Error(1)
}

func (m *callPolicyMockTool) Tool() *v1.Tool {
	return m.toolProto
}

func (m *callPolicyMockTool) GetCacheConfig() *configv1.CacheConfig {
	return nil
}

func (m *callPolicyMockTool) MCPTool() *mcp.Tool {
	t, _ := tool.ConvertProtoToMCPTool(m.toolProto)
	return t
}

type callPolicyMockToolManager struct {
	tool.ManagerInterface
	mock.Mock
}

func (m *callPolicyMockToolManager) GetServiceInfo(serviceID string) (*tool.ServiceInfo, bool) {
	args := m.Called(serviceID)
	if args.Get(0) == nil {
		return nil, args.Bool(1)
	}
	return args.Get(0).(*tool.ServiceInfo), args.Bool(1)
}

func (m *callPolicyMockToolManager) GetTool(toolName string) (tool.Tool, bool) {
	args := m.Called(toolName)
	if args.Get(0) == nil {
		return nil, args.Bool(1)
	}
	return args.Get(0).(tool.Tool), args.Bool(1)
}

func (m *callPolicyMockToolManager) ListServices() []*tool.ServiceInfo {
	return nil
}

func TestCallPolicyMiddleware(t *testing.T) {
	const successResult = "success"

	setup := func(policies []*configv1.CallPolicy) (*middleware.CallPolicyMiddleware, *callPolicyMockToolManager, *callPolicyMockTool) {
		mockToolManager := &callPolicyMockToolManager{}
		cpMiddleware := middleware.NewCallPolicyMiddleware(mockToolManager)

		// Use builder for Tool as it seems to work in other tests or standard struct
		toolProto := v1.Tool_builder{
			ServiceId: proto.String("service"),
		}.Build()
		mockTool := &callPolicyMockTool{toolProto: toolProto}

		compiledPolicies, _ := tool.CompileCallPolicies(policies)
		svcConfig := &configv1.UpstreamServiceConfig{}
		svcConfig.SetCallPolicies(policies)
		serviceInfo := &tool.ServiceInfo{
			Name: "test-service",
			Config: svcConfig,
			CompiledPolicies: compiledPolicies,
		}

		mockToolManager.On("GetTool", mock.Anything).Return(mockTool, true)
		mockToolManager.On("GetServiceInfo", "service").Return(serviceInfo, true)

		return cpMiddleware, mockToolManager, mockTool
	}

	t.Run("no policies -> allowed", func(t *testing.T) {
		cpMiddleware, _, mockTool := setup(nil)

		req := &tool.ExecutionRequest{
			ToolName:   "service.test-tool",
			ToolInputs: json.RawMessage(`{}`),
		}
		ctx := tool.NewContextWithTool(context.Background(), mockTool)

		nextCalled := false
		next := func(_ context.Context, _ *tool.ExecutionRequest) (any, error) {
			nextCalled = true
			return successResult, nil
		}

		result, err := cpMiddleware.Execute(ctx, req, next)
		assert.NoError(t, err)
		assert.Equal(t, successResult, result)
		assert.True(t, nextCalled)
	})

	t.Run("name regex deny -> blocked", func(t *testing.T) {
		policy := configv1.CallPolicy_builder{
			DefaultAction: configv1.CallPolicy_ALLOW.Enum(),
			Rules: []*configv1.CallPolicyRule{
				configv1.CallPolicyRule_builder{
					Action:    configv1.CallPolicy_DENY.Enum(),
					NameRegex: proto.String(".*test-tool"),
				}.Build(),
			},
		}.Build()

		cpMiddleware, _, mockTool := setup([]*configv1.CallPolicy{policy})

		req := &tool.ExecutionRequest{
			ToolName:   "service.test-tool",
			ToolInputs: json.RawMessage(`{}`),
		}
		ctx := tool.NewContextWithTool(context.Background(), mockTool)

		next := func(_ context.Context, _ *tool.ExecutionRequest) (any, error) {
			return successResult, nil
		}

		_, err := cpMiddleware.Execute(ctx, req, next)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "execution denied by policy")
	})

	t.Run("argument regex deny -> blocked", func(t *testing.T) {
		policy := configv1.CallPolicy_builder{
			DefaultAction: configv1.CallPolicy_ALLOW.Enum(),
			Rules: []*configv1.CallPolicyRule{
				configv1.CallPolicyRule_builder{
					Action:        configv1.CallPolicy_DENY.Enum(),
					ArgumentRegex: proto.String(".*dangerous.*"),
				}.Build(),
			},
		}.Build()

		cpMiddleware, _, mockTool := setup([]*configv1.CallPolicy{policy})

		req := &tool.ExecutionRequest{
			ToolName:   "service.test-tool",
			ToolInputs: json.RawMessage(`{"cmd": "dangerous command"}`),
		}
		ctx := tool.NewContextWithTool(context.Background(), mockTool)

		next := func(_ context.Context, _ *tool.ExecutionRequest) (any, error) {
			return successResult, nil
		}

		_, err := cpMiddleware.Execute(ctx, req, next)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "execution denied by policy")
	})

	t.Run("argument regex mismatch -> allowed", func(t *testing.T) {
		policy := configv1.CallPolicy_builder{
			DefaultAction: configv1.CallPolicy_ALLOW.Enum(),
			Rules: []*configv1.CallPolicyRule{
				configv1.CallPolicyRule_builder{
					Action:        configv1.CallPolicy_DENY.Enum(),
					ArgumentRegex: proto.String(".*dangerous.*"),
				}.Build(),
			},
		}.Build()

		cpMiddleware, _, mockTool := setup([]*configv1.CallPolicy{policy})

		req := &tool.ExecutionRequest{
			ToolName:   "service.test-tool",
			ToolInputs: json.RawMessage(`{"cmd": "safe command"}`),
		}
		ctx := tool.NewContextWithTool(context.Background(), mockTool)

		nextCalled := false
		next := func(_ context.Context, _ *tool.ExecutionRequest) (any, error) {
			nextCalled = true
			return successResult, nil
		}

		result, err := cpMiddleware.Execute(ctx, req, next)
		assert.NoError(t, err)
		assert.Equal(t, successResult, result)
		assert.True(t, nextCalled)
	})

	t.Run("default deny -> blocked", func(t *testing.T) {
		policy := configv1.CallPolicy_builder{
			DefaultAction: configv1.CallPolicy_DENY.Enum(),
		}.Build()

		cpMiddleware, _, mockTool := setup([]*configv1.CallPolicy{policy})

		req := &tool.ExecutionRequest{
			ToolName:   "service.test-tool",
			ToolInputs: json.RawMessage(`{}`),
		}
		ctx := tool.NewContextWithTool(context.Background(), mockTool)

		next := func(_ context.Context, _ *tool.ExecutionRequest) (any, error) {
			return successResult, nil
		}

		_, err := cpMiddleware.Execute(ctx, req, next)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "execution denied by policy")
	})

	t.Run("default deny but allowed by rule -> allowed", func(t *testing.T) {
		policy := configv1.CallPolicy_builder{
			DefaultAction: configv1.CallPolicy_DENY.Enum(),
			Rules: []*configv1.CallPolicyRule{
				configv1.CallPolicyRule_builder{
					Action:    configv1.CallPolicy_ALLOW.Enum(),
					NameRegex: proto.String(".*test-tool"),
				}.Build(),
			},
		}.Build()

		cpMiddleware, _, mockTool := setup([]*configv1.CallPolicy{policy})

		req := &tool.ExecutionRequest{
			ToolName:   "service.test-tool",
			ToolInputs: json.RawMessage(`{}`),
		}
		ctx := tool.NewContextWithTool(context.Background(), mockTool)

		nextCalled := false
		next := func(_ context.Context, _ *tool.ExecutionRequest) (any, error) {
			nextCalled = true
			return successResult, nil
		}

		result, err := cpMiddleware.Execute(ctx, req, next)
		assert.NoError(t, err)
		assert.Equal(t, successResult, result)
		assert.True(t, nextCalled)
	})
}
