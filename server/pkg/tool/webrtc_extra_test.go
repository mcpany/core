package tool

import (
	"context"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWebrtcTool_Execute_NoPoolFallback(t *testing.T) {
	t.Setenv("MCPANY_WEBRTC_DISABLE_STUN", "true")

	toolDef := &v1.Tool{}
	toolDef.SetName("test-webrtc-no-pool")

	// Set an invalid URL so it fails cleanly with HTTP request error
	toolDef.SetUnderlyingMethodFqn("WEBRTC http://127.0.0.1:0/invalid")

	callDef := &configv1.WebrtcCallDefinition{}

	// Pass a nil poolManager, which will result in webrtcPool being nil inside WebrtcTool
	wt, err := NewWebrtcTool(toolDef, nil, "webrtc-service-no-pool", nil, callDef)
	require.NoError(t, err)

	req := &ExecutionRequest{
		ToolName:   "test-webrtc-no-pool",
		ToolInputs: []byte(`{}`),
	}

	// Because webrtcPool is nil, executeWithoutPool is called
	_, err = wt.Execute(context.Background(), req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to send offer to signaling server")
}
