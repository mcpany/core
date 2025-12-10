/*
 * Copyright 2024 Author(s) of MCP Any
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

package service

import (
	"context"
	"net"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"

	"github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
)

type mockServiceRegistry struct {
	mu        sync.Mutex
	statusMap map[string]bool
}

func (m *mockServiceRegistry) UpdateHealthCheckStatus(serviceID string, isHealthy bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.statusMap[serviceID] = isHealthy
}

func (m *mockServiceRegistry) isHealthy(serviceID string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.statusMap[serviceID]
}

func TestHTTPChecker(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	registry := &mockServiceRegistry{statusMap: make(map[string]bool)}
	checker := NewHTTPChecker("test-service", &config.HTTPHealthCheck{Address: server.URL}, 100*time.Millisecond, registry)

	go checker.Start()
	time.Sleep(250 * time.Millisecond)
	checker.Stop()

	assert.True(t, registry.isHealthy("test-service"))
}

func TestHTTPChecker_Unhealthy(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer server.Close()

	registry := &mockServiceRegistry{statusMap: make(map[string]bool)}
	checker := NewHTTPChecker("test-service", &config.HTTPHealthCheck{Address: server.URL}, 100*time.Millisecond, registry)

	go checker.Start()
	time.Sleep(250 * time.Millisecond)
	checker.Stop()

	assert.False(t, registry.isHealthy("test-service"))
}

func TestHTTPChecker_NetworkFailure(t *testing.T) {
	registry := &mockServiceRegistry{statusMap: make(map[string]bool)}
	checker := NewHTTPChecker("test-service", &config.HTTPHealthCheck{Address: "http://localhost:12345"}, 100*time.Millisecond, registry)

	go checker.Start()
	time.Sleep(250 * time.Millisecond)
	checker.Stop()

	assert.False(t, registry.isHealthy("test-service"))
}
func TestGRPCChecker(t *testing.T) {
	lis, err := net.Listen("tcp", ":0")
	if err != nil {
		t.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	healthServer := health.NewServer()
	healthpb.RegisterHealthServer(s, healthServer)
	go s.Serve(lis)
	defer s.Stop()

	registry := &mockServiceRegistry{statusMap: make(map[string]bool)}
	checker := NewGRPCChecker("test-service", &config.GRPCHealthCheck{Address: lis.Addr().String()}, 100*time.Millisecond, registry)

	go checker.Start()
	time.Sleep(250 * time.Millisecond)
	checker.Stop()

	assert.True(t, registry.isHealthy("test-service"))
}

func TestNewChecker(t *testing.T) {
	registry := &mockServiceRegistry{statusMap: make(map[string]bool)}

	httpConfig := &config.HealthCheck{
		IntervalSeconds: 1,
		Check: &config.HealthCheck_HttpHealthCheck{
			HttpHealthCheck: &config.HTTPHealthCheck{
				Address: "http://localhost:8080",
			},
		},
	}
	httpChecker, err := NewChecker("test-http", httpConfig, registry)
	assert.NoError(t, err)
	assert.IsType(t, &httpChecker{}, httpChecker)

	grpcConfig := &config.HealthCheck{
		IntervalSeconds: 1,
		Check: &config.HealthCheck_GrpcHealthCheck{
			GrpcHealthCheck: &config.GRPCHealthCheck{
				Address: "localhost:50051",
			},
		},
	}
	grpcChecker, err := NewChecker("test-grpc", grpcConfig, registry)
	assert.NoError(t, err)
	assert.IsType(t, &grpcChecker{}, grpcChecker)
}
type mockHealthServer struct {
	healthpb.UnimplementedHealthServer
	status healthpb.HealthCheckResponse_ServingStatus
}

func TestGRPCChecker_NetworkFailure(t *testing.T) {
	registry := &mockServiceRegistry{statusMap: make(map[string]bool)}
	checker := NewGRPCChecker("test-service", &config.GRPCHealthCheck{Address: "localhost:12345"}, 100*time.Millisecond, registry)

	go checker.Start()
	time.Sleep(250 * time.Millisecond)
	checker.Stop()

	assert.False(t, registry.isHealthy("test-service"))
}

func (s *mockHealthServer) Check(ctx context.Context, in *healthpb.HealthCheckRequest) (*healthpb.HealthCheckResponse, error) {
	return &healthpb.HealthCheckResponse{Status: s.status}, nil
}

func TestGRPCChecker_Unhealthy(t *testing.T) {
	lis, err := net.Listen("tcp", ":0")
	if err != nil {
		t.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	healthServer := &mockHealthServer{status: healthpb.HealthCheckResponse_NOT_SERVING}
	healthpb.RegisterHealthServer(s, healthServer)
	go s.Serve(lis)
	defer s.Stop()

	registry := &mockServiceRegistry{statusMap: make(map[string]bool)}
	checker := NewGRPCChecker("test-service", &config.GRPCHealthCheck{Address: lis.Addr().String()}, 100*time.Millisecond, registry)

	go checker.Start()
	time.Sleep(250 * time.Millisecond)
	checker.Stop()

	assert.False(t, registry.isHealthy("test-service"))
}
