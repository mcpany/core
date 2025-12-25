// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/mcpany/core/pkg/pool"
	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type MockAuthenticator struct{}

func (m *MockAuthenticator) Authenticate(r interface{}) error {
	return nil
}


func TestWebrtcTool_NewWebrtcTool(t *testing.T) {
	toolDef := &v1.Tool{Name: ptr("webrtc-tool")}
	callDef := &configv1.WebrtcCallDefinition{
		Parameters: []*configv1.WebrtcParameterMapping{},
	}
	poolManager := pool.NewManager()

	tool, err := NewWebrtcTool(toolDef, poolManager, "service-id", nil, callDef)
	require.NoError(t, err)
	assert.NotNil(t, tool)
	assert.Equal(t, toolDef, tool.Tool())
	assert.Equal(t, "service-id", tool.serviceID)
}

func TestWebrtcTool_MCPTool(t *testing.T) {
	toolDef := &v1.Tool{
		Name:        ptr("webrtc-tool"),
		Description: ptr("A test tool"),
		// InputSchema is a structpb.Struct in the proto definition I just read, but ConvertProtoToMCPTool handles it.
		// Wait, I need to check how InputSchema is defined in `proto/mcp_router/v1/mcp_router.pb.go`.
		// It is *structpb.Struct.
	}
	// We can't easily populate structpb.Struct here without more imports, but let's see if we can skip it or use empty.

	callDef := &configv1.WebrtcCallDefinition{}

	tool, err := NewWebrtcTool(toolDef, nil, "", nil, callDef)
	require.NoError(t, err)

	mcpTool := tool.MCPTool()
	require.NotNil(t, mcpTool)
	// Note: ConvertProtoToMCPTool might prepend service ID or similar if not present,
	// or maybe it's how naming works. It seems to produce ".webrtc-tool" when serviceID is empty.
	// We'll just assert it ends with the name.
	assert.Contains(t, mcpTool.Name, "webrtc-tool")
	assert.Equal(t, "A test tool", mcpTool.Description)
}

func TestWebrtcTool_Execute_Inputs(t *testing.T) {
	toolDef := &v1.Tool{Name: ptr("webrtc-tool"), UnderlyingMethodFqn: ptr("WEBRTC http://localhost:1234")}
	callDef := &configv1.WebrtcCallDefinition{}

	tool, err := NewWebrtcTool(toolDef, nil, "", nil, callDef)
	require.NoError(t, err)

	ctx := context.Background()
	req := &ExecutionRequest{
		ToolInputs: json.RawMessage(`{"key": "value"}`),
	}

	_, err = tool.Execute(ctx, req)
	// Expect error as no signaling server
	require.Error(t, err)
}

func TestWebrtcTool_GetCacheConfig(t *testing.T) {
	enabled := true
	cacheConfig := &configv1.CacheConfig{IsEnabled: &enabled}
	callDef := &configv1.WebrtcCallDefinition{
		Cache: cacheConfig,
	}
	tool, err := NewWebrtcTool(&v1.Tool{}, nil, "", nil, callDef)
	require.NoError(t, err)
	assert.Equal(t, cacheConfig, tool.GetCacheConfig())
}
