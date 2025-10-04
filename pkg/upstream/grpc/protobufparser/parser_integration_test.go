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

package protobufparser

import (
	"context"
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	v1 "github.com/mcpxy/core/proto/examples/calculator/v1"
)

type mockCalculatorServer struct {
	v1.UnimplementedCalculatorServiceServer
}

func (s *mockCalculatorServer) Add(ctx context.Context, req *v1.AddRequest) (*v1.AddResponse, error) {
	resp := &v1.AddResponse{}
	resp.SetResult(req.GetA() + req.GetB())
	return resp, nil
}

func (s *mockCalculatorServer) Subtract(ctx context.Context, req *v1.SubtractRequest) (*v1.SubtractResponse, error) {
	resp := &v1.SubtractResponse{}
	resp.SetResult(req.GetA() - req.GetB())
	return resp, nil
}

func setupMockGRPCServer(t *testing.T) (string, func()) {
	t.Helper()
	lis, err := net.Listen("tcp", "localhost:0")
	require.NoError(t, err)

	server := grpc.NewServer()
	v1.RegisterCalculatorServiceServer(server, &mockCalculatorServer{})
	reflection.Register(server)

	go func() {
		if err := server.Serve(lis); err != nil {
			// Don't log error on graceful shutdown
			if err != grpc.ErrServerStopped {
				t.Logf("gRPC server error: %v", err)
			}
		}
	}()

	return lis.Addr().String(), func() {
		server.GracefulStop()
	}
}

func TestParseProtoByReflection_Integration(t *testing.T) {
	addr, cleanup := setupMockGRPCServer(t)
	defer cleanup()

	t.Run("successful reflection", func(t *testing.T) {
		fds, err := ParseProtoByReflection(context.Background(), addr)
		require.NoError(t, err)
		assert.NotNil(t, fds)
		assert.NotEmpty(t, fds.File, "Expected to find at least one file descriptor")

		// Verify that the calculator proto is found
		var foundCalculatorProto bool
		for _, f := range fds.File {
			if f.GetName() == "proto/examples/calculator/v1/calculator.proto" {
				foundCalculatorProto = true
				break
			}
		}
		assert.True(t, foundCalculatorProto, "Calculator service proto should be discovered")
	})

	t.Run("connection failure", func(t *testing.T) {
		_, err := ParseProtoByReflection(context.Background(), "localhost:9999") // Invalid address
		assert.Error(t, err)
	})
}
