
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
	dialer := func(ctx context.Context, addr string) (net.Conn, error) {
		return lis.Dial()
	}

	configJSON := `{"grpc_service": {"address": "bufnet"}}`
	config := &configv1.UpstreamServiceConfig{}
	require.NoError(t, protojson.Unmarshal([]byte(configJSON), config))

	p, err := NewGrpcPool(1, 5, 100, dialer, nil, config, true)
	require.NoError(t, err)
	assert.NotNil(t, p)
	defer p.Close()

	assert.Equal(t, 1, p.Len())

	client, err := p.Get(context.Background())
	require.NoError(t, err)
	assert.NotNil(t, client)
	assert.True(t, client.IsHealthy(context.Background()))

	p.Put(client)
	assert.Equal(t, 1, p.Len())
}
