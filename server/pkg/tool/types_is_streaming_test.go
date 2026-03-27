package tool

import (
	"testing"

	"github.com/stretchr/testify/assert"

	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
)

func TestLocalCommandTool_IsStreamingReal(t *testing.T) {
	t.Parallel()
	toolProto := &v1.Tool{}
	tool := NewLocalCommandTool(toolProto, nil, nil, nil, "")
	assert.False(t, tool.IsStreaming())
}

func TestCommandTool_IsStreamingReal(t *testing.T) {
	t.Parallel()
	toolProto := &v1.Tool{}
	tool := NewCommandTool(toolProto, nil, nil, nil, "")
	assert.False(t, tool.IsStreaming())
}

func TestHTTPTool_IsStreamingReal(t *testing.T) {
	t.Parallel()
	toolProto := &v1.Tool{}
	tool := NewHTTPTool(toolProto, nil, "", nil, &configv1.HttpCallDefinition{}, nil, nil, "")
	assert.False(t, tool.IsStreaming())
}

func TestOpenAPITool_IsStreamingReal(t *testing.T) {
	t.Parallel()
	toolProto := &v1.Tool{}
	tool := NewOpenAPITool(toolProto, nil, nil, "", "", nil, &configv1.OpenAPICallDefinition{})
	assert.False(t, tool.IsStreaming())
}

func TestMCPTool_IsStreamingReal(t *testing.T) {
	t.Parallel()
	toolProto := &v1.Tool{}
	tool := NewMCPTool(toolProto, nil, &configv1.MCPCallDefinition{})
	assert.False(t, tool.IsStreaming())
}
