// Copyright (C) 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package health

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/mcpany/core/pkg/upstream"
	"github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
)

// mockHealthServer is a mock implementation of the gRPC health checking server.
type mockHealthServer struct {
	grpc_health_v1.UnimplementedHealthServer
	status grpc_health_v1.HealthCheckResponse_ServingStatus
}

func (s *mockHealthServer) Check(ctx context.Context, in *grpc_health_v1.HealthCheckRequest) (*grpc_health_v1.HealthCheckResponse, error) {
	if in.Service == "not_found" {
		return nil, status.Error(codes.NotFound, "service not found")
	}
	return &grpc_health_v1.HealthCheckResponse{Status: s.status}, nil
}

func (s *mockHealthServer) Watch(in *grpc_health_v1.HealthCheckRequest, stream grpc_health_v1.Health_WatchServer) error {
	return status.Error(codes.Unimplemented, "watching is not supported")
}

func setupGrpcTestServer(status grpc_health_v1.HealthCheckResponse_ServingStatus) (*grpc.Server, *bufconn.Listener) {
	lis := bufconn.Listen(1024 * 1024)
	s := grpc.NewServer()
	grpc_health_v1.RegisterHealthServer(s, &mockHealthServer{status: status})
	go func() {
		if err := s.Serve(lis); err != nil {
			panic(err)
		}
	}()
	return s, lis
}

func TestGrpcChecker_Success(t *testing.T) {
	s, lis := setupGrpcTestServer(grpc_health_v1.HealthCheckResponse_SERVING)
	defer s.Stop()

	mockUpstream := &upstream.MockUpstream{
		MockAddress: lis.Addr().String(),
	}

	cfg := &config.GrpcHealthCheck{
		Service:  "my-service",
		Insecure: true,
		Interval: &durationpb.Duration{Seconds: 1},
	}

	checker, err := NewGrpcChecker("test-grpc-success", cfg, mockUpstream)
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err = checker.HealthCheck(ctx)
	assert.NoError(t, err)
}

func TestGrpcChecker_NotServing(t *testing.T) {
	s, lis := setupGrpcTestServer(grpc_health_v1.HealthCheckResponse_NOT_SERVING)
	defer s.Stop()

	mockUpstream := &upstream.MockUpstream{
		MockAddress: lis.Addr().String(),
	}

	cfg := &config.GrpcHealthCheck{
		Service:  "my-service",
		Insecure: true,
	}

	checker, err := NewGrpcChecker("test-grpc-not-serving", cfg, mockUpstream)
	require.NoError(t, err)

	err = checker.HealthCheck(context.Background())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "is not in serving state")
}

func TestGrpcChecker_ServiceNotFound(t *testing.T) {
	s, lis := setupGrpcTestServer(grpc_health_v1.HealthCheckResponse_SERVING)
	defer s.Stop()

	mockUpstream := &upstream.MockUpstream{
		MockAddress: lis.Addr().String(),
	}

	cfg := &config.GrpcHealthCheck{
		Service:  "not_found", // This service name will trigger a NotFound error from the mock server
		Insecure: true,
	}

	checker, err := NewGrpcChecker("test-grpc-not-found", cfg, mockUpstream)
	require.NoError(t, err)

	err = checker.HealthCheck(context.Background())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "service not found")
}

func TestGrpcChecker_ConnectionError(t *testing.T) {
	// Don't start a server, so the connection will fail.
	lis, err := net.Listen("tcp", "localhost:0")
	require.NoError(t, err)
	addr := lis.Addr().String()
	lis.Close() // Close the listener immediately

	mockUpstream := &upstream.MockUpstream{
		MockAddress: addr,
	}

	cfg := &config.GrpcHealthCheck{
		Service:  "my-service",
		Insecure: true,
	}

	checker, err := NewGrpcChecker("test-grpc-conn-error", cfg, mockUpstream)
	require.NoError(t, err)

	ctx, cancel := context.withTimeout(context.Background(), 1*time.Second)
	defer cancel()

	err = checker.HealthCheck(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to dial gRPC service")
}
