/*
 * Copyright 2025 Author(s) of MCP Any
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

package integration

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
)

func TestWaitForGRPCReady(t *testing.T) {
	t.Parallel()

	lis, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	addr := lis.Addr().String()

	server := grpc.NewServer()
	grpc_health_v1.RegisterHealthServer(server, health.NewServer())

	serverStopped := make(chan struct{})
	go func() {
		defer close(serverStopped)
		if err := server.Serve(lis); err != nil {
			if !strings.Contains(err.Error(), "use of closed network connection") {
				t.Logf("gRPC server error: %v", err)
			}
		}
	}()

	t.Run("succeeds when server is ready", func(t *testing.T) {
		assert.True(t, WaitForGRPCReady(t, addr, 5*time.Second), "WaitForGRPCReady should succeed for a running server")
	})

	t.Run("fails when server is not ready in time", func(t *testing.T) {
		freePort := FindFreePort(t)
		address := fmt.Sprintf("127.0.0.1:%d", freePort)
		assert.False(t, WaitForGRPCReady(t, address, 50*time.Millisecond), "WaitForGRPCReady should fail for a non-running server")
	})

	server.Stop()
	<-serverStopped
}

func TestWaitForWebsocketReady(t *testing.T) {
	t.Parallel()

	lis, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	addr := lis.Addr().String()

	server := &http.Server{
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			upgrader := websocket.Upgrader{}
			_, _ = upgrader.Upgrade(w, r, nil)
		}),
	}

	serverStopped := make(chan struct{})
	go func() {
		defer close(serverStopped)
		if err := server.Serve(lis); err != http.ErrServerClosed {
			t.Logf("Websocket server error: %v", err)
		}
	}()

	wsURL := "ws://" + addr

	t.Run("succeeds when server is ready", func(t *testing.T) {
		assert.True(t, WaitForWebsocketReady(t, wsURL, 5*time.Second), "WaitForWebsocketReady should succeed for a running server")
	})

	t.Run("fails when server is not ready in time", func(t *testing.T) {
		freePort := FindFreePort(t)
		address := fmt.Sprintf("ws://127.0.0.1:%d", freePort)
		assert.False(t, WaitForWebsocketReady(t, address, 50*time.Millisecond), "WaitForWebsocketReady should fail for a non-running server")
	})

	server.Close()
	<-serverStopped
}

func TestStartWebsocketEchoServer(t *testing.T) {
	t.Parallel()
	wsInfo := StartWebsocketEchoServer(t)
	defer wsInfo.CleanupFunc()

	conn, _, err := websocket.DefaultDialer.Dial(wsInfo.URL, nil)
	require.NoError(t, err)
	defer conn.Close()

	message := []byte("hello")
	err = conn.WriteMessage(websocket.TextMessage, message)
	require.NoError(t, err)

	_, p, err := conn.ReadMessage()
	require.NoError(t, err)
	assert.Equal(t, message, p)
}

func TestRegisterServiceFunctions(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode.")
	}

	serverInfo := StartInProcessMCPANYServer(t, "TestRegisterServiceFunctions")
	defer serverInfo.CleanupFunc()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	t.Run("RegisterHTTPService", func(t *testing.T) {
		RegisterHTTPService(t, serverInfo.RegistrationClient, "http-test", "http://example.com", "get-data", "/data", "GET", nil)
		tools, err := serverInfo.ListTools(ctx)
		require.NoError(t, err)
		assert.Contains(t, tools.Tools, "get-data")
	})

	t.Run("RegisterWebsocketService", func(t *testing.T) {
		RegisterWebsocketService(t, serverInfo.RegistrationClient, "ws-test", "ws://example.com", "ws-echo", nil)
		tools, err := serverInfo.ListTools(ctx)
		require.NoError(t, err)
		assert.Contains(t, tools.Tools, "ws-echo")
	})

	t.Run("RegisterWebrtcService", func(t *testing.T) {
		RegisterWebrtcService(t, serverInfo.RegistrationClient, "webrtc-test", "http://example.com/signal", "webrtc-data", nil)
		tools, err := serverInfo.ListTools(ctx)
		require.NoError(t, err)
		assert.Contains(t, tools.Tools, "webrtc-data")
	})

	t.Run("RegisterGRPCService", func(t *testing.T) {
		RegisterGRPCService(t, serverInfo.RegistrationClient, "grpc-test", "localhost:50051", nil)
	})

	t.Run("RegisterOpenAPIService", func(t *testing.T) {
		specPath := filepath.Join(ProjectRoot(t), "examples", "upstream", "openapi", "petstore.yaml")
		RegisterOpenAPIService(t, serverInfo.RegistrationClient, "openapi-test", specPath, "http://petstore.swagger.io/v2", nil)
		tools, err := serverInfo.ListTools(ctx)
		require.NoError(t, err)
		assert.Contains(t, tools.Tools, "listPets")
	})

	t.Run("RegisterStdioService", func(t *testing.T) {
		RegisterStdioService(t, serverInfo.RegistrationClient, "stdio-test", "echo", true, "hello")
	})

	t.Run("RegisterStdioServiceWithSetup", func(t *testing.T) {
		tmpDir := t.TempDir()
		scriptPath := filepath.Join(tmpDir, "test.sh")
		err := os.WriteFile(scriptPath, []byte(`#!/bin/sh
	echo '[{"name": "stdio-setup-test.test-tool"}]'`), 0755)
		require.NoError(t, err)
		RegisterStdioServiceWithSetup(t, serverInfo.RegistrationClient, "stdio-setup-test", scriptPath, true, tmpDir, "", nil)
	})

	t.Run("RegisterHTTPServiceWithJSONRPC", func(t *testing.T) {
		RegisterHTTPServiceWithJSONRPC(t, serverInfo.JSONRPCEndpoint+"/mcp", "http-jsonrpc-test", "http://example.com", "get-json-data", "/json-data", "POST", nil)
		tools, err := serverInfo.ListTools(ctx)
		require.NoError(t, err)
		assert.Contains(t, tools.Tools, "get-json-data")
	})
}

func TestManagedProcessCmd(t *testing.T) {
	mp := NewManagedProcess(t, "test", "echo", []string{"hello"}, nil)
	assert.NotNil(t, mp.Cmd())
	assert.Equal(t, "echo", filepath.Base(mp.Cmd().Path))
}

func TestMCPANYTestServerInfo_ListAndCallTools(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode.")
	}

	serverInfo := StartInProcessMCPANYServer(t, "TestMCPANYTestServerInfo_ListAndCallTools")
	defer serverInfo.CleanupFunc()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	RegisterHTTPService(t, serverInfo.RegistrationClient, "http-test", "http://example.com", "get-data", "/data", "GET", nil)

	tools, err := serverInfo.ListTools(ctx)
	require.NoError(t, err)
	assert.Contains(t, tools.Tools, "get-data")

	_, err = serverInfo.CallTool(ctx, nil)
	assert.Error(t, err)
}
