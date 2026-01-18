// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"context"
	"os"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	pb "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/mcpany/core/server/pkg/client"
	"github.com/mcpany/core/server/pkg/pool"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

func boolPtr(b bool) *bool {
	return &b
}

func TestCacheControl(t *testing.T) {
	ctx := context.Background()

	// Test default missing
	_, ok := GetCacheControl(ctx)
	assert.False(t, ok)

	// Test existing
	cc := &CacheControl{Action: ActionAllow}
	ctx = NewContextWithCacheControl(ctx, cc)
	got, ok := GetCacheControl(ctx)
	assert.True(t, ok)
	assert.Equal(t, cc, got)
	assert.Equal(t, ActionAllow, got.Action)
}

func TestTool_MCPTool_Method(t *testing.T) {
	// Test HTTPTool.MCPTool()
	t.Run("HTTPTool", func(t *testing.T) {
		toolDef := &pb.Tool{
			Name:        proto.String("http-tool"),
			ServiceId:   proto.String("service"),
			Description: proto.String("desc"),
			InputSchema: nil, // Add if needed for conversion test
		}
		// Minimal setup for HTTPTool
		ht := NewHTTPTool(
			toolDef,
			pool.NewManager(),
			"service-id",
			nil,
			&configv1.HttpCallDefinition{},
			nil,
			nil,
			"call-id",
		)

		mcpTool := ht.MCPTool()
		assert.NotNil(t, mcpTool)
		assert.Equal(t, "service.http-tool", mcpTool.Name)
		assert.Equal(t, "desc", mcpTool.Description)

		// Call again to test sync.Once
		mcpTool2 := ht.MCPTool()
		assert.Equal(t, mcpTool, mcpTool2)
	})

	// Test MCPTool.MCPTool()
	t.Run("MCPTool", func(t *testing.T) {
		toolDef := &pb.Tool{
			Name:      proto.String("mcp-tool"),
			ServiceId: proto.String("service"),
		}
		mt := NewMCPTool(
			toolDef,
			&MockMCPClient{}, // Fixed typo
			&configv1.MCPCallDefinition{},
		)

		mcpTool := mt.MCPTool()
		assert.NotNil(t, mcpTool)
		assert.Equal(t, "service.mcp-tool", mcpTool.Name)
	})

	// Test OpenAPITool.MCPTool()
	t.Run("OpenAPITool", func(t *testing.T) {
		toolDef := &pb.Tool{
			Name:      proto.String("openapi-tool"),
			ServiceId: proto.String("service"),
		}
		ot := NewOpenAPITool(
			toolDef,
			&client.HTTPClientWrapper{},
			nil,
			"GET",
			"http://example.com",
			nil,
			&configv1.OpenAPICallDefinition{},
		)

		mcpTool := ot.MCPTool()
		assert.NotNil(t, mcpTool)
		assert.Equal(t, "service.openapi-tool", mcpTool.Name)
	})

	// Test LocalCommandTool.MCPTool()
	t.Run("LocalCommandTool", func(t *testing.T) {
		toolDef := &pb.Tool{
			Name:      proto.String("cmd-tool"),
			ServiceId: proto.String("service"),
		}
		ct, err := NewLocalCommandTool(
			toolDef,
			&configv1.CommandLineUpstreamService{},
			&configv1.CommandLineCallDefinition{},
			nil,
			"call-id",
		)
		assert.NoError(t, err)

		mcpTool := ct.MCPTool() // It calls ConvertProtoToMCPTool
		assert.NotNil(t, mcpTool)
		assert.Equal(t, "service.cmd-tool", mcpTool.Name)
	})

	// Test WebrtcTool.MCPTool()
	t.Run("WebrtcTool", func(t *testing.T) {
		toolDef := &pb.Tool{
			Name:      proto.String("webrtc-tool"),
			ServiceId: proto.String("service"),
		}
		// Mock pool or nil
		wt, err := NewWebrtcTool(
			toolDef,
			nil,
			"service-id",
			nil,
			&configv1.WebrtcCallDefinition{},
		)
		assert.NoError(t, err)

		mcpTool := wt.MCPTool()
		assert.NotNil(t, mcpTool)
		assert.Equal(t, "service.webrtc-tool", mcpTool.Name)
	})

    // Test CallableTool.MCPTool() (via BaseTool)
	t.Run("CallableTool", func(t *testing.T) {
        toolDef := &configv1.ToolDefinition{
            Name: proto.String("callable-tool"),
            ServiceId: proto.String("service"),
        }
        ct, err := NewCallableTool(
            toolDef,
            &configv1.UpstreamServiceConfig{},
            nil, // Callable
            nil,
            nil,
        )
        assert.NoError(t, err)

        mcpTool := ct.MCPTool()
        assert.NotNil(t, mcpTool)
        assert.Equal(t, "service.callable-tool", mcpTool.Name)
    })

    // Test WebsocketTool.MCPTool()
    t.Run("WebsocketTool", func(t *testing.T) {
        toolDef := &pb.Tool{
            Name:      proto.String("websocket-tool"),
            ServiceId: proto.String("service"),
        }
        wst := NewWebsocketTool(
            toolDef,
            nil,
            "service-id",
            nil,
            &configv1.WebsocketCallDefinition{},
        )

        mcpTool := wst.MCPTool()
        assert.NotNil(t, mcpTool)
        assert.Equal(t, "service.websocket-tool", mcpTool.Name)
    })
}

func TestTool_GetCacheConfig(t *testing.T) {
	t.Run("HTTPTool", func(t *testing.T) {
		cacheCfg := &configv1.CacheConfig{IsEnabled: boolPtr(true)}
		ht := NewHTTPTool(
			&pb.Tool{},
			pool.NewManager(),
			"s",
			nil,
			&configv1.HttpCallDefinition{Cache: cacheCfg},
			nil,
			nil,
			"",
		)
		assert.Equal(t, cacheCfg, ht.GetCacheConfig())
	})

	t.Run("MCPTool", func(t *testing.T) {
		cacheCfg := &configv1.CacheConfig{IsEnabled: boolPtr(true)}
		mt := NewMCPTool(
			&pb.Tool{},
			nil,
			&configv1.MCPCallDefinition{Cache: cacheCfg},
		)
		assert.Equal(t, cacheCfg, mt.GetCacheConfig())
	})

	t.Run("OpenAPITool", func(t *testing.T) {
		cacheCfg := &configv1.CacheConfig{IsEnabled: boolPtr(true)}
		ot := NewOpenAPITool(
			&pb.Tool{},
			nil,
			nil,
			"",
			"",
			nil,
			&configv1.OpenAPICallDefinition{Cache: cacheCfg},
		)
		assert.Equal(t, cacheCfg, ot.GetCacheConfig())
	})

	t.Run("CommandTool", func(t *testing.T) {
		cacheCfg := &configv1.CacheConfig{IsEnabled: boolPtr(true)}
		// Note: CommandTool (remote) vs LocalCommandTool
		// NewCommandTool returns Tool interface.
		ct := NewCommandTool(
			&pb.Tool{},
			&configv1.CommandLineUpstreamService{},
			&configv1.CommandLineCallDefinition{Cache: cacheCfg},
			nil,
			"",
		)
		assert.Equal(t, cacheCfg, ct.GetCacheConfig())

		lct, err := NewLocalCommandTool(
			&pb.Tool{},
			&configv1.CommandLineUpstreamService{},
			&configv1.CommandLineCallDefinition{Cache: cacheCfg},
			nil,
			"",
		)
		assert.NoError(t, err)
		assert.Equal(t, cacheCfg, lct.GetCacheConfig())
	})

	t.Run("WebrtcTool", func(t *testing.T) {
		cacheCfg := &configv1.CacheConfig{IsEnabled: boolPtr(true)}
		wt, err := NewWebrtcTool(
			&pb.Tool{},
			nil,
			"s",
			nil,
			&configv1.WebrtcCallDefinition{Cache: cacheCfg},
		)
		assert.NoError(t, err)
		assert.Equal(t, cacheCfg, wt.GetCacheConfig())
	})
}

func TestManager_GetTool_NotFound(t *testing.T) {
	// This covers the !ok path in GetTool
	// Although simple, it ensures coverage
	m := NewManager(nil)
	tool, ok := m.GetTool("non-existent")
	assert.False(t, ok)
	assert.Nil(t, tool)
}

func TestWebrtcTool_Close_And_ExecuteWithoutPool(t *testing.T) {
	// Set env to disable STUN for faster/safer test
	os.Setenv("MCPANY_WEBRTC_DISABLE_STUN", "true")
	defer os.Unsetenv("MCPANY_WEBRTC_DISABLE_STUN")

	wt, err := NewWebrtcTool(
		&pb.Tool{Name: proto.String("tool"), UnderlyingMethodFqn: proto.String("WEBRTC http://localhost")},
		nil, // No pool -> triggers executeWithoutPool
		"s",
		nil,
		&configv1.WebrtcCallDefinition{},
	)
	assert.NoError(t, err)

	// Test Close (no pool)
	assert.NoError(t, wt.Close())

	// Trigger Execute -> executeWithoutPool
	// This will try to create connection and fail or hang?
	// It calls newPeerConnection (succeeds with disabled STUN)
	// Then executeWithPeerConnection -> unmarshal input -> ... -> http request
	// We pass invalid input JSON to fail early in executeWithPeerConnection,
	// verifying that executeWithoutPool was called and called executeWithPeerConnection.

	req := &ExecutionRequest{
		ToolName:   "tool",
		ToolInputs: []byte("invalid json"),
	}

	_, err = wt.Execute(context.Background(), req)
	assert.Error(t, err)
	// executeWithPeerConnection returns "failed to unmarshal tool inputs"
	assert.Contains(t, err.Error(), "failed to unmarshal tool inputs")
}
