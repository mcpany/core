// Copyright 2024 Author(s) of MCP Any
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"context"
	"io"
	"log/slog"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/mcpany/core/pkg/logging"
	app_testing "github.com/mcpany/core/proto/testing/app"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// mockBlockingService is a mock gRPC service that blocks until the context is canceled.
type mockBlockingService struct {
	app_testing.UnimplementedGreeterServer
	started chan struct{}
}

func (s *mockBlockingService) SayHello(ctx context.Context, in *app_testing.HelloRequest) (*app_testing.HelloReply, error) {
	close(s.started)
	<-ctx.Done()
	return nil, status.Error(codes.Canceled, "context canceled")
}

// TestStartGrpcServer_GracefulShutdownTimeout verifies that the gRPC server is
// forcefully stopped when the graceful shutdown period times out.
func TestStartGrpcServer_GracefulShutdownTimeout(t *testing.T) {
	logging.Init(slog.LevelDebug, io.Discard)
	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup
	errChan := make(chan error, 1)
	lis, err := net.Listen("tcp", ":0")
	if err != nil {
		t.Fatalf("failed to listen: %v", err)
	}

	service := &mockBlockingService{started: make(chan struct{})}

	startGrpcServer(ctx, &wg, errChan, "TestServer", lis, 10*time.Millisecond, func(s *grpc.Server) {
		app_testing.RegisterGreeterServer(s, service)
	})

	// Create a client and make a call to the service to ensure it's running
	conn, err := grpc.Dial(lis.Addr().String(), grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithBlock())
	if err != nil {
		t.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := app_testing.NewGreeterClient(conn)

	_, err = c.SayHello(context.Background(), &app_testing.HelloRequest{Name: "world"})
	if err != nil && status.Code(err) != codes.Canceled {
		t.Fatalf("could not greet: %v", err)
	}

	// Ensure the service's SayHello method has been called before we try to shut down the server
	<-service.started

	// Cancel the context to initiate shutdown
	cancel()

	// Wait for the server to shut down
	wg.Wait()

	// Check that no errors were reported
	assert.Len(t, errChan, 0, "no errors should be reported")
}
