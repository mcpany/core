//go:build e2e

package examples

import (
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/mcpxy/core/tests/integration"
	"github.com/stretchr/testify/require"
)

func TestWebRTCExample(t *testing.T) {
	root, err := integration.GetProjectRoot()
	require.NoError(t, err)

	cmd := exec.Command("/bin/bash", "start.sh")
	cmd.Dir = root + "/examples/upstream/webrtc"

	output, err := cmd.CombinedOutput()
	require.NoError(t, err)

	// It can take a moment for the client to connect and receive the message.
	// We'll retry a few times to make sure we don't have a race condition.
	require.Eventually(t, func() bool {
		return strings.Contains(string(output), "Message from data channel: Hello, world!")
	}, 5*time.Second, 1*time.Second, "Expected to receive 'Hello, world!' message")
}
