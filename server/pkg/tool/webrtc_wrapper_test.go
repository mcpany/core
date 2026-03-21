package tool

import (
	"context"
	"testing"
	"time"

	"github.com/pion/webrtc/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWebrtcTool_PeerConnectionWrapper_IsHealthy_And_Close(t *testing.T) {
	wrapper := &peerConnectionWrapper{PeerConnection: nil}
	assert.False(t, wrapper.IsHealthy(context.Background()))

	pc, err := webrtc.NewPeerConnection(webrtc.Configuration{})
	require.NoError(t, err)

	wrapper = &peerConnectionWrapper{PeerConnection: pc}

	// Initially, connection is in 'New' state, which is considered healthy
	assert.True(t, wrapper.IsHealthy(context.Background()))

	// Close the connection via the wrapper
	err = wrapper.Close()
	require.NoError(t, err)

	// Since state change might be async, wait briefly for the state to transition
	time.Sleep(100 * time.Millisecond)
	assert.False(t, wrapper.IsHealthy(context.Background()))
}
