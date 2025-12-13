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

package protobufparser

import (
	"context"
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	v1 "github.com/mcpany/core/proto/examples/weather/v1"
)

type mockWeatherServer struct {
	v1.UnimplementedWeatherServiceServer
}

func (s *mockWeatherServer) GetWeather(_ context.Context, req *v1.GetWeatherRequest) (*v1.GetWeatherResponse, error) {
	return v1.GetWeatherResponse_builder{Weather: "sunny"}.Build(), nil
}

func setupMockGRPCServer(t *testing.T) (string, func()) {
	t.Helper()
	lis, err := net.Listen("tcp", "localhost:0")
	require.NoError(t, err)

	server := grpc.NewServer()
	v1.RegisterWeatherServiceServer(server, &mockWeatherServer{})
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

		// Verify that the weather proto is found
		var foundWeatherProto bool
		for _, f := range fds.File {
			if f.GetName() == "proto/examples/weather/v1/weather.proto" {
				foundWeatherProto = true
				break
			}
		}
		assert.True(t, foundWeatherProto, "Weather service proto should be discovered")
	})

	t.Run("connection failure", func(t *testing.T) {
		_, err := ParseProtoByReflection(context.Background(), "localhost:9999") // Invalid address
		assert.Error(t, err)
	})
}
