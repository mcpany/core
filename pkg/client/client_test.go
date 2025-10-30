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
	"context"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	configv1 "github.com/mcpxy/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/encoding/protojson"
)

func TestHttpClientWrapper(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	configJSON := `{"http_service": {"address": "` + server.URL[7:] + `"}}`
	config := &configv1.UpstreamServiceConfig{}
	require.NoError(t, protojson.Unmarshal([]byte(configJSON), config))

	client := &HttpClientWrapper{
		Client: &http.Client{},
		config: config,
	}

	t.Run("IsHealthy", func(t *testing.T) {
		assert.True(t, client.IsHealthy(context.Background()))
	})

	t.Run("Close", func(t *testing.T) {
		err := client.Close()
		assert.NoError(t, err)
	})
}

func TestGrpcClientWrapper(t *testing.T) {
	// Set up a dummy gRPC server to connect to
	lis, err := net.Listen("tcp", "localhost:0")
	require.NoError(t, err)

	server := grpc.NewServer()
	go func() {
		// This may return an error on server.Stop(), which is expected.
		_ = server.Serve(lis)
	}()
	defer server.Stop()

	conn, err := grpc.NewClient(lis.Addr().String(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)

	configJSON := `{"grpc_service": {"address": "` + lis.Addr().String() + `"}}`
	config := &configv1.UpstreamServiceConfig{}
	require.NoError(t, protojson.Unmarshal([]byte(configJSON), config))

	wrapper := &GrpcClientWrapper{
		ClientConn: conn,
		config:     config,
	}

	t.Run("IsHealthy_Initially", func(t *testing.T) {
		// Wait for the connection to be ready or idle. This is a more robust check.
		require.Eventually(t, func() bool {
			state := wrapper.GetState()
			return state == connectivity.Ready || state == connectivity.Idle
		}, time.Second*5, 10*time.Millisecond, "gRPC client should connect")
		assert.True(t, wrapper.IsHealthy(context.Background()))
	})

	t.Run("Close and IsHealthy", func(t *testing.T) {
		err := wrapper.Close()
		require.NoError(t, err)

		// The state should eventually become Shutdown.
		require.Eventually(t, func() bool {
			return wrapper.GetState() == connectivity.Shutdown
		}, time.Second, 10*time.Millisecond, "gRPC client state should be Shutdown")

		assert.False(t, wrapper.IsHealthy(context.Background()))
	})
}

func TestWebsocketClientWrapper(t *testing.T) {
	upgrader := websocket.Upgrader{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer conn.Close()
		// Just handle control messages for the health check
		conn.SetReadDeadline(time.Now().Add(time.Second * 5))
		for {
			if _, _, err := conn.NextReader(); err != nil {
				break
			}
		}
	}))
	defer server.Close()

	wsURL := "ws" + server.URL[len("http"):]
	conn, resp, err := websocket.DefaultDialer.Dial(wsURL, nil)
	require.NoError(t, err)
	defer func() {
		err := resp.Body.Close()
		require.NoError(t, err)
	}()

	wrapper := &WebsocketClientWrapper{Conn: conn}

	t.Run("IsHealthy_Connected", func(t *testing.T) {
		assert.True(t, wrapper.IsHealthy(context.Background()))
	})

	t.Run("Close and IsHealthy", func(t *testing.T) {
		err := wrapper.Close()
		require.NoError(t, err)
		// After closing, IsHealthy should fail
		assert.False(t, wrapper.IsHealthy(context.Background()))
	})
}
