package tool

import (
	"context"
	"testing"

	"github.com/pion/webrtc/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWebrtcTool_PeerConnectionWrapper_Close(t *testing.T) {
	pc, err := webrtc.NewPeerConnection(webrtc.Configuration{})
	require.NoError(t, err)

	wrapper := &peerConnectionWrapper{PeerConnection: pc}

	// Close the connection
	err = wrapper.Close()
	require.NoError(t, err)

	// Since state change might be async, it could technically be considered healthy for a very brief moment
	// But let's check it anyway. If it fails, we can add a small sleep or rely on pion docs.
	// Typically, closing it doesn't immediately reflect in ICEConnectionState without signaling,
	// but let's see. If it fails we'll adjust the test.
	assert.False(t, wrapper.IsHealthy(context.Background()))
}
