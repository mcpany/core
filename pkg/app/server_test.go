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

import (
	"bytes"
	"log/slog"
	"strings"

	"github.com/mcpxy/core/pkg/logging"
)

func TestSetup(t *testing.T) {
	t.Run("with nil fs", func(t *testing.T) {
		logging.ForTestsOnlyResetLogger()
		var buf bytes.Buffer
		logging.Init(slog.LevelWarn, &buf)

		fs := setup(nil)
		assert.NotNil(t, fs, "setup should return a valid fs even if input is nil")
		_, ok := fs.(*afero.OsFs)
		assert.True(t, ok, "setup should default to OsFs")

		assert.True(t, strings.Contains(buf.String(), "setup called with nil afero.Fs"), "Warning message should be logged")
	})

	t.Run("with existing fs", func(t *testing.T) {
		logging.ForTestsOnlyResetLogger()
		var buf bytes.Buffer
		logging.Init(slog.LevelWarn, &buf)

		memFs := afero.NewMemMapFs()
		fs := setup(memFs)
		assert.Equal(t, memFs, fs, "setup should return the provided fs")
		assert.False(t, strings.Contains(buf.String(), "setup called with nil afero.Fs"), "Warning message should not be logged")
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

func TestGRPCServer_PortReleasedAfterShutdown(t *testing.T) {
	// Find an available port for the gRPC server to listen on.
	lis, err := net.Listen("tcp", "localhost:0")
	require.NoError(t, err, "Failed to find a free port.")
	port := lis.Addr().(*net.TCPAddr).Port
	// We close the listener immediately and just use the port number.
	// This is to ensure the port is available for the gRPC server to use.
	lis.Close()

	ctx, cancel := context.WithCancel(context.Background())
	errChan := make(chan error, 1)
	var wg sync.WaitGroup

	// Start the gRPC server in a goroutine.
	startGrpcServer(ctx, &wg, errChan, "TestGRPC", fmt.Sprintf(":%d", port), func(s *gogrpc.Server) {
		// No services need to be registered for this test.
	})

	// Allow some time for the server to start up.
	time.Sleep(100 * time.Millisecond)

	// Cancel the context to initiate a graceful shutdown.
	cancel()
	// Wait for the server to fully shut down.
	wg.Wait()

	// Check if any errors occurred during startup or shutdown.
	select {
	case err := <-errChan:
		require.NoError(t, err, "The gRPC server should not have returned an error.")
	default:
		// No error, which is the expected outcome.
	}

	// After shutdown, attempt to listen on the same port again.
	// If the original listener was properly closed, this should succeed.
	lis, err = net.Listen("tcp", fmt.Sprintf("localhost:%d", port))
	require.NoError(t, err, "The port should be available for reuse after the server has shut down.")
	if lis != nil {
		lis.Close()
	}
}

func TestGRPCServer_FastShutdownRace(t *testing.T) {
	// This test is designed to be flaky if the race condition exists.
	// We run it multiple times to increase the chance of catching it.
	for i := 0; i < 20; i++ {
		t.Run(fmt.Sprintf("iteration_%d", i), func(t *testing.T) {
			lis, err := net.Listen("tcp", "localhost:0")
			require.NoError(t, err)
			port := lis.Addr().(*net.TCPAddr).Port
			lis.Close() // Close immediately, we just needed a free port.

			ctx, cancel := context.WithCancel(context.Background())
			errChan := make(chan error, 2)
			var wg sync.WaitGroup

			startGrpcServer(ctx, &wg, errChan, "TestGRPC_Race", fmt.Sprintf(":%d", port), func(s *gogrpc.Server) {})

			// Immediately cancel the context. This creates a race between
			// the server starting up and shutting down.
			cancel()

			wg.Wait() // Wait for the server goroutine to finish.

			close(errChan)
			for err := range errChan {
				// The race condition would manifest as the server trying to use a listener
				// that has already been closed by the parent goroutine exiting.
				assert.NotContains(t, err.Error(), "use of closed network connection", "gRPC server tried to use a closed listener on iteration %d", i)
			}
		})
	}
}

func TestHTTPServer_HangOnListenError(t *testing.T) {
	// Find a free port and occupy it
	l, err := net.Listen("tcp", "localhost:0")
	require.NoError(t, err)
	defer l.Close()
	port := l.Addr().(*net.TCPAddr).Port

	// This channel will be used to signal that the test is complete
	done := make(chan bool)

	go func() {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		errChan := make(chan error, 1)
		var wg sync.WaitGroup

		// This call should hang because wg.Done() is never called in the error case
		startHTTPServer(ctx, &wg, errChan, "TestHTTP_Hang", fmt.Sprintf("localhost:%d", port), nil)

		// The test will hang here waiting for the goroutine to finish
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		// The test completed without hanging, which is the expected behavior after the fix.
	case <-time.After(2 * time.Second):
		t.Fatal("Test hung for 2 seconds. The bug is still present.")
	}
}

func TestGRPCServer_GracefulShutdown(t *testing.T) {
	errChan := make(chan error, 1)
	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup

	startGrpcServer(ctx, &wg, errChan, "TestGRPC", ":0", func(_ *gogrpc.Server) {})

	// Immediately cancel to trigger shutdown
	cancel()
	wg.Wait()

	select {
	case err := <-errChan:
		assert.NoError(t, err, "Graceful shutdown should not produce an error")
	default:
		// No error, which is what we want
	}
}

func TestGRPCServer_ShutdownWithoutRace(t *testing.T) {
	// This test is designed to fail if the double-close issue is present.
	// It runs the shutdown sequence multiple times to ensure stability.
	for i := 0; i < 10; i++ {
		t.Run(fmt.Sprintf("iteration_%d", i), func(t *testing.T) {
			lis, err := net.Listen("tcp", "localhost:0")
			require.NoError(t, err)
			port := lis.Addr().(*net.TCPAddr).Port
			lis.Close()

			ctx, cancel := context.WithCancel(context.Background())
			errChan := make(chan error, 1)
			var wg sync.WaitGroup

			// Start the gRPC server.
			startGrpcServer(ctx, &wg, errChan, "TestGRPC_NoRace", fmt.Sprintf(":%d", port), func(s *gogrpc.Server) {})

			// Give the server a moment to start listening.
			time.Sleep(20 * time.Millisecond)

			// Trigger graceful shutdown.
			cancel()
			wg.Wait()

			// Check for errors. The double-close would likely cause a "use of closed network connection" error.
			select {
			case err := <-errChan:
				require.NoError(t, err, "Shutdown should be clean and not produce an error.")
			default:
				// This is the expected path.
			}
		})
	}
}
