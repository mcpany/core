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
	config := &configv1.UpstreamServiceConfig{}
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
		"grpc_service": {"address": "localhost:50051"},
		"upstream_authentication": {
			"mtls": {
				"client_cert_path": "/non/existent/cert",
				"client_key_path": "/non/existent/key",
				"ca_cert_path": "/non/existent/ca"
			}
		}
	}`
	config := &configv1.UpstreamServiceConfig{}
	require.NoError(t, protojson.Unmarshal([]byte(configJSON), config))

	// minSize=1 forces connection creation immediately, triggering factory error
	p, err := NewGrpcPool(1, 5, 100, nil, nil, config, true)
	assert.Error(t, err)
	assert.Nil(t, p)
}

func TestGrpcPool_New_InvalidTarget(t *testing.T) {
	serviceConfig := &configv1.UpstreamServiceConfig{}
	serviceConfig.SetGrpcService(&configv1.GrpcUpstreamService{
		Address: proto.String(""), // Empty address
	})

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
	config := &configv1.UpstreamServiceConfig{}
	require.NoError(t, protojson.Unmarshal([]byte(configJSON), config))

	p, err := NewGrpcPool(1, 1, 100, dialer, nil, config, true)
	require.NoError(t, err)
	defer func() { _ = p.Close() }()

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	_, err = p.Get(ctx)
	require.Error(t, err)
}
