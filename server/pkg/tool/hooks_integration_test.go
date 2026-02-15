package tool

import (
	"context"
	"testing"

	busproto "github.com/mcpany/core/proto/bus"
	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/mcpany/core/server/pkg/bus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

// ptrIntegration is a helper to get a pointer to a value.
func ptrIntegration[T any](v T) *T {
	return &v
}

func TestToolManager_ExecuteTool_WithHooks(t *testing.T) {
	t.Parallel()
	// Setup Manager
	busProvider, err := bus.NewProvider(&busproto.MessageBus{})
	require.NoError(t, err)
	tm := NewManager(busProvider)

	// Define Tool
	toolName := "my-tool"
	serviceID := "service-1"

	protoTool := v1.Tool_builder{
		Name:      proto.String(toolName),
		ServiceId: proto.String(serviceID),
	}.Build()

	// ToolID is conventionally serviceID.toolName (sanitized)
	toolID := serviceID + "." + toolName

	// 1. Test PreCallHook (Policy) Deny
	t.Run("PreCallHook_Deny", func(t *testing.T) {
		mockTool := &MockTool{
			ToolFunc: func() *v1.Tool { return protoTool },
			// ExecuteFunc should not be called if hook denies
		}

		err := tm.AddTool(mockTool)
		require.NoError(t, err)
		tm.AddServiceInfo(serviceID, &ServiceInfo{
			Config: configv1.UpstreamServiceConfig_builder{
				PreCallHooks: []*configv1.CallHook{
					configv1.CallHook_builder{
						CallPolicy: configv1.CallPolicy_builder{
							DefaultAction: configv1.CallPolicy_DENY.Enum(),
						}.Build(),
					}.Build(),
				},
			}.Build(),
		})

		req := &ExecutionRequest{ToolName: toolID}
		_, err = tm.ExecuteTool(context.Background(), req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "denied")
	})

	// 2. Test PostCallHook (TextTruncation)

	// 3. Test CallPolicy (Legacy)
	t.Run("LegacyCallPolicy_Deny", func(t *testing.T) {
		mockTool := &MockTool{
			ToolFunc: func() *v1.Tool { return protoTool },
		}

		err := tm.AddTool(mockTool)
		require.NoError(t, err)
		tm.AddServiceInfo(serviceID, &ServiceInfo{
			Config: configv1.UpstreamServiceConfig_builder{
				CallPolicies: []*configv1.CallPolicy{
					configv1.CallPolicy_builder{
						DefaultAction: configv1.CallPolicy_DENY.Enum(),
					}.Build(),
				},
			}.Build(),
		})

		req := &ExecutionRequest{ToolName: toolID}
		_, err = tm.ExecuteTool(context.Background(), req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "denied")
	})

	// 4. Test Tool Management (Get, List, Clear)
	t.Run("ToolManagement", func(t *testing.T) {
		// Add duplicate check? AddTool overwrites?
		mockTool := &MockTool{
			ToolFunc: func() *v1.Tool { return protoTool },
		}
		err := tm.AddTool(mockTool)
		require.NoError(t, err)

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
			ExecuteFunc: func(ctx context.Context, req *ExecutionRequest, next ExecutionFunc) (any, error) {
				middlewareCalled = true
				return next(ctx, req)
			},
		}

		tmWithMw := NewManager(busProvider)
		tmWithMw.AddMiddleware(mw)

		mockToolMw := &MockTool{
			ToolFunc: func() *v1.Tool { return protoTool },
			ExecuteFunc: func(_ context.Context, _ *ExecutionRequest) (any, error) {
				return "ok", nil
			},
		}
		err := tmWithMw.AddTool(mockToolMw)
		require.NoError(t, err)

		req := &ExecutionRequest{ToolName: toolID}
		res, err := tmWithMw.ExecuteTool(context.Background(), req)
		require.NoError(t, err)
		assert.Equal(t, "ok", res)
		assert.True(t, middlewareCalled)
	})
}

type MockMiddleware struct {
	ExecuteFunc func(ctx context.Context, req *ExecutionRequest, next ExecutionFunc) (any, error)
}

func (m *MockMiddleware) Execute(
	ctx context.Context,
	req *ExecutionRequest,
	next ExecutionFunc,
) (any, error) {
	if m.ExecuteFunc != nil {
		return m.ExecuteFunc(ctx, req, next)
	}
	return next(ctx, req)
}
