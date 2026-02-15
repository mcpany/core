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
		toolDef := pb.Tool_builder{
			Name:        proto.String("http-tool"),
			ServiceId:   proto.String("service"),
			Description: proto.String("desc"),
		}.Build()
		// Minimal setup for HTTPTool
		ht := NewHTTPTool(
			toolDef,
			pool.NewManager(),
			"service-id",
			nil,
			configv1.HttpCallDefinition_builder{}.Build(),
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
		toolDef := pb.Tool_builder{
			Name:      proto.String("mcp-tool"),
			ServiceId: proto.String("service"),
		}.Build()
		mt := NewMCPTool(
			toolDef,
			&MockMCPClient{}, // Fixed typo
			configv1.MCPCallDefinition_builder{}.Build(),
		)

		mcpTool := mt.MCPTool()
		assert.NotNil(t, mcpTool)
		assert.Equal(t, "service.mcp-tool", mcpTool.Name)
	})

	// Test OpenAPITool.MCPTool()
	t.Run("OpenAPITool", func(t *testing.T) {
		toolDef := pb.Tool_builder{
			Name:      proto.String("openapi-tool"),
			ServiceId: proto.String("service"),
		}.Build()
		ot := NewOpenAPITool(
			toolDef,
			&client.HTTPClientWrapper{},
			nil,
			"GET",
			"http://example.com",
			nil,
			configv1.OpenAPICallDefinition_builder{}.Build(),
		)

		mcpTool := ot.MCPTool()
		assert.NotNil(t, mcpTool)
		assert.Equal(t, "service.openapi-tool", mcpTool.Name)
	})

	// Test LocalCommandTool.MCPTool()
	t.Run("LocalCommandTool", func(t *testing.T) {
		toolDef := pb.Tool_builder{
			Name:      proto.String("cmd-tool"),
			ServiceId: proto.String("service"),
		}.Build()
		ct := NewLocalCommandTool(
			toolDef,
			&configv1.CommandLineUpstreamService{},
			configv1.CommandLineCallDefinition_builder{}.Build(),
			nil,
			"call-id",
		)

		mcpTool := ct.MCPTool() // It calls ConvertProtoToMCPTool
		assert.NotNil(t, mcpTool)
		assert.Equal(t, "service.cmd-tool", mcpTool.Name)
	})

	// Test WebrtcTool.MCPTool()
	t.Run("WebrtcTool", func(t *testing.T) {
		toolDef := pb.Tool_builder{
			Name:      proto.String("webrtc-tool"),
			ServiceId: proto.String("service"),
		}.Build()
		// Mock pool or nil
		wt, err := NewWebrtcTool(
			toolDef,
			nil,
			"service-id",
			nil,
			configv1.WebrtcCallDefinition_builder{}.Build(),
		)
		assert.NoError(t, err)

		mcpTool := wt.MCPTool()
		assert.NotNil(t, mcpTool)
		assert.Equal(t, "service.webrtc-tool", mcpTool.Name)
	})

    // Test CallableTool.MCPTool() (via BaseTool)
	t.Run("CallableTool", func(t *testing.T) {
        toolDef := configv1.ToolDefinition_builder{
            Name: proto.String("callable-tool"),
            ServiceId: proto.String("service"),
        }.Build()
        ct, err := NewCallableTool(
            toolDef,
            configv1.UpstreamServiceConfig_builder{}.Build(),
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
        toolDef := pb.Tool_builder{
            Name:      proto.String("websocket-tool"),
            ServiceId: proto.String("service"),
        }.Build()
        wst := NewWebsocketTool(
            toolDef,
            nil,
            "service-id",
            nil,
            configv1.WebsocketCallDefinition_builder{}.Build(),
        )

        mcpTool := wst.MCPTool()
        assert.NotNil(t, mcpTool)
        assert.Equal(t, "service.websocket-tool", mcpTool.Name)
    })
}

func TestTool_GetCacheConfig(t *testing.T) {
	t.Run("HTTPTool", func(t *testing.T) {
		cacheCfg := configv1.CacheConfig_builder{IsEnabled: proto.Bool(true)}.Build()
		ht := NewHTTPTool(
			&pb.Tool{},
			pool.NewManager(),
			"s",
			nil,
			configv1.HttpCallDefinition_builder{Cache: cacheCfg}.Build(),
			nil,
			nil,
			"",
		)
		assert.Equal(t, cacheCfg, ht.GetCacheConfig())
	})

	t.Run("MCPTool", func(t *testing.T) {
		cacheCfg := configv1.CacheConfig_builder{IsEnabled: proto.Bool(true)}.Build()
		mt := NewMCPTool(
			&pb.Tool{},
			nil,
			configv1.MCPCallDefinition_builder{Cache: cacheCfg}.Build(),
		)
		assert.Equal(t, cacheCfg, mt.GetCacheConfig())
	})

	t.Run("OpenAPITool", func(t *testing.T) {
		cacheCfg := configv1.CacheConfig_builder{IsEnabled: proto.Bool(true)}.Build()
		ot := NewOpenAPITool(
			&pb.Tool{},
			nil,
			nil,
			"",
			"",
			nil,
			configv1.OpenAPICallDefinition_builder{Cache: cacheCfg}.Build(),
		)
		assert.Equal(t, cacheCfg, ot.GetCacheConfig())
	})

	t.Run("CommandTool", func(t *testing.T) {
		cacheCfg := configv1.CacheConfig_builder{IsEnabled: proto.Bool(true)}.Build()
		// Note: CommandTool (remote) vs LocalCommandTool
		// NewCommandTool returns Tool interface.
	wt, err := NewWebrtcTool(
		pb.Tool_builder{Name: proto.String("tool"), UnderlyingMethodFqn: proto.String("WEBRTC http://127.0.0.1")}.Build(),
		nil, // No pool -> triggers executeWithoutPool
		"s",
		nil,
		configv1.WebrtcCallDefinition_builder{}.Build(),
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

	lct := NewLocalCommandTool(
		pb.Tool_builder{}.Build(),
		configv1.CommandLineUpstreamService_builder{}.Build(),
		configv1.CommandLineCallDefinition_builder{Cache: cacheCfg}.Build(),
		nil,
		"",
	)
	assert.Equal(t, cacheCfg, lct.GetCacheConfig())
	})

	t.Run("WebrtcTool", func(t *testing.T) {
		cacheCfg := configv1.CacheConfig_builder{IsEnabled: proto.Bool(true)}.Build()
		wt, err := NewWebrtcTool(
			&pb.Tool{},
			nil,
			"s",
			nil,
			configv1.WebrtcCallDefinition_builder{Cache: cacheCfg}.Build(),
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
		pb.Tool_builder{Name: proto.String("tool"), UnderlyingMethodFqn: proto.String("WEBRTC http://127.0.0.1")}.Build(),
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
