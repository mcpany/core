package client

import (
	"context"
	"time"

	"github.com/gorilla/websocket"
)

// WebsocketClientWrapper wraps a *websocket.Conn to adapt it for use in a
// connection pool, implementing the pool.ClosableClient interface.
type WebsocketClientWrapper struct {
	Conn *websocket.Conn
}

// IsHealthy checks if the underlying WebSocket connection is still active. It
// sends a ping message with a short deadline to verify the connection's liveness.
func (w *WebsocketClientWrapper) IsHealthy(_ context.Context) bool {
	// Send a ping to check the connection.
	// A short deadline is used to prevent blocking.
	err := w.Conn.WriteControl(websocket.PingMessage, []byte{}, time.Now().Add(time.Second*2))
	return err == nil
}

// Close terminates the underlying WebSocket connection.
//
// Returns an error if the operation fails.
func (w *WebsocketClientWrapper) Close() error {
	return w.Conn.Close()
}
