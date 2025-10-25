//go:build e2e

package examples

import (
	"os/exec"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/mcpxy/core/tests/integration"
	"github.com/stretchr/testify/require"
)

func TestWebsocketExample(t *testing.T) {
	root, err := integration.GetProjectRoot()
	require.NoError(t, err)

	cmd := exec.Command("/bin/bash", "start.sh")
	cmd.Dir = root + "/examples/upstream/websocket"
	err = cmd.Start()
	require.NoError(t, err)
	defer cmd.Process.Kill()

	// Wait for the server to start
	time.Sleep(2 * time.Second)

	// Connect to the WebSocket server
	var conn *websocket.Conn
	for i := 0; i < 5; i++ {
		conn, _, err = websocket.DefaultDialer.Dial("ws://localhost:8082/echo", nil)
		if err == nil {
			break
		}
		time.Sleep(1 * time.Second)
	}
	require.NoError(t, err)
	defer conn.Close()

	// Send a message to the server
	message := "Hello, world!"
	err = conn.WriteMessage(websocket.TextMessage, []byte(message))
	require.NoError(t, err)

	// Read the message back from the server
	_, p, err := conn.ReadMessage()
	require.NoError(t, err)

	// Verify that the message is the same
	require.Equal(t, message, string(p))
}
