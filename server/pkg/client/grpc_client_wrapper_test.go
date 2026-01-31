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
	"google.golang.org/protobuf/proto"
)

// mockGrpcConn mocks the Conn interface.
type mockGrpcConn struct {
	state      connectivity.State
	closeErr   error
	closeCalled bool
}

func (m *mockGrpcConn) Invoke(ctx context.Context, method string, args interface{}, reply interface{}, opts ...grpc.CallOption) error {
	return nil
}

func (m *mockGrpcConn) NewStream(ctx context.Context, desc *grpc.StreamDesc, method string, opts ...grpc.CallOption) (grpc.ClientStream, error) {
	return nil, nil
}

func (m *mockGrpcConn) Close() error {
	m.closeCalled = true
	return m.closeErr
}

func (m *mockGrpcConn) GetState() connectivity.State {
	return m.state
}

func TestNewGrpcClientWrapper(t *testing.T) {
	conn := &mockGrpcConn{}
	// Empty config results in nil checker
	config := configv1.UpstreamServiceConfig_builder{}.Build()

	wrapper := NewGrpcClientWrapper(conn, config, nil)
	assert.NotNil(t, wrapper)
	assert.Equal(t, conn, wrapper.Conn)
	assert.Equal(t, config, wrapper.config)
	assert.Nil(t, wrapper.checker)

	// Config that produces a checker
	configWithCheck := configv1.UpstreamServiceConfig_builder{
		GrpcService: configv1.GrpcUpstreamService_builder{
			Address: proto.String("localhost:50051"),
			HealthCheck: configv1.GrpcHealthCheck_builder{}.Build(),
		}.Build(),
	}.Build()

	wrapperWithCheck := NewGrpcClientWrapper(conn, configWithCheck, nil)
	assert.NotNil(t, wrapperWithCheck.checker)

	// Test with provided checker
	checker := health.NewChecker()
	wrapper2 := NewGrpcClientWrapper(conn, config, checker)
	assert.NotNil(t, wrapper2)
	assert.Equal(t, checker, wrapper2.checker)
}

func TestIsHealthy(t *testing.T) {
	t.Run("StateShutdown", func(t *testing.T) {
		conn := &mockGrpcConn{state: connectivity.Shutdown}
		config := configv1.UpstreamServiceConfig_builder{}.Build()
		wrapper := NewGrpcClientWrapper(conn, config, nil)

		assert.False(t, wrapper.IsHealthy(context.Background()))
	})

	t.Run("BufnetAddress", func(t *testing.T) {
		conn := &mockGrpcConn{state: connectivity.Ready}
		config := configv1.UpstreamServiceConfig_builder{
			GrpcService: configv1.GrpcUpstreamService_builder{
				Address: proto.String("bufnet"),
			}.Build(),
		}.Build()

		wrapper := NewGrpcClientWrapper(conn, config, nil)

		assert.True(t, wrapper.IsHealthy(context.Background()))
	})

	t.Run("NilChecker", func(t *testing.T) {
		conn := &mockGrpcConn{state: connectivity.Ready}
		config := configv1.UpstreamServiceConfig_builder{}.Build()
		// We need to bypass NewGrpcClientWrapper to set checker to nil,
		// because NewGrpcClientWrapper creates one if nil is passed.
		wrapper := &GrpcClientWrapper{
			Conn:    conn,
			config:  config,
			checker: nil,
		}

		assert.True(t, wrapper.IsHealthy(context.Background()))
	})

	t.Run("CheckerUp", func(t *testing.T) {
		conn := &mockGrpcConn{state: connectivity.Ready}
		config := configv1.UpstreamServiceConfig_builder{}.Build()

		// Create a checker that returns Up
		checker := health.NewChecker(
			health.WithCheck(health.Check{
				Name: "always-up",
				Check: func(ctx context.Context) error {
					return nil
				},
			}),
		)

		wrapper := NewGrpcClientWrapper(conn, config, checker)
		assert.True(t, wrapper.IsHealthy(context.Background()))
	})

	t.Run("CheckerDown", func(t *testing.T) {
		conn := &mockGrpcConn{state: connectivity.Ready}
		config := configv1.UpstreamServiceConfig_builder{}.Build()

		// Create a checker that returns Down
		checker := health.NewChecker(
			health.WithCheck(health.Check{
				Name: "always-down",
				Check: func(ctx context.Context) error {
					return errors.New("something went wrong")
				},
			}),
		)

		wrapper := NewGrpcClientWrapper(conn, config, checker)
		assert.False(t, wrapper.IsHealthy(context.Background()))
	})
}

func TestClose(t *testing.T) {
	conn := &mockGrpcConn{}
	config := configv1.UpstreamServiceConfig_builder{}.Build()
	wrapper := NewGrpcClientWrapper(conn, config, nil)

	err := wrapper.Close()
	assert.NoError(t, err)
	assert.True(t, conn.closeCalled)

	// Test error propagation
	connErr := &mockGrpcConn{closeErr: errors.New("close error")}
	wrapperErr := NewGrpcClientWrapper(connErr, config, nil)
	err = wrapperErr.Close()
	assert.Error(t, err)
	assert.Equal(t, "close error", err.Error())
}
