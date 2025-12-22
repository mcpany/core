package tool

import (
	"context"
	"testing"

	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
)

type MockMCPServerProvider struct {
	server *mcp.Server
}

func (m *MockMCPServerProvider) Server() *mcp.Server {
	return m.server
}

func TestManager_AddTool_Errors(t *testing.T) {
	tm := NewManager(nil)

	// Test empty service ID
	toolDef := &v1.Tool{
		Name:      proto.String("tool"),
		ServiceId: proto.String(""),
	}
	mockTool := &MockTool{
		ToolFunc: func() *v1.Tool { return toolDef },
	}
	err := tm.AddTool(mockTool)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "service ID cannot be empty")

	// Test sanitization error?
	// Sanitization usually just replaces chars, hard to fail unless name is super weird or implementation changes.
	// We can trust util.SanitizeToolName coverage elsewhere or skip for now.
}

func TestManager_AddTool_WithServer(t *testing.T) {
	tm := NewManager(nil)

	// Setup real MCP server
	impl := &mcp.Implementation{
		Name:    "test-server",
		Version: "1.0.0",
	}
	opts := &mcp.ServerOptions{
		HasTools: true,
	}
	server := mcp.NewServer(impl, opts)

	provider := &MockMCPServerProvider{server: server}
	tm.SetMCPServer(provider)

	toolDef := &v1.Tool{
		Name:        proto.String("test-tool"),
		ServiceId:   proto.String("service-1"),
		Description: proto.String("A test tool"),
		InputSchema: &structpb.Struct{
			Fields: map[string]*structpb.Value{
				"type": structpb.NewStringValue("object"),
			},
		},
	}
	mockTool := &MockTool{
		ToolFunc: func() *v1.Tool { return toolDef },
	}

	err := tm.AddTool(mockTool)
	require.NoError(t, err)

	// Verify tool is in manager
	_, ok := tm.GetTool("service-1.test-tool")
	assert.True(t, ok)

	// Verify tool is in server?
	// mcp.Server doesn't expose ListTools easily for inspection without a client?
	// But at least we covered the registration code path.
}

// Additional tests for context/execution flow?
func TestManager_ExecuteTool_NotFound(t *testing.T) {
	tm := NewManager(nil)
	_, err := tm.ExecuteTool(context.Background(), &ExecutionRequest{ToolName: "missing"})
	assert.Error(t, err)
}

func TestManager_ExecuteTool_Chain(t *testing.T) {
	// Test middleware chain
	tm := NewManager(nil)

	toolDef := &v1.Tool{Name: proto.String("t"), ServiceId: proto.String("s")}
	mt := &MockTool{
		ToolFunc: func() *v1.Tool { return toolDef },
		ExecuteFunc: func(_ context.Context, _ *ExecutionRequest) (any, error) {
			return "final-result", nil
		},
	}

	_ = tm.AddTool(mt) // s.t

	// Add middleware
	mw := &MockMiddleware{
		ExecuteFunc: func(ctx context.Context, req *ExecutionRequest, next ExecutionFunc) (any, error) {
			return next(ctx, req)
		},
	}
	tm.AddMiddleware(mw)

	res, err := tm.ExecuteTool(context.Background(), &ExecutionRequest{ToolName: "s.t"})
	assert.NoError(t, err)
	assert.Equal(t, "final-result", res)
}
