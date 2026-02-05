// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package grpc

import (
	"context"
	"net"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

func startMockGrpcServer(t *testing.T) *bufconn.Listener {
	lis := bufconn.Listen(1024 * 1024)
	s := grpc.NewServer()
	go func() {
		if err := s.Serve(lis); err != nil {
			t.Logf("Server exited with error: %v", err)
		}
	}()
	t.Cleanup(func() {
		s.Stop()
	})
	return lis
}

func TestGrpcPool_New(t *testing.T) {
	lis := startMockGrpcServer(t)
	dialer := func(_ context.Context, _ string) (net.Conn, error) {
		return lis.Dial()
	}

	configJSON := `{"grpc_service": {"address": "bufnet"}}`
	config := configv1.UpstreamServiceConfig_builder{}.Build()
	require.NoError(t, protojson.Unmarshal([]byte(configJSON), config))

	p, err := NewGrpcPool(1, 5, 100, dialer, nil, config, true)
	require.NoError(t, err)
	assert.NotNil(t, p)
	defer func() { _ = p.Close() }()

	assert.Equal(t, 1, p.Len())

	client, err := p.Get(context.Background())
	require.NoError(t, err)
	assert.NotNil(t, client)
	assert.True(t, client.IsHealthy(context.Background()))

	p.Put(client)
	assert.Equal(t, 1, p.Len())
}

func TestGrpcPool_New_MtlsFailure(t *testing.T) {
	// mTLS with missing files
	configJSON := `{
		"grpc_service": {"address": "127.0.0.1:50051"},
		"upstream_auth": {
			"mtls": {
				"client_cert_path": "/non/existent/cert",
				"client_key_path": "/non/existent/key",
				"ca_cert_path": "/non/existent/ca"
			}
		}
	}`
	config := configv1.UpstreamServiceConfig_builder{}.Build()
	require.NoError(t, protojson.Unmarshal([]byte(configJSON), config))

	// minSize=1 forces connection creation immediately, triggering factory error
	p, err := NewGrpcPool(1, 5, 100, nil, nil, config, true)
	assert.Error(t, err)
	assert.Nil(t, p)
}

func TestGrpcPool_New_InvalidTarget(t *testing.T) {
	serviceConfig := configv1.UpstreamServiceConfig_builder{
		GrpcService: configv1.GrpcUpstreamService_builder{
			Address: proto.String(""), // Empty address
		}.Build(),
	}.Build()

	// minSize=1 to force factory usage
	_, err := NewGrpcPool(1, 5, 100, nil, nil, serviceConfig, false)
	require.Error(t, err)
}

func TestGrpcPool_New_NilConfig(t *testing.T) {
	_, err := NewGrpcPool(1, 1, 100, nil, nil, nil, false)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "service config is nil")
}

func TestGrpcPool_Get_ContextCancelled(t *testing.T) {
	lis := startMockGrpcServer(t)
	dialer := func(_ context.Context, _ string) (net.Conn, error) {
		return lis.Dial()
	}

	configJSON := `{"grpc_service": {"address": "bufnet"}}`
	config := configv1.UpstreamServiceConfig_builder{}.Build()
	require.NoError(t, protojson.Unmarshal([]byte(configJSON), config))

	p, err := NewGrpcPool(1, 1, 100, dialer, nil, config, true)
	require.NoError(t, err)
	defer func() { _ = p.Close() }()

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	_, err = p.Get(ctx)
	require.Error(t, err)
}

func TestGrpcPool_Close_StopsChecker(t *testing.T) {
	// This test verifies that Close() can be called multiple times without panic
	// and ostensibly cleans up resources.
	// Since we can't easily verify internal state of health.Checker without mocking,
	// we ensure that the new code path (Close -> Stop) is executed safely.

	lis := startMockGrpcServer(t)
	dialer := func(_ context.Context, _ string) (net.Conn, error) {
		return lis.Dial()
	}

	configJSON := `{"grpc_service": {"address": "bufnet"}}`
	config := configv1.UpstreamServiceConfig_builder{}.Build()
	require.NoError(t, protojson.Unmarshal([]byte(configJSON), config))

	p, err := NewGrpcPool(1, 5, 100, dialer, nil, config, true)
	require.NoError(t, err)

	// Verify pool type is what we expect (wrapped) via interface check if exported?
	// poolWithChecker is unexported, so we can't check type easily.
	// But we can check that Close works.
	err = p.Close()
	assert.NoError(t, err)

	// Double close should be safe (pool implementation usually handles this,
	// but health checker Stop is idempotent? The health library says Stop is safe to call multiple times)
	// Even if not, our Close wrapper calls p.Pool.Close() which handles pool state.
	// Let's call it again to be sure no panic.
	err = p.Close()
	assert.NoError(t, err)
}
