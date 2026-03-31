package tool

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	v1 "github.com/mcpany/core/proto/api/v1"
	configv1 "github.com/mcpany/core/proto/config/v1"
)

func TestWebrtcTool_Close_And_ExecuteWithoutPool(t *testing.T) {
	tool := &v1.Tool{Name: "test-tool"}
	callDef := &configv1.WebrtcCallDefinition{}

	webrtcTool, err := NewWebrtcTool(tool, nil, "test-service", nil, callDef)
	require.NoError(t, err)

	err = webrtcTool.Close()
	assert.NoError(t, err)

	req := &ExecutionRequest{
		ToolName:   "test-tool",
		ToolInputs: []byte(`{"param1": "value1"}`),
	}
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	_, err = webrtcTool.Execute(ctx, req)
	assert.Error(t, err)
}

func TestWebrtcTool_StreamExecute(t *testing.T) {
	tool := &v1.Tool{Name: "test-tool"}
	callDef := &configv1.WebrtcCallDefinition{}

	webrtcTool, err := NewWebrtcTool(tool, nil, "test-service", nil, callDef)
	require.NoError(t, err)

	assert.False(t, webrtcTool.IsStreaming())

	req := &ExecutionRequest{
		ToolName:   "test-tool",
		ToolInputs: []byte(`{"param1": "value1"}`),
	}

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	ch, err := webrtcTool.StreamExecute(ctx, req)
	require.NoError(t, err)

	result := <-ch
	if errResult, ok := result.(error); ok {
		assert.Error(t, errResult)
	} else {
		t.Errorf("Expected an error from StreamExecute, got %v", result)
	}
}

func TestWebrtcTool_ExecuteWithPeerConnection_InvalidJSON(t *testing.T) {
	tool := &v1.Tool{Name: "test-tool"}
	callDef := &configv1.WebrtcCallDefinition{}
	webrtcTool, err := NewWebrtcTool(tool, nil, "test-service", nil, callDef)
	require.NoError(t, err)

	req := &ExecutionRequest{
		ToolName:   "test-tool",
		ToolInputs: []byte(`invalid json`),
	}

	ctx := context.Background()
	_, err = webrtcTool.executeWithPeerConnection(ctx, req, nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to unmarshal tool inputs")
}
