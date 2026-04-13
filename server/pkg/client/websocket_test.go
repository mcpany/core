package client

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"time"
)

func TestWebsocketClientWrapper(t *testing.T) {
	upgrader := websocket.Upgrader{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer conn.Close()

		conn.SetPingHandler(func(appData string) error {
		    return conn.WriteControl(websocket.PongMessage, []byte(appData), validTime())
		})

		for {
			_, _, err := conn.ReadMessage()
			if err != nil {
				break
			}
		}
	}))
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")
	conn, resp, err := websocket.DefaultDialer.Dial(wsURL, nil)
	require.NoError(t, err)
	defer resp.Body.Close()

	wrapper := &WebsocketClientWrapper{Conn: conn}

	assert.True(t, wrapper.IsHealthy(context.Background()))
	assert.NoError(t, wrapper.Close())

	// Check healthy after close should fail
	assert.False(t, wrapper.IsHealthy(context.Background()))
}

func validTime() time.Time {
    return time.Now().Add(time.Second)
}
