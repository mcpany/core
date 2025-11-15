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
	"net"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	gogrpc "google.golang.org/grpc"
)

// TestStartGrpcServer_PanicHandling verifies that the startGrpcServer function
// gracefully handles panics that occur during the gRPC server initialization
// and registration phase.
//
// This test sets up a scenario where the registration function, which is
// responsible for setting up the gRPC services, intentionally panics. The test
// ensures that:
//
//  1. The panic is caught and does not crash the test process.
//  2. An error, indicating the panic, is sent to the provided error channel.
//  3. The function's WaitGroup counter is correctly decremented, signaling
//     that the goroutine has completed its execution.
//
// This test is crucial for ensuring the robustness of the server, as it
// validates that critical failures during startup are handled in a controlled
// manner, allowing the main application to manage the failure gracefully.
func TestStartGrpcServer_PanicHandling(t *testing.T) {
	// 1. Setup
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	var wg sync.WaitGroup
	errChan := make(chan error, 1)

	// Create a dummy net.Listener
	lis, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		t.Fatalf("Failed to create listener: %v", err)
	}
	defer lis.Close()

	// 2. Execution
	// This registration function will panic, simulating a failure during service setup.
	registerFunc := func(s *gogrpc.Server) {
		panic("registration failed")
	}

	startGrpcServer(ctx, &wg, errChan, "TestServer", lis, 1*time.Second, registerFunc)

	// 3. Verification
	select {
	case err := <-errChan:
		// Check that the error indicates a panic occurred.
		assert.Contains(t, err.Error(), "panic during gRPC service registration", "Expected error message to contain mention of a panic")
		assert.Contains(t, err.Error(), "registration failed", "Expected error message to contain the panic message")
	case <-time.After(2 * time.Second):
		t.Fatal("Test timed out, expected an error to be sent to the error channel")
	}

	// The WaitGroup should be done, as the goroutine should exit after the panic.
	// We use a channel to wait for the WaitGroup to be done to avoid a race condition.
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		// The WaitGroup was correctly handled.
	case <-time.After(2 * time.Second):
		t.Fatal("Test timed out, expected WaitGroup to be done")
	}
}
