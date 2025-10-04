/*
 * Copyright 2025 Author(s) of MCP-XY
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package client

import (
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
func (w *WebsocketClientWrapper) IsHealthy() bool {
	// Send a ping to check the connection.
	// A short deadline is used to prevent blocking.
	err := w.Conn.WriteControl(websocket.PingMessage, []byte{}, time.Now().Add(time.Second*2))
	return err == nil
}

// Close terminates the underlying WebSocket connection.
func (w *WebsocketClientWrapper) Close() error {
	return w.Conn.Close()
}
