// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/mcpany/core/pkg/pool"
	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/pion/webrtc/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func TestWebrtcTool_Close_WithPool(t *testing.T) {
	// Create a tool with pool
	pm := pool.NewManager()
	wtWithPool, err := NewWebrtcTool(
		&v1.Tool{Name: proto.String("test")},
		pm,
		"serviceID",
		nil,
		&configv1.WebrtcCallDefinition{},
	)
	require.NoError(t, err)
	err = wtWithPool.Close()
	assert.NoError(t, err)
}

func TestWebrtcTool_MCPTool_Caching(t *testing.T) {
	toolDef := &v1.Tool{
		Name: proto.String("test_tool"),
		Annotations: &v1.ToolAnnotations{
			// Description is not a field of ToolAnnotations?
			// Checking proto definition, it might be separate.
			// Re-checking existing tests or code.
		},
	}
	wt := &WebrtcTool{
		tool: toolDef,
	}

	// First call should create it
	mcpTool1 := wt.MCPTool()
	require.NotNil(t, mcpTool1)
	assert.Contains(t, mcpTool1.Name, "test_tool")

	// Second call should return cached instance
	mcpTool2 := wt.MCPTool()
	assert.Equal(t, mcpTool1, mcpTool2)
}

func TestWebrtcTool_Execute_WithoutPool(t *testing.T) {
	// Mock context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	wt := &WebrtcTool{
		tool: &v1.Tool{
			UnderlyingMethodFqn: proto.String("WEBRTC http://localhost:12345/signal"),
		},
		parameters: []*configv1.WebrtcParameterMapping{},
	}

	req := &ExecutionRequest{
		ToolInputs: json.RawMessage(`{}`),
	}

	// This is expected to fail because there's no signaling server at localhost:12345
	// But it should cover the executeWithoutPool -> newPeerConnection path.
	_, err := wt.Execute(ctx, req)
	assert.Error(t, err)
}

func TestWebrtcTool_PoolCreation(t *testing.T) {
	pm := pool.NewManager()
	callDef := &configv1.WebrtcCallDefinition{
		Parameters: []*configv1.WebrtcParameterMapping{},
	}

	// Create first instance
	wt1, err := NewWebrtcTool(
		&v1.Tool{Name: proto.String("tool1")},
		pm,
		"service_a",
		nil,
		callDef,
	)
	require.NoError(t, err)
	require.NotNil(t, wt1.webrtcPool)

	// Create second instance for SAME service, should share pool
	wt2, err := NewWebrtcTool(
		&v1.Tool{Name: proto.String("tool2")},
		pm,
		"service_a",
		nil,
		callDef,
	)
	require.NoError(t, err)
	assert.Equal(t, wt1.webrtcPool, wt2.webrtcPool)
}

func TestPeerConnectionWrapper(t *testing.T) {
	// Create a real peer connection to test wrapper methods
	pc, err := webrtc.NewPeerConnection(webrtc.Configuration{})
	require.NoError(t, err)

	w := &peerConnectionWrapper{PeerConnection: pc}

	// IsHealthy
	assert.False(t, w.IsHealthy(context.Background())) // Not connected

	// Close
	err = w.Close()
	assert.NoError(t, err)

	// Close nil
	wNil := &peerConnectionWrapper{PeerConnection: nil}
	assert.NoError(t, wNil.Close())
	assert.False(t, wNil.IsHealthy(context.Background()))
}
