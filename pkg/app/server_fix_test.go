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

package app

import (
	"context"
	"fmt"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	gogrpc "google.golang.org/grpc"
)

// TestGRPCServer_GracefulShutdownWithTimeout_Fix confirms that the gRPC server's graceful
// shutdown times out as expected when a request is hanging, preventing the
// server from blocking indefinitely.
func TestGRPCServer_GracefulShutdownWithTimeout_Fix(t *testing.T) {
	// Find a free port to run the test server on.
	lis, err := net.Listen("tcp", "localhost:0")
	require.NoError(t, err, "Failed to find a free port.")
	port := lis.Addr().(*net.TCPAddr).Port
	lis.Close()

	ctx, cancel := context.WithCancel(context.Background())
	errChan := make(chan error, 1)
	var wg sync.WaitGroup

	// Start the gRPC server with a mock service that hangs.
	lis, err = net.Listen("tcp", fmt.Sprintf("localhost:%d", port))
	require.NoError(t, err)
	shutdownTimeout := 50 * time.Millisecond
	startGrpcServer(ctx, &wg, errChan, "TestGRPC_Hang", lis, shutdownTimeout, func(s *gogrpc.Server) {
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
	shutdownStartTime := time.Now()
	cancel()
	wg.Wait()
	shutdownDuration := time.Since(shutdownStartTime)
	require.Less(t, shutdownDuration, 5*time.Second)

	// The test should complete without error, as the timeout mechanism allows
	// the server to shut down without waiting for the hanging connection.
	select {
	case err := <-errChan:
		require.NoError(t, err, "The gRPC server should shut down gracefully without errors.")
	default:
		// Expected outcome.
	}
}
