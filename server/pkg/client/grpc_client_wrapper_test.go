// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package client

import (
	"context"
	"errors"
	"testing"

	"github.com/alexliesenfeld/health"
	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
)

// MockConn implements the Conn interface for testing.
type MockConn struct {
	grpc.ClientConnInterface
	State       connectivity.State
	CloseErr    error
	CloseCalled bool
}

func (m *MockConn) GetState() connectivity.State {
	return m.State
}

func (m *MockConn) Close() error {
	m.CloseCalled = true
	return m.CloseErr
}

// MockChecker implements the health.Checker interface for testing.
type MockChecker struct {
	Result health.CheckerResult
}

func (m *MockChecker) Check(ctx context.Context) health.CheckerResult {
	return m.Result
}

func (m *MockChecker) Start() {}
func (m *MockChecker) Stop()  {}
func (m *MockChecker) GetRunningPeriodicCheckCount() int { return 0 }
func (m *MockChecker) IsStarted() bool { return true }

func TestNewGrpcClientWrapper(t *testing.T) {
	tests := []struct {
		name          string
		config        *configv1.UpstreamServiceConfig
		checker       health.Checker
		expectChecker bool
	}{
		{
			name: "with provided checker",
			config: func() *configv1.UpstreamServiceConfig {
                cfg := &configv1.UpstreamServiceConfig{}
                grpcCfg := &configv1.GrpcUpstreamService{}
                grpcCfg.SetAddress("localhost:50051")
                grpcCfg.SetHealthCheck(&configv1.GrpcHealthCheck{})
                cfg.SetGrpcService(grpcCfg)
                return cfg
            }(),
			checker:       &MockChecker{},
			expectChecker: true,
		},
		{
			name: "without provided checker",
			config: func() *configv1.UpstreamServiceConfig {
                cfg := &configv1.UpstreamServiceConfig{}
                grpcCfg := &configv1.GrpcUpstreamService{}
                grpcCfg.SetAddress("localhost:50051")
                grpcCfg.SetHealthCheck(&configv1.GrpcHealthCheck{})
                cfg.SetGrpcService(grpcCfg)
                return cfg
            }(),
			checker:       nil,
			expectChecker: true, // Should create a default checker
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			conn := &MockConn{}
			wrapper := NewGrpcClientWrapper(conn, tt.config, tt.checker)

			assert.NotNil(t, wrapper)
			assert.Equal(t, conn, wrapper.Conn)
			assert.Equal(t, tt.config, wrapper.config)

			if tt.expectChecker {
				assert.NotNil(t, wrapper.checker)

				if tt.checker != nil {
					assert.Equal(t, tt.checker, wrapper.checker)
				}
			} else {
				assert.Nil(t, wrapper.checker)
			}
		})
	}
}

func TestIsHealthy(t *testing.T) {
	tests := []struct {
		name     string
		state    connectivity.State
		address  string
		checker  health.Checker
		expected bool
	}{
		{
			name:     "connection shutdown",
			state:    connectivity.Shutdown,
			address:  "localhost:50051",
			checker:  &MockChecker{Result: health.CheckerResult{Status: health.StatusUp}},
			expected: false,
		},
		{
			name:     "bufnet address",
			state:    connectivity.Ready,
			address:  "bufnet",
			checker:  &MockChecker{Result: health.CheckerResult{Status: health.StatusDown}},
			expected: true,
		},
		{
			name:     "no checker",
			state:    connectivity.Ready,
			address:  "localhost:50051",
			checker:  nil,
			expected: true,
		},
		{
			name:     "checker up",
			state:    connectivity.Ready,
			address:  "localhost:50051",
			checker:  &MockChecker{Result: health.CheckerResult{Status: health.StatusUp}},
			expected: true,
		},
		{
			name:     "checker down",
			state:    connectivity.Ready,
			address:  "localhost:50051",
			checker:  &MockChecker{Result: health.CheckerResult{Status: health.StatusDown}},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			conn := &MockConn{State: tt.state}
			cfg := &configv1.UpstreamServiceConfig{}
			grpcCfg := &configv1.GrpcUpstreamService{}
			grpcCfg.SetAddress(tt.address)
			cfg.SetGrpcService(grpcCfg)

			wrapper := &GrpcClientWrapper{
				Conn:    conn,
				config:  cfg,
				checker: tt.checker,
			}

			result := wrapper.IsHealthy(context.Background())
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestClose(t *testing.T) {
	tests := []struct {
		name     string
		closeErr error
	}{
		{
			name:     "successful close",
			closeErr: nil,
		},
		{
			name:     "error on close",
			closeErr: errors.New("close failed"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			conn := &MockConn{CloseErr: tt.closeErr}
			wrapper := &GrpcClientWrapper{Conn: conn}

			err := wrapper.Close()

			assert.True(t, conn.CloseCalled)
			assert.Equal(t, tt.closeErr, err)
		})
	}
}
