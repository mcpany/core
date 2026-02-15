package tool

import (
	"context"

	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// MockTool is a mock implementation of the Tool interface for testing purposes.
type MockTool struct {
	ToolFunc           func() *v1.Tool
	MCPToolFunc        func() *mcp.Tool
	ExecuteFunc        func(ctx context.Context, req *ExecutionRequest) (any, error)
	GetCacheConfigFunc func() *configv1.CacheConfig
}

// Tool returns the protobuf definition of the mock tool.
//
// Returns the result.
func (m *MockTool) Tool() *v1.Tool {
	if m.ToolFunc != nil {
		return m.ToolFunc()
	}
	return &v1.Tool{}
}

// MCPTool returns the MCP tool definition.
//
// Returns the result.
func (m *MockTool) MCPTool() *mcp.Tool {
	if m.MCPToolFunc != nil {
		return m.MCPToolFunc()
	}
	return nil
}

// Execute calls the mock ExecuteFunc if set, otherwise returns nil.
//
// ctx is the context for the request.
// req is the request object.
//
// Returns the result.
// Returns an error if the operation fails.
func (m *MockTool) Execute(ctx context.Context, req *ExecutionRequest) (any, error) {
	if m.ExecuteFunc != nil {
		return m.ExecuteFunc(ctx, req)
	}
	return nil, nil
}

// GetCacheConfig calls the mock GetCacheConfigFunc if set, otherwise returns nil.
//
// Returns the result.
func (m *MockTool) GetCacheConfig() *configv1.CacheConfig {
	if m.GetCacheConfigFunc != nil {
		return m.GetCacheConfigFunc()
	}
	return nil
}
