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

	"bytes"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/mcpxy/core/pkg/logging"
	"github.com/mcpxy/core/pkg/mcpserver"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	gogrpc "google.golang.org/grpc"
)

func TestHealthCheck(t *testing.T) {
	t.Run("successful health check", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/healthz", r.URL.Path)
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		// Extract port from server URL
		_, port, err := net.SplitHostPort(strings.TrimPrefix(server.URL, "http://"))
		require.NoError(t, err)

		err = HealthCheck(port)
		assert.NoError(t, err)
	})

	t.Run("response body is read and closed", func(t *testing.T) {
		handlerFinished := make(chan bool)
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			// This handler will hang until the client reads the body and closes it.
			<-r.Context().Done()
			handlerFinished <- true
		}))
		defer server.Close()

		_, port, err := net.SplitHostPort(strings.TrimPrefix(server.URL, "http://"))
		require.NoError(t, err)

		err = HealthCheck(port)
		assert.NoError(t, err)

		select {
		case <-handlerFinished:
			// The handler finished, which means the client-side transport has
			// closed the request context, which it only does after the body is closed.
		case <-time.After(1 * time.Second):
			t.Fatal("HealthCheck did not read and close the response body in time")
		}
	})

	t.Run("failed health check with non-200 status", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer server.Close()

		_, port, err := net.SplitHostPort(strings.TrimPrefix(server.URL, "http://"))
		require.NoError(t, err)

		err = HealthCheck(port)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "health check failed with status code: 500")
	})

	t.Run("failed health check with connection error", func(t *testing.T) {
		// Find a free port, then close it, to ensure it's not listening
		l, err := net.Listen("tcp", "localhost:0")
		require.NoError(t, err)
		port := l.Addr().(*net.TCPAddr).Port
		l.Close()

		err = HealthCheck(fmt.Sprintf("%d", port))
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "health check failed:")
	})
}


func TestSetup(t *testing.T) {
	t.Run("with nil fs", func(t *testing.T) {
		logging.ForTestsOnlyResetLogger()
		var buf bytes.Buffer
		logging.Init(slog.LevelError, &buf)

		fs, err := setup(nil)
		assert.Error(t, err, "setup should return an error when fs is nil")
		assert.Nil(t, fs, "setup should return a nil fs when an error occurs")
		assert.True(t, strings.Contains(buf.String(), "setup called with nil afero.Fs"), "Error message should be logged")
	})

	t.Run("with existing fs", func(t *testing.T) {
		memFs := afero.NewMemMapFs()
		fs, err := setup(memFs)
		assert.NoError(t, err, "setup should not return an error for a valid fs")
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
        - schema:
            name: "echo"
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
		errChan <- app.Run(ctx, fs, false, "0", "0", []string{"/config.yaml"}, 5*time.Second)
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
	err = app.Run(ctx, fs, false, "0", "0", []string{"/config.yaml"}, 5*time.Second)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to load services from config")
}

func TestRun_EmptyConfig(t *testing.T) {
	fs := afero.NewMemMapFs()
	// Create an empty config file
	err := afero.WriteFile(fs, "/config.yaml", []byte(""), 0o644)
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	app := NewApplication()
	// This should not panic
	err = app.Run(ctx, fs, false, "0", "0", []string{"/config.yaml"}, 5*time.Second)
	require.NoError(t, err)
}

func TestRun_StdioMode(t *testing.T) {
	var stdioModeCalled bool
	mockStdioFunc := func(ctx context.Context, mcpSrv *mcpserver.Server) error {
		stdioModeCalled = true
		return fmt.Errorf("stdio mode error")
	}

	app := &Application{
		runStdioModeFunc: mockStdioFunc,
	}

	fs := afero.NewMemMapFs()
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	err := app.Run(ctx, fs, true, "0", "0", nil, 5*time.Second)

	assert.True(t, stdioModeCalled, "runStdioMode should have been called")
	assert.EqualError(t, err, "stdio mode error")
}

func TestRun_NoGrpcServer(t *testing.T) {
	fs := afero.NewMemMapFs()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	app := NewApplication()
	errChan := make(chan error, 1)
	go func() {
		errChan <- app.Run(ctx, fs, false, "0", "", nil, 5*time.Second)
	}()

	err := <-errChan
	assert.NoError(t, err, "app.Run should return nil on graceful shutdown")
}

func TestRun_ServerStartupErrors(t *testing.T) {
	app := NewApplication()

	t.Run("nil_fs_fail", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancel()

		err := app.Run(ctx, nil, false, "0", "0", nil, 5*time.Second)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to setup filesystem")
	})

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
		err = app.Run(ctx, fs, false, fmt.Sprintf("%d", port), "0", nil, 5*time.Second)
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
		err = app.Run(ctx, fs, false, "0", fmt.Sprintf("%d", port), nil, 5*time.Second)
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
	startGrpcServer(ctx, &wg, errChan, "TestGRPC", fmt.Sprintf(":%d", port), 5*time.Second, func(s *gogrpc.Server) {
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

			startGrpcServer(ctx, &wg, errChan, "TestGRPC_Race", fmt.Sprintf(":%d", port), 5*time.Second, func(s *gogrpc.Server) {})

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

func TestHTTPServer_ShutdownTimesOut(t *testing.T) {
	// This test verifies that the HTTP server's graceful shutdown waits for
	// the timeout duration when a request hangs.

	lis, err := net.Listen("tcp", "localhost:0")
	require.NoError(t, err, "Failed to find a free port.")
	port := lis.Addr().(*net.TCPAddr).Port
	lis.Close()

	ctx, cancel := context.WithCancel(context.Background())
	errChan := make(chan error, 1)
	var wg sync.WaitGroup

	handlerStarted := make(chan struct{})
	shutdownTimeout := 100 * time.Millisecond
	handlerSleep := 5 * time.Second

	startHTTPServer(ctx, &wg, errChan, "TestHTTP_Hang", fmt.Sprintf(":%d", port), http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		close(handlerStarted)
		time.Sleep(handlerSleep)
		w.WriteHeader(http.StatusOK)
	}), shutdownTimeout)

	time.Sleep(50 * time.Millisecond) // give server time to start

	go func() {
		_, _ = http.Get(fmt.Sprintf("http://localhost:%d", port))
	}()

	// Wait for the handler to receive the request before we shutdown
	select {
	case <-handlerStarted:
		// Great, handler is running
	case <-time.After(1 * time.Second):
		t.Fatal("timed out waiting for handler to start")
	}

	shutdownStartTime := time.Now()
	cancel()
	wg.Wait()
	shutdownDuration := time.Since(shutdownStartTime)

	// With the bug, shutdown is immediate. With the fix, it should wait for the timeout.
	// We expect the shutdown to take at least as long as the timeout.
	assert.GreaterOrEqual(t, shutdownDuration, shutdownTimeout, "Shutdown should take at least the shutdown timeout duration.")
	// And it should not wait for the full handler sleep duration.
	assert.Less(t, shutdownDuration, handlerSleep, "Shutdown should not wait for the handler to complete.")

	select {
	case err := <-errChan:
		require.NoError(t, err, "The HTTP server should shut down gracefully without errors.")
	default:
		// Expected outcome.
	}
}


func TestGRPCServer_GracefulShutdownHangs(t *testing.T) {
	// This test verifies that the gRPC server hangs on graceful shutdown if an
	// RPC is in progress, because GracefulStop() has no timeout.
	// The test is expected to FAIL by timing out before the fix is applied.
	// After the fix, the server will force a shutdown after a timeout,
	// and this test will PASS.

	// Find a free port.
	lis, err := net.Listen("tcp", "localhost:0")
	require.NoError(t, err, "Failed to find a free port.")
	port := lis.Addr().(*net.TCPAddr).Port
	lis.Close()

	ctx, cancel := context.WithCancel(context.Background())
	errChan := make(chan error, 1)
	var wg sync.WaitGroup

	// Start the gRPC server with a service that will hang.
	startGrpcServer(ctx, &wg, errChan, "TestGRPC_Hang", fmt.Sprintf(":%d", port), 1*time.Second, func(s *gogrpc.Server) {
		hangService := &mockHangService{hangTime: 5 * time.Second}
		desc := &gogrpc.ServiceDesc{
			ServiceName: "testhang.HangService",
			HandlerType: (*interface{})(nil),
			Methods: []gogrpc.MethodDesc{
				{
					MethodName: "Hang",
					Handler: func(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor gogrpc.UnaryServerInterceptor) (interface{}, error) {
						return srv.(*mockHangService).Hang(ctx, nil)
					},
				},
			},
			Streams:  []gogrpc.StreamDesc{},
			Metadata: "testhang.proto",
		}
		s.RegisterService(desc, hangService)
	})

	// Give the server time to start.
	time.Sleep(100 * time.Millisecond)

	// Make a call to the hanging RPC in a separate goroutine.
	go func() {
		conn, err := gogrpc.Dial(fmt.Sprintf("localhost:%d", port), gogrpc.WithInsecure(), gogrpc.WithBlock())
		if err != nil {
			t.Logf("Failed to dial gRPC server: %v", err)
			return
		}
		defer conn.Close()
		_ = conn.Invoke(context.Background(), "/testhang.HangService/Hang", &struct{}{}, &struct{}{})
	}()

	// Allow the RPC call to be initiated.
	time.Sleep(100 * time.Millisecond)

	// Trigger the graceful shutdown.
	cancel()

	// With the bug, wg.Wait() will return quickly, but the shutdown goroutine
	// inside startGrpcServer will hang. The test only works if startGrpcServer
	// is structured to wait for shutdown before calling wg.Done().
	// We expect this test to time out.
	wg.Wait()
}

// mockHangService is a mock gRPC service that has a method designed to hang
// for a specified duration. This is used to test graceful shutdown behavior
// under load.
type mockHangService struct {
	gogrpc.ServerStream
	hangTime time.Duration
}

// Hang is a mock RPC that simulates a long-running operation by sleeping
// for the configured hangTime.
func (s *mockHangService) Hang(ctx context.Context, req interface{}) (interface{}, error) {
	time.Sleep(s.hangTime)
	return &struct{}{}, nil
}

func TestGRPCServer_GracefulShutdownWithTimeout(t *testing.T) {
	// This test verifies that the gRPC server's graceful shutdown times out
	// correctly when a request hangs, preventing the server from blocking
	// indefinitely.

	// Find a free port to run the test server on.
	lis, err := net.Listen("tcp", "localhost:0")
	require.NoError(t, err, "Failed to find a free port.")
	port := lis.Addr().(*net.TCPAddr).Port
	lis.Close()

	ctx, cancel := context.WithCancel(context.Background())
	errChan := make(chan error, 1)
	var wg sync.WaitGroup

	// Start the gRPC server with a mock service that hangs.
	startGrpcServer(ctx, &wg, errChan, "TestGRPC_Hang", fmt.Sprintf(":%d", port), 50*time.Millisecond, func(s *gogrpc.Server) {
		// This service will hang for 10 seconds, which is much longer than our
		// shutdown timeout.
		hangService := &mockHangService{hangTime: 10 * time.Second}
		desc := &gogrpc.ServiceDesc{
			ServiceName: "testhang.HangService",
			HandlerType: (*interface{})(nil),
			Methods: []gogrpc.MethodDesc{
				{
					MethodName: "Hang",
					Handler: func(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor gogrpc.UnaryServerInterceptor) (interface{}, error) {
						return srv.(*mockHangService).Hang(ctx, nil)
					},
				},
			},
			Streams:  []gogrpc.StreamDesc{},
			Metadata: "testhang.proto",
		}
		s.RegisterService(desc, hangService)
	})

	// Give the server a moment to start up.
	time.Sleep(100 * time.Millisecond)

	// In a separate goroutine, make a call to the hanging RPC.
	go func() {
		conn, err := gogrpc.Dial(fmt.Sprintf("localhost:%d", port), gogrpc.WithInsecure(), gogrpc.WithBlock())
		if err != nil {
			// If we can't connect, there's no point in continuing the test.
			t.Logf("Failed to dial gRPC server: %v", err)
			return
		}
		defer conn.Close()

		// This call will hang until the server is forcefully shut down.
		_ = conn.Invoke(context.Background(), "/testhang.HangService/Hang", &struct{}{}, &struct{}{})
	}()

	// Allow the RPC call to be initiated.
	time.Sleep(100 * time.Millisecond)

	// Trigger the graceful shutdown.
	cancel()

	// This WaitGroup should be released quickly, as the shutdown should not
	// wait for the hanging RPC to complete. If the bug is present, this test
	// will time out here.
	wg.Wait()

	// The test should complete without error, as the timeout mechanism allows
	// the server to shut down without waiting for the hanging connection.
	select {
	case err := <-errChan:
		require.NoError(t, err, "The gRPC server should shut down gracefully without errors.")
	default:
		// Expected outcome.
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
		startHTTPServer(ctx, &wg, errChan, "TestHTTP_Hang", fmt.Sprintf("localhost:%d", port), nil, 5*time.Second)

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

	startGrpcServer(ctx, &wg, errChan, "TestGRPC", ":0", 5*time.Second, func(_ *gogrpc.Server) {})

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
			startGrpcServer(ctx, &wg, errChan, "TestGRPC_NoRace", fmt.Sprintf(":%d", port), 5*time.Second, func(s *gogrpc.Server) {})

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
