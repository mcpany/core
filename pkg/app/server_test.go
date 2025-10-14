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

package app

import (
	"context"
	"fmt"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/mcpxy/core/pkg/mcpserver"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	gogrpc "google.golang.org/grpc"
)

func TestSetup(t *testing.T) {
	t.Run("with nil fs", func(t *testing.T) {
		fs := setup(nil)
		assert.NotNil(t, fs, "setup should return a valid fs even if input is nil")
		_, ok := fs.(*afero.OsFs)
		assert.True(t, ok, "setup should default to OsFs")
	})

	t.Run("with existing fs", func(t *testing.T) {
		memFs := afero.NewMemMapFs()
		fs := setup(memFs)
		assert.Equal(t, memFs, fs, "setup should return the provided fs")
	})
}

func TestRun_ServerMode(t *testing.T) {
	fs := afero.NewMemMapFs()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// Create a dummy config file
	configContent := `
upstream_services:
  - name: "test-http-service"
    http_service:
      address: "http://localhost:8080"
      calls:
        - operation_id: "echo"
          endpoint_path: "/echo"
          method: "HTTP_METHOD_POST"
`
	err := afero.WriteFile(fs, "/config.yaml", []byte(configContent), 0o644)
	require.NoError(t, err)

	app := NewApplication()
	errChan := make(chan error, 1)
	go func() {
		// Use ephemeral ports by passing "0"
		// The test will hang if we use a real port that's not available.
		// We expect the Run function to exit gracefully when the context is canceled.
		errChan <- app.Run(ctx, fs, false, "0", "0", []string{"/config.yaml"})
	}()

	// We expect the server to run until the context is canceled, at which point it should
	// shut down gracefully and return nil.
	err = <-errChan
	assert.NoError(t, err, "app.Run should return nil on graceful shutdown")
}

func TestRun_ConfigLoadError(t *testing.T) {
	fs := afero.NewMemMapFs()
	// Create a malformed config file
	err := afero.WriteFile(fs, "/config.yaml", []byte("malformed yaml:"), 0o644)
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	app := NewApplication()
	err = app.Run(ctx, fs, false, "0", "0", []string{"/config.yaml"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to load services from config")
}

func TestRun_StdioMode(t *testing.T) {
	var stdioModeCalled bool
	mockStdioFunc := func(ctx context.Context, mcpSrv *mcpserver.Server) error {
		stdioModeCalled = true
		return nil
	}

	app := &Application{
		runStdioModeFunc: mockStdioFunc,
	}

	fs := afero.NewMemMapFs()
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	app.Run(ctx, fs, true, "0", "0", nil)

	assert.True(t, stdioModeCalled, "runStdioMode should have been called")
}

func TestRun_NoGrpcServer(t *testing.T) {
	fs := afero.NewMemMapFs()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	app := NewApplication()
	errChan := make(chan error, 1)
	go func() {
		errChan <- app.Run(ctx, fs, false, "0", "", nil)
	}()

	err := <-errChan
	assert.NoError(t, err, "app.Run should return nil on graceful shutdown")
}

func TestRun_ServerStartupErrors(t *testing.T) {
	app := NewApplication()

	t.Run("http_server_fail", func(t *testing.T) {
		// Find a free port and occupy it
		l, err := net.Listen("tcp", "localhost:0")
		require.NoError(t, err)
		defer l.Close()
		port := l.Addr().(*net.TCPAddr).Port

		fs := afero.NewMemMapFs()
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		// Attempt to run the server on the occupied port
		err = app.Run(ctx, fs, false, fmt.Sprintf("%d", port), "0", nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to start a server")
	})

	t.Run("grpc_server_fail", func(t *testing.T) {
		// Find a free port and occupy it
		l, err := net.Listen("tcp", "localhost:0")
		require.NoError(t, err)
		defer l.Close()
		port := l.Addr().(*net.TCPAddr).Port

		fs := afero.NewMemMapFs()
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		// Attempt to run the server on the occupied port
		err = app.Run(ctx, fs, false, "0", fmt.Sprintf("%d", port), nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to start a server")
	})
}

func TestGRPCServer_PortReleasedImmediatelyAfterShutdown(t *testing.T) {
	// Find an available port for the gRPC server to listen on.
	lis, err := net.Listen("tcp", "localhost:0")
	require.NoError(t, err, "Failed to find a free port.")
	port := lis.Addr().(*net.TCPAddr).Port
	// We close the listener immediately and just use the port number.
	lis.Close()

	ctx, cancel := context.WithCancel(context.Background())
	errChan := make(chan error, 1)
	var wg sync.WaitGroup

	// Start the gRPC server in a goroutine.
	startGrpcServer(ctx, &wg, errChan, "TestGRPC", fmt.Sprintf(":%d", port), func(s *gogrpc.Server) {
		// No services need to be registered for this test.
	})

	// Wait for the server to start by polling the port.
	require.Eventually(t, func() bool {
		conn, err := net.DialTimeout("tcp", fmt.Sprintf("localhost:%d", port), 10*time.Millisecond)
		if err != nil {
			return false
		}
		conn.Close()
		return true
	}, 1*time.Second, 10*time.Millisecond, "gRPC server did not start in time")

	// Cancel the context to initiate a graceful shutdown.
	cancel()
	// Wait for the server to fully shut down.
	wg.Wait()

	// After shutdown, attempt to listen on the same port again.
	// If the original listener was properly closed, this should succeed immediately.
	lis, err = net.Listen("tcp", fmt.Sprintf("localhost:%d", port))
	require.NoError(t, err, "The port should be available for reuse immediately after the server has shut down.")
	if lis != nil {
		lis.Close()
	}
}
